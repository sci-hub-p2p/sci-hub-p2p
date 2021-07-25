// Copyright 2021 Trim21<trim21.me@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

// nolint
package main

import (
	"archive/zip"
	"fmt"
	"path/filepath"

	"github.com/cheggaaa/pb/v3"
	"github.com/ipfs/go-cid"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/pkg/dagserv"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/vars"
)

func main() {
	LoadTestData()
}

func LoadTestData() {
	bar := pb.StartNew(100000 - 4)
	zipFiles, err := filepath.Glob("d:/data/11200*/*.zip")
	if err != nil {
		logger.Fatal(err)
	}

	for i, file := range zipFiles {
		err = func() error {
			db, err := bbolt.Open(fmt.Sprintf("./tmp/test--%d.bolt", i), 0600, &bbolt.Options{
				FreelistType: bbolt.FreelistMapType,
				NoSync:       true,
			})
			if err != nil {
				return err
			}
			err = dagserv.InitDB(db)
			if err != nil {
				logger.Fatal(err)
			}
			r, err := zip.OpenReader(file)
			if err != nil {
				return err
			}
			defer r.Close()
			for _, f := range r.File {
				bar.Increment()
				offset, err := f.DataOffset()
				size := f.CompressedSize64
				r, err := f.Open()
				if err != nil {
					return err
				}
				_, err = dagserv.Add(db, r, file, size, uint64(offset))
				if err != nil {
					_ = r.Close()
					return err
				}
				_ = r.Close()
			}

			db.View(func(tx *bbolt.Tx) error {
				return tx.Bucket(vars.NodeBucketName()).ForEach(func(k, v []byte) error {
					_, err := cid.Cast(k)
					if err != nil {
						logger.Fatal(err, file)
					}
					return nil
				})
			})
			return nil
		}()
		if err != nil {
			logger.Error(err)
		}

	}

}

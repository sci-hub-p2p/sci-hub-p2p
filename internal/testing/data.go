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

package testing

import (
	"archive/zip"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/pkg/dagserv"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/variable"
)

func LoadTestData() {
	const count = 8

	bar := pb.StartNew(100000 - 4)
	zipFiles, err := filepath.Glob("d:/data/11200*/*.zip")
	if err != nil {
		logger.Fatal(err)
	}

	c := make(chan string, count)
	wg := sync.WaitGroup{}
	wg.Add(count)

	for i := 0; i < count; i++ {
		db, err := bbolt.Open(fmt.Sprintf("./test-%d.bolt", i), 0600, &bbolt.Options{
			FreelistType: bbolt.FreelistMapType,
			NoSync:       true,
		})
		if err != nil {
			logger.Fatal(err)
		}
		go func(db *bbolt.DB) {
			for file := range c {
				err = func() error {
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
							return err
						}
						r.Close()
					}
					return nil
				}()
				if err != nil {
					logger.Error(err)
				}
			}
			db.Sync()
			db.Close()
			wg.Done()
		}(db)
	}

	for _, file := range zipFiles {
		c <- file
	}

	for len(c) != 0 {
		time.Sleep(time.Second)
	}

	close(c)

	wg.Wait()
	bar.Finish()

	db, err := bbolt.Open("./test.bolt", 0600, &bbolt.Options{FreelistType: bbolt.FreelistMapType})
	if err != nil {
		logger.Fatal(err)
	}
	for i := 0; i < count; i++ {
		_db, err := bbolt.Open(fmt.Sprintf("./test-%d.bolt", i), 0600, &bbolt.Options{FreelistType: bbolt.FreelistMapType})
		if err != nil {
			logger.Fatal(err)
		}
		err = db.Batch(func(tx *bbolt.Tx) error {
			db := tx.Bucket(variable.NodeBucketName())
			return _db.View(func(_tx *bbolt.Tx) error {
				_b := _tx.Bucket(variable.NodeBucketName())
				return _b.ForEach(func(k, v []byte) error {
					return db.Put(k, v)
				})
			})
		})
		if err != nil {
			logger.Error(err)
		}
		_db.Close()
	}
	db.Close()
}

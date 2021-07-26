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
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"

	"sci_hub_p2p/pkg/dagserv"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/persist"
	"sci_hub_p2p/pkg/variable"
)

func main() {
	LoadTestData()
}

func LoadTestData() {
	const count = 8
	bar := pb.StartNew(100000 - 4)
	zipFiles, err := filepath.Glob("d:/data/11200*/*.zip")
	if err != nil {
		logger.Fatal("", zap.Error(err))
	}

	err = os.Remove("./test.bolt")
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logger.Fatal("", zap.Error(err))
		}
	}

	c := make(chan string, count)
	wg := sync.WaitGroup{}
	wg.Add(count)
	var dbSlice []*bbolt.DB
	for i := 0; i < count; i++ {
		db, err := bbolt.Open(fmt.Sprintf("./tmp/test-%d.bolt", i), 0600, &bbolt.Options{
			FreelistType: bbolt.FreelistMapType,
			NoSync:       true,
		})
		if err != nil {
			logger.Fatal("", zap.Error(err))
		}
		dbSlice = append(dbSlice, db)
		db.Update(func(tx *bbolt.Tx) error {
			err := tx.DeleteBucket(variable.BlockBucketName())
			if err != nil {
				logger.Fatal("", zap.Error(err))
			}
			err = tx.DeleteBucket(variable.NodeBucketName())
			if err != nil {
				logger.Fatal("", zap.Error(err))
			}
			return nil
		})
		err = dagserv.InitDB(db)
		if err != nil {
			logger.Fatal("", zap.Error(err))
		}
		go func(db *bbolt.DB) {
			for file := range c {
				err := dagserv.AddZip(db, file)
				if err != nil {
					logger.Error("", zap.Error(err))
				}
			}

			err := db.Sync()
			if err != nil {
				logger.Error("", zap.Error(err))
			}
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

	db, err := bbolt.Open("./test.bolt", 0600, &bbolt.Options{
		FreelistType: bbolt.FreelistMapType,
		NoSync:       true,
	})
	if err != nil {
		logger.Fatal("", zap.Error(err))
	}
	err = dagserv.InitDB(db)
	if err != nil {
		logger.Fatal("", zap.Error(err))
	}

	for i, srcDB := range dbSlice {
		fmt.Println("copy db", i)
		err = persist.CopyBucket(srcDB, db, variable.NodeBucketName())

		if err != nil {
			logger.Error("", zap.Error(err))
		}

		err = persist.CopyBucket(srcDB, db, variable.BlockBucketName())

		if err != nil {
			logger.Error("", zap.Error(err))
		}

		err = srcDB.Close()
		if err != nil {
			logger.Fatal("", zap.Error(err))
		}

		err := db.Sync()
		if err != nil {
			logger.Fatal("", zap.Error(err))
		}

	}

	err = db.Close()
	if err != nil {
		logger.Fatal("", zap.Error(err))
	}
}

// Copyright 2021 Trim21 <trim21.me@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.

package indexes

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/itchio/lzma"
	"github.com/pkg/errors"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"

	"sci_hub_p2p/internal/utils"
	"sci_hub_p2p/pkg/consts"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/vars"
)

var loadCmd = &cobra.Command{
	Use:           "load",
	Short:         "Load indexes into database.",
	Example:       "indexes load /path/to/*.jsonlines.lzma [--glob '/path/to/data/*.jsonlines.lzma']",
	SilenceErrors: false,
	PreRunE:       utils.EnsureDir(vars.GetAppBaseDir()),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		args, err = utils.MergeGlob(args, glob)
		if err != nil {
			return errors.Wrap(err, "can't load any index files")
		}
		sort.Strings(args)
		fmt.Printf("find %d files to load", len(args))

		db, err := bbolt.Open(vars.IndexesBoltPath(), consts.DefaultFilePerm, &bbolt.Options{
			FreelistType: bbolt.FreelistMapType,
			NoSync:       true,
		})
		if err != nil {
			return errors.Wrap(err, "cant' open database file, maybe another process is running")
		}
		defer func(db *bbolt.DB) {
			if e := db.Close(); e != nil {
				e = errors.Wrap(e, "can't save data to disk")
				if err == nil {
					err = e
				} else {
					logger.Error("", zap.Error(e))
				}
			}
		}(db)

		err = db.Update(func(tx *bbolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists(consts.IndexBucketName())

			return errors.Wrap(err, "failed to create bucket")
		})
		if err != nil {
			return err
		}

		bar := progressbar.Default(int64(len(args)))
		for _, file := range args {
			_ = bar.Add64(1)

			err = db.Batch(func(tx *bbolt.Tx) error {
				err := loadIndexFile(tx.Bucket(consts.IndexBucketName()), file)
				if err != nil {
					return errors.Wrap(err, "can't load indexes file "+file)
				}

				return nil
			})

			if err != nil {
				return errors.Wrap(err, "can't save torrent data to database")
			}
			if err := db.Sync(); err != nil {
				return errors.Wrap(err, "failed to save data to disk")
			}

		}

		return nil
	},
}

var glob string

func init() {
	loadCmd.Flags().StringVar(&glob, "glob", "",
		"glob pattern to search indexes to avoid 'Argument list too long' error")
}

func loadIndexFile(b *bbolt.Bucket, name string) (err error) {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	reader := lzma.NewReader(f)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		var s []string

		err = json.Unmarshal(scanner.Bytes(), &s)
		if err != nil || len(s) != 2 {
			return errors.Wrap(err, "can't parse json "+scanner.Text())
		}

		value, err := base64.StdEncoding.DecodeString(s[1])
		if err != nil {
			return errors.Wrap(err, "can't decode base64")
		}

		key, err := url.QueryUnescape(strings.TrimSuffix(s[0], ".pdf"))
		if err != nil {
			return err
		}

		err = b.Put([]byte(key), value)
		if err != nil {
			return errors.Wrap(err, "can't save record to database")
		}
	}

	err = scanner.Err()
	if err != nil {
		return errors.Wrap(err, "can't scan file")
	}

	return errors.Wrap(scanner.Err(), "can't scan file")
}

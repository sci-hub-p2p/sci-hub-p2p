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

package indexes

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/itchio/lzma"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/internal/utils"
	"sci_hub_p2p/pkg/constants"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/variable"
)

var loadCmd = &cobra.Command{
	Use:           "load",
	Short:         "Load indexes into database.",
	Example:       "indexes load /path/to/*.jsonlines.lzma [--glob '/path/to/data/*.jsonlines.lzma']",
	SilenceErrors: false,
	PreRunE:       utils.EnsureDir(variable.GetAppBaseDir()),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		var s []string
		if glob != "" {
			s, err = utils.GlobWithExpand(glob)
			if err != nil {
				return errors.Wrapf(err, "can't search torrents with glob '%s'", glob)
			}
		}
		db, err := bbolt.Open(variable.GetPaperBoltPath(), constants.DefaultFileMode, bbolt.DefaultOptions)
		if err != nil {
			return errors.Wrap(err, "cant' open database file, maybe another process is running")
		}
		defer func(db *bbolt.DB) {
			if e := db.Close(); e != nil {
				e = errors.Wrap(e, "can't save data to disk")
				if err == nil {
					err = e
				} else {
					logger.Error(e)
				}
			}
		}(db)

		s = utils.Unique(append(args, s...))
		if len(s) == 0 {
			return fmt.Errorf("cant' find any index file to load")
		}
		err = db.Batch(func(tx *bbolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists(constants.PaperBucket())
			if err != nil {
				return errors.Wrap(err, "can't create bucket in database")
			}

			for _, file := range s {
				err := loadIndexFile(b, file)
				if err != nil {
					return errors.Wrap(err, "can't load indexes file "+file)
				}

			}

			return nil
		})
		if err != nil {
			return errors.Wrap(err, "can't save torrent data to database")
		}
		fmt.Printf("successfully load %d torrents into database\n", len(s))

		return nil
	},
}

var glob string

func init() {
	loadCmd.Flags().StringVar(&glob, "glob", "",
		"glob pattern to search indexes to avoid 'Argument list too long' error")
}

func loadIndexFile(b *bbolt.Bucket, name string) error {
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

	return errors.Wrap(scanner.Err(), "can't scan file")
}

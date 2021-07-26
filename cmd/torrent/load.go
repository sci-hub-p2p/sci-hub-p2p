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

package torrent

import (
	"encoding/hex"
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"

	"sci_hub_p2p/internal/torrent"
	"sci_hub_p2p/internal/utils"
	"sci_hub_p2p/pkg/constants"
	"sci_hub_p2p/pkg/constants/size"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/persist"
	"sci_hub_p2p/pkg/variable"
)

var Cmd = &cobra.Command{
	Use:           "torrent",
	Short:         "load torrent into database",
	SilenceErrors: false,
}

var loadCmd = &cobra.Command{
	Use:           "load",
	Short:         "Load torrents into database.",
	Example:       "torrent load /path/to/*.torrents [--glob '/path/to/data/*.torrents']",
	SilenceErrors: false,
	PreRunE:       utils.EnsureDir(variable.GetTorrentStoragePath()),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		args, err = utils.MergeGlob(args, glob)
		if err != nil {
			return errors.Wrap(err, "can't load any torrent files")
		}

		db, err := bbolt.Open(filepath.Join(variable.GetAppBaseDir(), "torrent.bolt"),
			constants.DefaultFilePerm, bbolt.DefaultOptions)

		if err != nil {
			return errors.Wrap(err, "can't open database file, maybe another process is running")
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

		err = db.Batch(func(tx *bbolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists(constants.TorrentBucket())
			if err != nil {
				return errors.Wrap(err, "can't create bucket in database")
			}
			for _, file := range args {
				f, err := torrent.ParseFile(file)
				if err != nil {
					return err
				}
				err = persist.PutTorrent(b, f)
				if err != nil {
					return err
				}
				dst := filepath.Join(variable.GetTorrentStoragePath(), f.InfoHash+".torrent")
				err = utils.Copy(file, dst)
				if err != nil {
					return errors.Wrapf(err, "can't copy torrent file to %s", dst)
				}
			}

			return nil
		})
		if err != nil {
			return errors.Wrap(err, "can't save torrent data to database")
		}
		fmt.Printf("successfully load %d torrents into database\n", len(args))

		return nil
	},
}

var getCmd = &cobra.Command{
	Use:           "get",
	Short:         "get torrent data database.",
	Example:       "torrent get ${InfoHash}",
	Args:          cobra.ExactArgs(1),
	SilenceErrors: false,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if len(args[0]) != size.Sha1Hex {
			return fmt.Errorf("%s is not a valid sha1", args[0])
		}

		var db *bbolt.DB
		db, err = bbolt.Open(filepath.Join(variable.GetAppBaseDir(), "torrent.bolt"),
			constants.DefaultFilePerm, bbolt.DefaultOptions)
		if err != nil {
			return errors.Wrap(err, "cant' open database file, maybe another process is running?")
		}
		defer func(db *bbolt.DB) {
			e := db.Close()
			if e != nil {
				if err == nil {
					err = e
				} else {
					logger.Error("", zap.Error(e))
				}
			}
		}(db)
		p, err := hex.DecodeString(args[0])
		if err != nil {
			return errors.Wrap(err, "info hash is not valid hex string")
		}

		err = db.View(func(tx *bbolt.Tx) error {
			b := tx.Bucket(constants.TorrentBucket())
			if b == nil {
				return fmt.Errorf("can't find data in database")
			}
			t, err := persist.GetTorrent(b, p)
			if err != nil {
				return err
			}

			s, err := t.DumpIndent()
			if err != nil {
				return errors.Wrap(err, "can't dump torrent data into json format")
			}

			fmt.Println(s)

			return nil
		})

		if err != nil {
			return errors.Wrap(err, "can't get torrent from database")
		}

		return nil
	},
}

var glob string

func init() {
	Cmd.AddCommand(loadCmd, getCmd, verifyCmd)

	loadCmd.Flags().StringVar(&glob, "glob", "",
		"glob pattern to search torrents to avoid 'Argument list too long' error")
}

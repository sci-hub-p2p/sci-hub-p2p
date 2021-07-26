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

package ipfs

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"

	"sci_hub_p2p/internal/utils"
	"sci_hub_p2p/pkg/constants"
	"sci_hub_p2p/pkg/dag"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/variable"
)

var addCmd = &cobra.Command{
	Use:     "add",
	Short:   "add all files in a zip files",
	PreRunE: utils.EnsureDir(variable.GetAppBaseDir()),
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		args, err = utils.MergeGlob(args, glob)
		if err != nil {
			return errors.Wrap(err, "no zip files to add")
		}
		db, err := bbolt.Open(filepath.Join(variable.GetAppBaseDir(), constants.IPFSBlockDB),
			constants.DefaultFilePerm, &bbolt.Options{NoSync: true})
		if err != nil {
			return errors.Wrap(err, "failed to open database")
		}
		defer func(db *bbolt.DB) {
			err := db.Close()
			if err != nil {
				logger.Error("failed to close DataBase", zap.Error(err))
			}
		}(db)
		err = dag.InitDB(db)
		if err != nil {
			return errors.Wrap(err, "failed to initialize database")
		}

		width := len(strconv.Itoa(len(args)))

		for i, file := range args {
			logger.Info(fmt.Sprintf("processing file %0*d/%d %s", width, i, len(args), file))
			if err := dag.AddZip(db, file); err != nil {
				logger.Error("failed to add files from zip archive", zap.Error(err))
			}

			if i%10 == 0 {
				err := db.Sync()
				if err != nil {
					logger.Error("failed to sync database to DB", zap.Error(err))
				}
			}
		}

		return errors.Wrap(db.Sync(), "failed to flush data to disk")
	},
}

var glob string

func init() {
	addCmd.Flags().StringVar(&glob, "glob", "", "glob pattern")
}

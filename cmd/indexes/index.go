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
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"sci_hub_p2p/cmd/util"
	"sci_hub_p2p/internal/torrent"
	"sci_hub_p2p/pkg/indexes"
	"sci_hub_p2p/pkg/logger"
)

var IndexCmd = &cobra.Command{
	Use:   "indexes",
	Short: "Generate or load indexes",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("indexes command")

		return cmd.Help()
	},
}

var genCmd = &cobra.Command{
	Use:     "gen",
	Short:   "Generate indexes from a data file.",
	Example: "indexes gen -t /path/to/torrentPath -f /path/to/data",
	RunE: func(cmd *cobra.Command, args []string) error {
		torrentFile, err := os.Open(torrentPath)
		if err != nil {
			return errors.Wrapf(err, "can't open file %s", torrentPath)
		}
		defer torrentFile.Close()

		t, err := torrent.ParseReader(torrentFile)
		if err != nil {
			return err
		}

		s, err := os.Stat(out)
		if err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(out, os.ModeDir)
				if err != nil {
					return errors.Wrapf(err, "Can't create output dir %s", out)
				}
			}
		} else {
			if !s.IsDir() {
				return fmt.Errorf("output path is not a directory")
			}
		}

		err = indexes.Generate(dataDir, out, t)

		if err != nil {
			return errors.Wrap(err, "can't generate indexes from file")
		}

		logger.Debugf("data: %s\n", dataDir)
		logger.Debugf("torrent: %s\n", torrentPath)
		logger.Debugf("out dir: %s\n", out)

		return err
	},
}

var dataDir string
var torrentPath string
var out string

func init() {
	IndexCmd.AddCommand(genCmd)

	genCmd.Flags().StringVarP(&dataDir, "data", "d", "", "Path to data directory")
	genCmd.Flags().StringVarP(&torrentPath, "torrent", "t", "",
		"TorrentPath path of this data file")
	genCmd.Flags().StringVarP(&out, "out", "o", "./out/", "Output directory")

	if err := util.MarkFlagsRequired(genCmd, "data", "torrent"); err != nil {
		log.Fatalln(err)
	}
	if err := genCmd.MarkFlagFilename("torrent", "torrentPath"); err != nil {
		log.Fatalln(err)
	}
}

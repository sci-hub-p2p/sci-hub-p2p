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

package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"sci_hub_p2p/internal/torrent"
	"sci_hub_p2p/pkg/index"
)

var indexCmd = &cobra.Command{
	Use: "index",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("index command")
		return cmd.Help()
	},
}

var genCmd = &cobra.Command{
	Use:     "gen",
	Short:   "generate index from a data file.",
	Example: "index gen -t /path/to/torrentPath -f /path/to/data",
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
		}

		if !s.IsDir() {
			return fmt.Errorf("output path is not a directory")
		}

		index, err := index.FromZip(data)
		if err != nil {
			return err
		}
		index.InfoHash = t.InfoHash

		if debug {
			fmt.Printf("data: %s\n", data)
			fmt.Printf("torrent: %s\n", torrentPath)
			fmt.Printf("out dir: %s\n", out)
		}

		zipPath := filepath.Join(out, index.InfoHash+".zip")
		log.Println("save indexes to local file",zipPath)
		zipFile, err := os.Create(zipPath)
		if err != nil {
			return errors.Wrapf(err, "can't create file %s", zipPath)
		}
		defer zipFile.Close()

		return index.OutToFile(zipFile)
	},
}

var genReadCmd = &cobra.Command{
	Use:     "read",
	Example: "index read ./path/to/index.json.gz",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		r, err := os.Open(args[0])
		if err != nil {
			return errors.Wrapf(err, "can't open file %s", args[0])
		}
		defer r.Close()
		f, err := index.Read(r)
		if err != nil {
			return errors.Wrapf(err, "can't parse file %s", args[0])

		}
		fmt.Println(f)
		return nil
	},
}

var data string
var torrentPath string
var out string

func init() {
	indexCmd.AddCommand(genCmd, genReadCmd)

	genCmd.Flags().StringVarP(&data, "data", "d", "", "path to data file")
	genCmd.Flags().StringVarP(&torrentPath, "torrent", "t", "", "torrentPath path of this data file")
	genCmd.Flags().StringVarP(&out, "out", "o", "./out/", "output directory")

	if err := MarkFlagsRequired(genCmd, "data", "torrent"); err != nil {
		log.Fatalln(err)
	}
	if err := genCmd.MarkFlagFilename("torrent", "torrentPath"); err != nil {
		log.Fatalln(err)
	}
}

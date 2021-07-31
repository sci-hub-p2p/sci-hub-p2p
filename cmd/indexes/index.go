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
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"sci_hub_p2p/cmd/flag"
	"sci_hub_p2p/internal/torrent"
	"sci_hub_p2p/internal/utils"
	"sci_hub_p2p/pkg/consts"
	"sci_hub_p2p/pkg/indexes"
	"sci_hub_p2p/pkg/logger"
)

var Cmd = &cobra.Command{
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
				err = os.MkdirAll(out, os.ModeDir|consts.DefaultFilePerm)
				if err != nil {
					return errors.Wrapf(err, "Can't create output dir %s", out)
				}
			}
		} else {
			if !s.IsDir() {
				return fmt.Errorf("output path is not a directory")
			}
		}

		logger.Info("start generate indexes for torrent", zap.String("torrent", t.Name))
		err = indexes.Generate(dataDir, out, t, flag.DisableProgressBar)

		if err != nil {
			return errors.Wrapf(err, "can't generate indexes from torrent %s", t.Name)
		}

		logger.Debug("data: " + dataDir)
		logger.Debug("torrent: " + torrentPath)
		logger.Debug("out dir: " + out)

		return err
	},
}

var dataDir string
var torrentPath string
var out string

func init() {
	Cmd.AddCommand(genCmd, loadCmd)

	genCmd.Flags().StringVarP(&dataDir, "data", "d", "", "Path to data directory")
	genCmd.Flags().StringVarP(&torrentPath, "torrent", "t", "",
		"TorrentPath path of this data file")
	genCmd.Flags().StringVarP(&out, "out", "o", "./out/", "Output directory")
	genCmd.Flags().BoolVar(
		&flag.DisableProgressBar, "disable-progress", false, "disable progress bar if you don't like it",
	)

	if err := utils.MarkFlagsRequired(genCmd, "data", "torrent"); err != nil {
		log.Fatalln(err)
	}

	if err := genCmd.MarkFlagFilename("torrent", "torrentPath"); err != nil {
		panic(err)
	}
}

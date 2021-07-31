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
package torrent

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"sci_hub_p2p/internal/torrent"
	"sci_hub_p2p/internal/utils"
	"sci_hub_p2p/pkg/hash"
	"sci_hub_p2p/pkg/logger"
)

var verifyCmd = &cobra.Command{
	Use:           "verify",
	Short:         "verify downloaded data of a torrent.",
	Example:       "torrent verify -t /path/123456.torrent -d /path/to/data/123456/",
	SilenceErrors: false,
	Args:          cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		t, err := torrent.ParseFile(torrentPath)
		if err != nil {
			return errors.Wrap(err, "failed to parse torrent")
		}

		ok, err := utils.DirExist(filepath.Join(dataDir, t.Name))
		if err != nil {
			return errors.Wrap(err, "failed to find data directory")
		}
		if ok {
			dataDir = filepath.Join(dataDir, t.Name)
		}

		logger.Info("find torrent data in " + dataDir)

		bar := pb.StartNew(t.PieceCount())
		r := pieceReader{t: t, path: dataDir}
		for i, piece := range t.Pieces {
			bar.Increment()
			p, err := r.readPiece(i)
			if err != nil {
				return errors.Wrap(err, "failed to read piece")
			}
			if !bytes.Equal(hash.Sha1SumBytes(p), piece) {
				logger.Error("piece hash mismatch", zap.Int("index", i))
			}
		}
		bar.Finish()
		fmt.Printf("successfully load %d torrents into database\n", len(args))

		return nil
	},
}

var dataDir string
var torrentPath string

func init() {
	verifyCmd.Flags().StringVarP(&torrentPath, "torrent", "t", "", "torrent path")
	verifyCmd.Flags().StringVarP(&dataDir, "data", "d", "", "path to data directory")

	if err := utils.MarkFlagsRequired(verifyCmd, "torrent", "data"); err != nil {
		panic(err)
	}
}

type pieceReader struct {
	t    *torrent.Torrent
	path string
}

// this can be better, because we won't need to verify a single piece, we just need to verify whole torrent.
func (r pieceReader) readPiece(i int) ([]byte, error) {
	var currentFileStart int64
	var currentFileEnd int64
	var pieceLength = int(r.t.PieceLength)
	var bytesStart = r.t.PieceLength * int64(i)
	var currentFileIndex = 0
	var pieceExpectedLength = pieceLength

	if i == r.t.PieceCount()-1 {
		var count int64
		for _, file := range r.t.Files {
			count += file.Length
		}

		pieceExpectedLength = int(count % int64(pieceLength))
	}
	var needToRead = pieceExpectedLength
	var p = make([]byte, pieceExpectedLength)

	for needToRead > 0 {
		readOffset := +int64(pieceExpectedLength - needToRead)
		file := r.t.Files[currentFileIndex]

		currentFileEnd = currentFileStart + file.Length
		if currentFileEnd < bytesStart {
			currentFileStart += file.Length
			currentFileIndex++

			continue
		}

		err := func() error {
			f, err := os.Open(filepath.Join(r.path, file.Name()))
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err = f.Seek(bytesStart-currentFileStart+readOffset, io.SeekStart); err != nil {
				return err
			}

			size, err := f.Read(p[readOffset:])
			needToRead -= size

			if errors.Is(err, io.EOF) {
				currentFileStart += file.Length
				err = nil
				currentFileIndex++
			}

			return err
		}()

		if err != nil {
			return nil, err
		}
	}

	return p, nil
}

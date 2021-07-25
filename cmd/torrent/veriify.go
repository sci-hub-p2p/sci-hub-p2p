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
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"sci_hub_p2p/internal/torrent"
	"sci_hub_p2p/pkg/hash"
	"sci_hub_p2p/pkg/logger"
)

var verifyCmd = &cobra.Command{
	Use:           "verify",
	Short:         "verify data of a torrent.",
	Example:       "torrent verify /path/to.torrent /path/to/data/dir/",
	SilenceErrors: false,
	Args:          cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		torrentPath := args[0]
		t, err := torrent.ParseFile(torrentPath)
		if err != nil {
			return errors.Wrap(err, "failed to parse torrent")
		}
		dataDir := filepath.Join(args[1], t.Name)
		bar := pb.StartNew(t.PieceCount())
		r := pieceReader{t: t, path: dataDir}
		for i, piece := range t.Pieces {
			bar.Increment()
			p, err := r.readPiece(i)
			if err != nil {
				return errors.Wrap(err, "failed to read piece")
			}
			if !bytes.Equal(hash.Sha1SumBytes(p), piece) {
				logger.Errorf("piece %d mismatch", i)
			}
		}
		bar.Finish()
		fmt.Printf("successfully load %d torrents into database\n", len(args))

		return nil
	},
}

type pieceReader struct {
	t    *torrent.Torrent
	path string
}

// todo: this can be better, because we won't need to verify a single piece, we just need to verify whole torrent.
func (r pieceReader) readPiece(i int) ([]byte, error) {
	var (
		currentFileStart    int64
		currentFileEnd      int64
		pieceLength         = int(r.t.PieceLength)
		bytesStart          = r.t.PieceLength * int64(i)
		currentFileIndex    = 0
		pieceExpectedLength = pieceLength
	)
	if i == r.t.PieceCount()-1 {
		var count int64 = 0
		for _, file := range r.t.Files {
			count += file.Length
		}
		pieceExpectedLength = int(count % int64(pieceLength))
	}
	var needToRead = pieceExpectedLength

	var p = make([]byte, pieceExpectedLength)

	for needToRead > 0 {
		var readOffset = +int64(pieceExpectedLength - needToRead)

		file := r.t.Files[currentFileIndex]
		currentFileEnd = currentFileStart + file.Length
		if currentFileEnd < bytesStart {
			currentFileIndex++
			currentFileStart += file.Length

			continue
		}
		err := func() error {
			f, err := os.Open(filepath.Join(r.path, file.Name()))
			if err != nil {
				return err
			}
			defer f.Close()
			_, err = f.Seek(bytesStart-currentFileStart+readOffset, io.SeekStart)
			if err != nil {
				return err
			}
			size, err := f.Read(p[readOffset:])
			needToRead -= size
			if errors.Is(err, io.EOF) {
				currentFileIndex++
				currentFileStart += file.Length
				err = nil
			}
			if err != nil {
				return err
			}

			return nil
		}()

		if err != nil {
			return nil, err
		}
	}

	return p, nil
}

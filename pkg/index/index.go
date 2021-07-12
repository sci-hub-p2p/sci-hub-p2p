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

// Package index zip file index
package index

import (
	"bytes"
	"errors"
	"io"

	"sci_hub_p2p/internal/torrent"
	"sci_hub_p2p/internal/zip"
)

type Index struct {
	Name       string
	Doi        string
	DataOffset int64
	// compressed data length
	CompressedSize   int64
	CompressedMethod uint16
	Crc32            uint32
	Torrent          torrent.Torrent
}

var ErrCheckSum = errors.New("checksum mismatch")

// DecompressFromPiece combine all wanted pieces first
func (i Index) DecompressFromPiece(pieces []byte) ([]byte, error) {
	offset := int(i.DataOffset % int64(i.Torrent.PieceLength))
	compressed := pieces[offset : int64(offset)+i.CompressedSize]
	return i.Decompress(compressed)
}

// Decompress raw bytes data.
func (i Index) Decompress(data []byte) ([]byte, error) {
	var decompressed, err = zip.TryDecompressor(bytes.NewReader(data), i.CompressedMethod)
	if err != nil {
		return nil, err
	}
	content, err := io.ReadAll(decompressed)
	if err != nil {
		return nil, err
	}
	if !zip.CheckSum(content, i.Crc32) {
		return nil, ErrCheckSum
	}
	return content, nil
}

// WantedPieces starts from 0
func (i Index) WantedPieces() []int {
	t := i.Torrent

	// 应该不可能会溢出吧
	start := int(i.DataOffset / int64(t.PieceLength))
	end := int((i.DataOffset + i.CompressedSize) / int64(t.PieceLength))

	return makeRange(start, end)
}

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}

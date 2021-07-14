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

// Package indexes zip file indexes
package indexes

import (
	"archive/zip"
	"encoding/binary"

	"github.com/pkg/errors"

	"sci_hub_p2p/internal/torrent"
)

type IndexInDB struct {
	InfoHash         [20]byte // This can be empty when indexing data from same torrent.
	PieceStart       uint32   // 4 bytes
	DataOffset       uint32   // should be uint32 I think
	CompressedMethod uint16   // 2 bytes
	CompressedSize   uint64   // 8 bytes
	Sha256           [32]byte // For IPFS, not vary necessarily
}

func (i IndexInDB) Dump() []byte {
	// TODO: use binary.Write
	var p = make([]byte, 20+4+4+2+8+32)
	copy(p, i.InfoHash[:])
	binary.LittleEndian.PutUint32(p[20:], i.PieceStart)
	binary.LittleEndian.PutUint32(p[20+4:], i.DataOffset)
	binary.LittleEndian.PutUint16(p[20+4+4:], i.CompressedMethod)
	binary.LittleEndian.PutUint64(p[20+4+4+2:], i.CompressedSize)
	copy(p[20+4+4+2+8:], i.Sha256[:])

	return p
}

func (i *IndexInDB) Load(p []byte) {
	// TODO: use binary.Read
	copy(i.InfoHash[:], p)
	i.PieceStart = binary.LittleEndian.Uint32(p[20:])
	i.DataOffset = binary.LittleEndian.Uint32(p[20+4:])
	i.CompressedMethod = binary.LittleEndian.Uint16(p[20+4+4:])
	i.CompressedSize = binary.LittleEndian.Uint64(p[20+4+4+2:])
	copy(i.Sha256[:], p[20+4+4+2+8:])
}

type Index struct {
	Doi        string
	Name       string
	DataOffset int64
	// compressed data length
	CompressedSize   int64
	CompressedMethod uint16
	Crc32            uint32
	Torrent          torrent.Torrent
}

var ErrCheckSum = errors.New("checksum mismatch")

// DecompressFromPiece combine all wanted pieces first.
func (i Index) DecompressFromPiece(pieces []byte) ([]byte, error) {
	offset := int(i.DataOffset % int64(i.Torrent.PieceLength))
	compressed := pieces[offset : int64(offset)+i.CompressedSize]

	return i.Decompress(compressed)
}

// Decompress raw bytes data.
func (i Index) Decompress(data []byte) ([]byte, error) {
	switch i.CompressedMethod {
	case zip.Store:
		// storage
	case zip.Deflate:
		// should decompress with deflate
	}

	return nil, nil
}

// WantedPieces starts from 0.
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

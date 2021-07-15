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

	"sci_hub_p2p/internal/torrent"
)

type PerFile struct {
	Doi string `json:"doi"`

	FileName        string `json:"file_name"` // duplicated with doi maybe
	CompressMethod  uint16 `json:"method"`    // seems that almost all files are just store in zip without compress.
	CompressedSize  int64  `json:"size"`
	Sha256          string `json:"sha256"`
	OffsetFromZip   int64  `json:"offset_from_zip"`
	OffsetFromPiece uint32 `json:"offset_from_piece"`
	Pieces          []int  `json:"pieces"`
	PieceLength     int    `json:"piece_length"`
	Torrent         torrent.Torrent
	File            torrent.File
}

func (f PerFile) String() string {
	return fmt.Sprintf("PerFile{name: %s, method: %d, size: %d, OffsetFromZip: %d}",
		f.FileName, f.CompressMethod, f.CompressedSize, f.OffsetFromZip)
}

//
// // DecompressFromPiece combine all wanted pieces first.
// func (i PerFile) DecompressFromPiece(pieces []byte) ([]byte, error) {
// 	offset := int(i.DataOffset % int64(i.Torrent.PieceLength))
// 	compressed := pieces[offset : int64(offset)+i.CompressedSize]
//
// 	return i.Decompress(compressed)
// }

// // Decompress raw bytes data.
// func (i PerFile) Decompress(data []byte) ([]byte, error) {
// 	switch i.CompressedMethod {
// 	case zip.Store:
// 		// storage
// 	case zip.Deflate:
// 		// should decompress with deflate
// 	}
//
// 	return nil, nil
// }
//
// // WantedPieces starts from 0.
// func (i PerFile) WantedPieces() []int {
// 	t := i.Torrent
//
// 	// 应该不可能会溢出吧
// 	start := int(i.DataOffset / int64(t.PieceLength))
// 	end := int((i.DataOffset + i.CompressedSize) / int64(t.PieceLength))
//
// 	return makeRange(start, end)
// }

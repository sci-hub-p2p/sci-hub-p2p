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

	FileName        string // duplicated with doi maybe
	CompressMethod  uint16 // seems that almost all files are just store in zip without compress.
	CompressedSize  int64
	MultiHash       string
	OffsetFromZip   int64
	OffsetFromPiece int64
	FileIndex       int
	Pieces          []int
	PieceStart      int
	PieceEnd        int
	PieceLength     int64
	Torrent         torrent.Torrent
	File            torrent.File
}

func (f PerFile) String() string {
	return fmt.Sprintf("PerFile{name: %s, method: %d, size: %d, OffsetFromZip: %d, pieceStart: %d, pieceEnd: %d}",
		f.FileName, f.CompressMethod, f.CompressedSize, f.OffsetFromZip, f.PieceStart, f.PieceEnd)
}

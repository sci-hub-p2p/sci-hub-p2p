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

	"github.com/ipfs/go-cid"

	"sci_hub_p2p/internal/torrent"
)

type PerFile struct {
	FileName        string
	CID             cid.Cid
	Doi             string
	Pieces          []int
	File            torrent.File
	Torrent         torrent.Torrent
	PieceStart      int
	OffsetFromZip   int64
	CompressedSize  int64
	OffsetFromPiece int64
	PieceLength     int64
	PieceEnd        int
	FileIndex       int
	CompressMethod  uint16
}

func (f PerFile) String() string {
	return fmt.Sprintf("PerFile{name: %s, method: %d, size: %d, OffsetFromZip: %d, pieceStart: %d, pieceEnd: %d}",
		f.FileName, f.CompressMethod, f.CompressedSize, f.OffsetFromZip, f.PieceStart, f.PieceEnd)
}

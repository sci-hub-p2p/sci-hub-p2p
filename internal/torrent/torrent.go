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

// Package torrent parse raw torrent file and generate info hash.
package torrent

import (
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"
)

const SizeOfSha1 = 20

type Node struct {
	Host string `tuple:"0"`
	Port int    `tuple:"1"`
}

type File struct {
	Length int64
	Path   []string
}

type Torrent struct {
	Announce     string
	Files        []File
	Name         string
	PieceLength  int
	AnnounceList [][]string
	CreationDate int
	Nodes        []Node
	InfoHash     string
	// avoid change, only return copy
	pieces [][]byte
}

var ErrWrongPieces = errors.New("The length of the pieces can't be divided by 20")

func (t *Torrent) SetPieces(s string) error {
	p := []byte(s)
	if len(p)%SizeOfSha1 != 0 {
		return ErrWrongPieces
	}
	t.pieces = make([][]byte, len(p)/SizeOfSha1)
	for i := 0; i < len(p)/SizeOfSha1; i++ {
		t.pieces[i] = p[i*SizeOfSha1 : (i+1)*SizeOfSha1]
	}

	return nil
}

func (t Torrent) PieceCount() int {
	return len(t.pieces) / SizeOfSha1
}

func (t Torrent) Hex(i int) string {
	return hex.EncodeToString(t.pieces[i])
}

func (t Torrent) Piece(i int) []byte {
	var s = make([]byte, SizeOfSha1)
	copy(s, t.pieces[i])

	return s
}

func (t Torrent) SetFiles(files []file) {
	t.Files = make([]File, len(files))
	for i, f := range files {
		t.Files[i] = File{
			Length: f.Length,
			Path:   f.GetPath(),
		}
	}
}

func (t Torrent) String() string {
	return fmt.Sprintf("Torrent{Name=%s, info_hash=%s}", t.Name, t.InfoHash)
}

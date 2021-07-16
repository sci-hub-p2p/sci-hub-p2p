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
	"math/bits"

	"github.com/pkg/errors"

	"sci_hub_p2p/internal/convert"
)

const (
	MaxInt int64 = (1<<bits.UintSize)/2 - 1
)

type file struct {
	Length   int64    `json:"length" bencode:"length"`
	Path     []string `json:"path" bencode:"path"`
	PathUTF8 []string `bencode:"path.utf-8,omitempty"`
}

func (f file) GetPath() []string {
	if f.PathUTF8 != nil {
		return f.PathUTF8
	}

	return f.Path
}

type info struct {
	Files       []file `json:"files" bencode:"files"`
	Name        string `json:"name" bencode:"name"`
	NameUTF8    string `bencode:"name.utf-8,omitempty"`
	PieceLength int64  `json:"piece length" bencode:"piece length"`
	Pieces      string `json:"pieces" bencode:"pieces"`
}

func (i info) GetName() string {
	if i.NameUTF8 != "" {
		return i.NameUTF8
	}

	return i.Name
}

type torrentFile struct {
	Announce     string
	AnnounceList [][]string
	CreationDate int
	Info         info
	// should be a [][string, int], golang didn't support this
	Nodes [][]interface{}
}

func (t torrentFile) toTorrent() (*Torrent, error) {
	var torrent Torrent
	torrent.Name = t.Info.GetName()
	torrent.PieceLength = t.Info.PieceLength
	err := torrent.SetPieces(t.Info.Pieces)
	if err != nil {
		return nil, err
	}
	torrent.Announce = t.Announce
	torrent.AnnounceList = t.AnnounceList
	torrent.SetFiles(t.Info.Files)

	n, err := castNodes(t.Nodes)
	if err != nil {
		return nil, err
	}

	torrent.Nodes = n

	return &torrent, nil
}

func castNodes(i [][]interface{}) ([]Node, error) {
	nodes := make([]Node, len(i))

	for index, item := range i {
		var (
			n   Node
			err = convert.ScanSlice(item, &n)
		)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to convert from %s", item)
		}

		nodes[index] = n
	}

	return nodes, nil
}

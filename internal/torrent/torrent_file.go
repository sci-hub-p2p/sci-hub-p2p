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

	"sci_hub_p2p/internal/convert"
)

const (
	MaxInt int64 = (1<<bits.UintSize)/2 - 1
)

type file struct {
	Length int64    `json:"length" bencode:"length"`
	Path   []string `json:"path" bencode:"path"`
}

type info struct {
	Files       []file `json:"files" bencode:"files"`
	Name        string `json:"name" bencode:"name"`
	PieceLength int    `json:"piece length" bencode:"piece length"`
	Pieces      string `json:"pieces" bencode:"pieces"`
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
	torrent.Name = t.Info.Name
	torrent.Files = t.Info.Files
	torrent.PieceLength = t.Info.PieceLength
	torrent.Pieces = []byte(t.Info.Pieces)
	torrent.Announce = t.Announce
	torrent.AnnounceList = t.AnnounceList

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
		var n Node
		err := convert.ScanSlice(item, &n)
		if err != nil {
			return nil, err
		}
		nodes[index] = n
	}
	return nodes, nil
}

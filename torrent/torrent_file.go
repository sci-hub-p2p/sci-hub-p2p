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
	"errors"
	"fmt"
	"math/bits"
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

	n, err := caseNodes(t.Nodes)
	if err != nil {
		return nil, err
	}

	torrent.Nodes = n

	return &torrent, nil
}

var ErrorType = errors.New("can't not cast type")

func caseNodes(i [][]interface{}) ([]Node, error) {
	nodes := make([]Node, len(i))

	for index, item := range i {
		if len(item) != 2 {
			return nil, fmt.Errorf("%w node %d has wrong value %v", ErrorType, index, item)
		}

		host, ok1 := item[0].(string)
		if !ok1 {
			return nil, fmt.Errorf("%w can't cast node[%d][0] to string, got %v",
				ErrorType, index, item[0])
		}

		port, ok2 := item[1].(int64)
		if !ok2 || port > MaxInt {
			return nil, fmt.Errorf("%w can't cast node[%d][1] to int, got %v",
				ErrorType, index, item[1])
		}

		nodes[index] = Node{host, int(port)}
	}

	return nodes, nil
}

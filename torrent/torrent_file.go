package torrent

import (
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
	Announce     string     `json:"announce"`
	AnnounceList [][]string `json:"announce-list"`
	CreationDate int        `json:"creation date"`
	Info         info       `json:"info"`
	// should be a [][string, int], golang didn't support this
	Nodes [][]interface{} `json:"nodes"`
}

func (t torrentFile) toTorrent() (*Torrent, error) {
	var torrent Torrent
	torrent.Name = t.Info.Name
	torrent.Files = t.Info.Files
	torrent.PieceLength = t.Info.PieceLength
	torrent.Pieces = t.Info.Pieces
	torrent.Announce = t.Announce
	torrent.AnnounceList = t.AnnounceList
	n, err := caseNodes(t.Nodes)
	if err != nil {
		return nil, err
	}

	torrent.Nodes = n

	return &torrent, nil
}

func caseNodes(i [][]interface{}) ([]Node, error) {
	var nodes = make([]Node, len(i))

	for index, item := range i {
		if len(item) != 2 {
			return nil, fmt.Errorf("node %d has worng value %v", index, item)
		}

		var host, ok1 = item[0].(string)
		if !ok1 {
			return nil, fmt.Errorf("can't cast node[%d][0] to string, got %v", index, item[0])
		}

		var port, ok2 = item[1].(int64)
		if !ok2 || port > MaxInt {
			return nil, fmt.Errorf("can't cast node[%d][1] to int, got %v", index, item[1])
		}

		nodes = append(nodes, Node{host, int(port)})

	}
	return nodes, nil
}

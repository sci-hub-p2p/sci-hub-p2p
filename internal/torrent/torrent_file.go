// Copyright 2021 Trim21 <trim21.me@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
package torrent

type file struct {
	Path     []string `json:"path" bencode:"path"`
	PathUTF8 []string `bencode:"path.utf-8,omitempty"`
	Length   int64    `json:"length" bencode:"length"`
}

func (f file) GetPath() []string {
	if f.PathUTF8 != nil {
		return f.PathUTF8
	}

	return f.Path
}

type info struct {
	Name        string `json:"name" bencode:"name"`
	NameUTF8    string `bencode:"name.utf-8,omitempty"`
	Pieces      string `json:"pieces" bencode:"pieces"`
	Files       []file `json:"files" bencode:"files"`
	PieceLength int64  `json:"piece length" bencode:"piece length"`
}

func (i info) GetName() string {
	if i.NameUTF8 != "" {
		return i.NameUTF8
	}

	return i.Name
}

type torrentFile struct {
	AnnounceList [][]string
	// should be a [][string, int], golang didn't support this
	Nodes        [][]interface{}
	Announce     string
	Info         info
	CreationDate int
}

func (t torrentFile) toTorrent() (*Torrent, error) {
	var torrent Torrent
	torrent.Name = t.Info.GetName()
	torrent.PieceLength = t.Info.PieceLength

	err := torrent.setPieces(t.Info.Pieces)
	if err != nil {
		return nil, err
	}

	torrent.Announce = t.Announce
	torrent.AnnounceList = t.AnnounceList
	torrent.setFiles(t.Info.Files)

	return &torrent, nil
}

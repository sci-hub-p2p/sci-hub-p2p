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
	"encoding/json"
	"fmt"
	"path"

	"github.com/pkg/errors"

	"sci_hub_p2p/pkg/consts/size"
)

type Node struct {
	Host string `tuple:"0"`
	Port int    `tuple:"1"`
}

type File struct {
	Path   []string
	Length int64
}

func (f File) Name() string {
	return path.Join(f.Path...)
}

func (f File) Copy() File {
	var n = f
	copy(n.Path, f.Path)

	return n
}

type Torrent struct {
	InfoHash     string
	Announce     string
	Name         string
	infoHash     []byte
	raw          []byte
	Pieces       [][]byte
	Files        []File
	AnnounceList [][]string
	Nodes        []Node
	PieceLength  int64
	CreationDate int
}

var ErrWrongPieces = errors.New("The length of the pieces can't be divided by 20")

func (t *Torrent) RawInfoHash() []byte {
	return t.infoHash
}

func (t *Torrent) Raw() []byte {
	return t.raw
}

func (t *Torrent) setInfoHash(p []byte) {
	t.infoHash = make([]byte, size.Sha1Bytes)
	copy(t.infoHash, p)
	t.InfoHash = hex.EncodeToString(p)
}

func (t *Torrent) setPieces(s string) error {
	sizeOfSha1 := size.Sha1Bytes
	p := []byte(s)
	if len(p)%sizeOfSha1 != 0 {
		return ErrWrongPieces
	}
	t.Pieces = make([][]byte, len(p)/sizeOfSha1)
	for i := 0; i < len(p)/sizeOfSha1; i++ {
		t.Pieces[i] = p[i*sizeOfSha1 : (i+1)*sizeOfSha1]
	}

	return nil
}

func (t Torrent) PieceCount() int {
	return len(t.Pieces)
}

func (t Torrent) Hex(i int) string {
	return hex.EncodeToString(t.Pieces[i])
}

func (t Torrent) Piece(i int) []byte {
	var s = make([]byte, size.Sha1Bytes)
	copy(s, t.Pieces[i])

	return s
}

func (t *Torrent) setFiles(files []file) {
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

func (t *Torrent) Copy() Torrent {
	n := Torrent{
		Announce:     t.Announce,
		Name:         t.Name,
		PieceLength:  t.PieceLength,
		CreationDate: t.CreationDate,
		InfoHash:     t.InfoHash,
		infoHash:     t.infoHash,
	}

	copy(n.Pieces, t.Pieces)
	copy(n.Nodes, t.Nodes)
	copy(n.Files, t.Files)

	return n
}

func (t *Torrent) Dump() ([]byte, error) {
	v, err := json.Marshal(t)
	if err != nil {
		return nil, errors.Wrap(err, "can't encode torrent to JSON format")
	}

	return v, nil
}

func (t *Torrent) DumpIndent() (string, error) {
	var m = make(map[string]interface{})
	m["Files"] = t.Files
	m["InfoHash"] = t.InfoHash
	m["Name"] = t.Name
	m["PieceLength"] = t.PieceLength
	m["..."] = "..."
	v, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", errors.Wrap(err, "can't encode torrent to JSON format")
	}

	return string(v), nil
}

func Load(p []byte) (*Torrent, error) {
	var t = &Torrent{}
	if err := json.Unmarshal(p, t); err != nil {
		return nil, errors.Wrap(err, "can't decode torrent from JSON format")
	}
	v, err := hex.DecodeString(t.InfoHash)
	if err != nil {
		return nil, errors.Wrap(err, "can't decode InfoHash, maybe data broken, you need to reload this torrent")
	}

	t.setInfoHash(v)

	return t, nil
}

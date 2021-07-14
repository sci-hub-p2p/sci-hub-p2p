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

// Package indexes zip file indexes
package indexes

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

const (
	offsetFileName = "indexes.json"
	sha1FileName   = "hash.sha1"
	sha256FileName = "hash.sha256"
)

// File describe the struct of indexes file.
type File struct {
	InfoHash        string   `json:"info_hash"`
	FileNames       []string `json:"file_names"`
	Offset          []int64  `json:"offset"`
	Methods         []uint16 `json:"methods"`
	Crc32           []uint32 `json:"crc32"`
	CompressedSizes []uint64 `json:"compressed_sizes"`
	// store hash in binary
	// hex are too big
	Sha1   []byte `json:"-"`
	Sha256 []byte `json:"-"`
}

func (f File) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("File{InfoHash=%s, files = [\n", f.InfoHash))
	for _, file := range f.Files() {
		buffer.WriteString("  ")
		buffer.WriteString(file.String())
		buffer.WriteString(",\n")
	}
	buffer.WriteString("]")

	return buffer.String()
}

func (f File) Files() []perFile {
	var s = make([]perFile, len(f.FileNames))

	for i := 0; i < len(f.FileNames); i++ {
		s[i] = perFile{
			FileName:       f.FileNames[i],
			Method:         f.Methods[i],
			Offset:         f.Offset[i],
			CompressedSize: f.CompressedSizes[i],
			Sha1:           hex.EncodeToString(f.Sha1[i*20 : i*20+20]),   // sha1 has 20 length
			Sha256:         hex.EncodeToString(f.Sha256[i*32 : i*32+32]), // sha256 has 32 length
		}
	}

	return s
}

var ErrTorrentDataBroken = errors.New("torrent data is broken")

const filesPerTorrent = 100000 // 100k pdf per file

func NewWithPre(n int) File {
	var f = File{
		FileNames:       make([]string, 0, n),
		Offset:          make([]int64, 0, n),
		Methods:         make([]uint16, 0, n),
		Crc32:           make([]uint32, 0, n),
		CompressedSizes: make([]uint64, 0, n),
	}

	return f
}
func (f File) writeJSON(w *zip.Writer) error {
	j, err := w.Create(offsetFileName)
	if err != nil {
		return errors.Wrap(err, "can't write zip file")
	}

	err = json.NewEncoder(j).Encode(f)
	if err != nil {
		return errors.Wrap(err, "can't serialize to json")
	}

	return nil
}

func (f File) writeHASH(w *zip.Writer) error {
	/* Hash is hard to compress. Only about 2% size reduce when using zip.Deflate
	so just store them
	*/
	s1, err := w.CreateHeader(&zip.FileHeader{
		Name:   sha1FileName,
		Method: zip.Store,
	})
	if err != nil {
		return errors.Wrap(err, "can't write zip file")
	}

	if _, err := s1.Write(f.Sha1); err != nil {
		return errors.Wrap(err, "can't write to zip file")
	}

	s2, err := w.CreateHeader(&zip.FileHeader{
		Name:   sha256FileName,
		Method: zip.Store,
	})
	if err != nil {
		return errors.Wrap(err, "can't write to zip file")
	}

	_, err = s2.Write(f.Sha256)

	return errors.Wrap(err, "can't write to zip file")
}

func (f File) OutToFile(w io.Writer) error {
	zipW := zip.NewWriter(w)
	defer zipW.Close()

	if err := f.writeJSON(zipW); err != nil {
		return errors.Wrap(err, "can't write offset to zip file")
	}

	if err := f.writeHASH(zipW); err != nil {
		return errors.Wrap(err, "can't write hash to zip file")
	}

	return nil
}

func Read(r io.Reader) (*File, error) {
	var f File

	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, errors.Wrap(err, "can't decompress with gzip")
	}
	defer gr.Close()

	err = json.NewDecoder(gr).Decode(&f)
	if err != nil {
		return nil, errors.Wrap(err, "can't parse json ")
	}

	return &f, nil
}

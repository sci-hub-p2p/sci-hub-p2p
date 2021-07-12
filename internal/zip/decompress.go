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

// Package zip generate a index from downloaded zip file and decompress from
// fetched data.
package zip

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
)

func Init() {
	raw, err := os.ReadFile("test.zip")
	if err != nil {
		log.Fatal(err)
	}

	buf := bytes.NewReader(raw)

	read, err := OpenReader("test.zip")
	if err != nil {
		msg := "Failed to open: %s"
		log.Fatalf(msg, err)
	}
	defer read.Close()

	for _, file := range read.File {
		dataOffset, err := file.DataOffset()
		if err != nil {
			log.Println(err)

			continue
		}

		if _, err = buf.Seek(dataOffset, io.SeekStart); err != nil {
			log.Println(err)

			continue
		}

		var compressed = make([]byte, file.CompressedSize64)

		if length, err := io.ReadFull(buf, compressed); err != nil ||
			uint64(length) != file.CompressedSize64 {
			log.Println(err)

			continue
		}

		r, err := TryDecompressor(
			bytes.NewReader(compressed),
			// dataOffset,
			// int64(file.CompressedSize64),
			file.Method,
		)
		if err != nil {
			log.Println(err)

			continue
		}

		decompressed, err := io.ReadAll(r)
		if err != nil {
			log.Println(err)

			continue
		}

		fmt.Println(string(decompressed))
	}
}

func CheckSum(b []byte, crc uint32) bool {
	if crc != 0 {
		hash := crc32.ChecksumIEEE(b)
		if hash != crc {
			log.Println("hash check failed", "expected:", crc, "actual:", hash)

			return false
		}
	}

	return true
}

func TryDecompressor(r io.Reader, method uint16) (io.ReadCloser, error) {
	dcomp := decompressor(method)
	if dcomp == nil {
		return nil, ErrAlgorithm
	}

	rc := dcomp(r)

	return rc, nil
}

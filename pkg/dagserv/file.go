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

package dagserv

import (
	"bytes"
	"io/fs"
	"os"
	"time"

	"sci_hub_p2p/pkg/constants"
)

type CompressedFile struct {
	reader             *bytes.Reader
	info               CompressedFileInfo
	zipPath            string
	compressedFilePath string
	size               int64
}

func (c CompressedFile) Read(p []byte) (n int, err error) {
	return c.reader.Read(p)
}

func (c CompressedFile) Close() error {
	c.reader = nil
	return nil
}

func (c CompressedFile) Size() (int64, error) {
	return c.size, nil
}

func (c CompressedFile) AbsPath() string {
	return c.zipPath
}

func (c CompressedFile) Stat() os.FileInfo {
	return CompressedFileInfo{c.zipPath, c.compressedFilePath, c.size}
}

type CompressedFileInfo struct {
	zipPath            string
	compressedFilePath string
	size               int64
}

func (c CompressedFileInfo) Name() string {
	return c.compressedFilePath
}

func (c CompressedFileInfo) Size() int64 {
	return c.size
}

func (c CompressedFileInfo) Mode() fs.FileMode {
	return constants.DefaultFilePerm
}

func (c CompressedFileInfo) ModTime() time.Time {
	return time.Now()
}

func (c CompressedFileInfo) IsDir() bool {
	return false
}

func (c CompressedFileInfo) Sys() interface{} {
	return nil
}

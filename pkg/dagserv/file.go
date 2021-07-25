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
	"io"
	"io/fs"
	"os"
	"time"

	files "github.com/ipfs/go-ipfs-files"
	"github.com/pkg/errors"

	"sci_hub_p2p/pkg/constants"
)

var _ interface {
	files.FileInfo
	io.Reader
} = (*CompressedFile)(nil)

func wrapZipFile(r io.Reader, abs string, size uint64) *CompressedFile {
	return &CompressedFile{
		r:    r,
		abs:  abs,
		size: size,
	}
}

type CompressedFile struct {
	r    io.Reader
	abs  string
	size uint64
}

func (c CompressedFile) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	if err == io.EOF {
		return n, io.EOF
	}

	return n, errors.Wrapf(err, "can't read from r %s", c.abs)
}

func (c CompressedFile) Close() error {
	return nil
}

func (c CompressedFile) Size() (int64, error) {
	return int64(c.size), nil
}

func (c CompressedFile) AbsPath() string {
	return c.abs
}

func (c CompressedFile) Stat() os.FileInfo {
	return CompressedFileInfo{c.abs, int64(c.size)}
}

type CompressedFileInfo struct {
	zipPath string
	size    int64
}

func (c CompressedFileInfo) Name() string {
	return c.zipPath
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

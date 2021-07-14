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

package indexes

import "fmt"

type perFile struct {
	FileName       string `json:"file_name"`
	Offset         int64  `json:"offset"`
	Method         uint16 `json:"method"`
	Crc32          uint32 `json:"crc32"`
	CompressedSize uint64 `json:"size"`
	Sha1           string `json:"sha1"`
	Sha256         string `json:"sha256"`
}

func (f perFile) String() string {
	return fmt.Sprintf("perFile{name: %s, method: %d, offset: %d, size: %d}",
		f.FileName, f.Method, f.Offset, f.CompressedSize)
}

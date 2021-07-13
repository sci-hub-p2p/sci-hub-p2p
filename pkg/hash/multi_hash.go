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

package hash

import (
	"crypto/sha1"
	"crypto/sha256"
	"io"

	"github.com/pkg/errors"
)

const defaultBlockSize = 64 * 1024 // read 64k once

func Sha1Sha256SumReader(f io.Reader) ([]byte, []byte, error) {
	h1 := sha1.New()
	h2 := sha256.New()
	buf := make([]byte, defaultBlockSize)
	for {
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			return nil, nil, errors.Wrap(err, "can't hash, read error")
		}

		h1.Write(buf[:n])
		h2.Write(buf[:n])

		if err == io.EOF {
			break
		}
	}

	return h1.Sum(nil), h2.Sum(nil), nil
}

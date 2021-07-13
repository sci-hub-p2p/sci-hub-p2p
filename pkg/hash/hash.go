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
	"encoding/hex"
	"io"

	"github.com/pkg/errors"
)

func Sha1SumReader(r io.Reader) (string, error) {
	h := sha1.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", errors.Wrap(err, "can't hash content")
	}
	sum := h.Sum(nil)

	return hex.EncodeToString(sum), nil
}

func Sha256SumReader(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", errors.Wrap(err, "can't hash content")
	}
	sum := h.Sum(nil)

	return hex.EncodeToString(sum), nil
}

func Sha1Sum(b []byte) string {
	h := sha1.New()
	_, _ = h.Write(b)
	sum := h.Sum(nil)

	return hex.EncodeToString(sum)
}

func Sha256Sum(b []byte) string {
	h := sha256.New()
	_, _ = h.Write(b)
	sum := h.Sum(nil)

	return hex.EncodeToString(sum)
}

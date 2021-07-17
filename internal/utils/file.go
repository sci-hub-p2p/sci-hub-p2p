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

package utils

import (
	"errors"
	"io"
	"os"
)

// Copy the src file to dst. Any existing file will be overwritten and will not
// copy file attributes.
func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		_ = out.Close()

		return err
	}

	return out.Close()
}

var ErrNotAFile = errors.New("not a file")
var ErrNotADir = errors.New("not a dir")

func FileExist(name string) (bool, error) {
	s, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false, nil
	}
	if s.IsDir() {
		return false, ErrNotAFile
	}

	return true, err
}

func DirExist(name string) (bool, error) {
	s, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false, nil
	}
	if !s.IsDir() {
		return false, ErrNotADir
	}

	return true, err
}

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

package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertSlice(t *testing.T) {
	t.Parallel()

	type T1 struct {
		Host  string   `tuple:"0"`
		Port  int      `tuple:"1"`
		Names []string `tuple:"3"`
		Raw   []byte   `tuple:"2"`
	}

	var (
		v   T1
		row = []interface{}{"233", 66, []byte("string"), []string{"1", "2", "3"}}
		err = ScanSlice(row, &v)
	)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, row[0], v.Host)
	assert.Equal(t, row[1], v.Port)
	assert.Equal(t, row[3], v.Names)
	assert.Equal(t, row[2], v.Raw)
}

func TestDontPanic(t *testing.T) {
	t.Parallel()

	type T1 struct {
		Host  string   `tuple:"0"`
		Port  int      `tuple:"1"`
		Names []string `tuple:"3"`
		Raw   []byte   `tuple:"2"`
	}

	var (
		row = []interface{}{"1"}
		v   T1
	)

	err := ScanSlice(row, &v)
	assert.NotNil(t, err)
}

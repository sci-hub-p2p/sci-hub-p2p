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

package dagserv_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/pkg/constants"
	"sci_hub_p2p/pkg/dagserv"
)

func TestZipArchive(t *testing.T) {
	raw, err := os.ReadFile("./../../testdata/big_file.bin")
	t.Parallel()
	assert.Nil(t, err)

	db, err := bbolt.Open(filepath.Join(t.TempDir(), "test.bolt"), constants.DefaultFilePerm, bbolt.DefaultOptions)
	assert.Nil(t, err)
	defer db.Close()

	_, err = dagserv.Add(db, bytes.NewReader(raw), "../../testdata/big_file.bin", int64(len(raw)), 0)
	assert.Nil(t, err)
}

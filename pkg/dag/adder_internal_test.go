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

package dag

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	ipld "github.com/ipfs/go-ipld-format"
	ufsio "github.com/ipfs/go-unixfs/io"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/pkg/consts"
)

func Test_AddSingleFile(t *testing.T) {
	raw, err := os.ReadFile("./../../testdata/big_file.bin")
	t.Parallel()
	var (
		start  = 642
		length = 256*1024*3 + 8
	)
	assert.Nil(t, err)

	db, err := bbolt.Open(filepath.Join(t.TempDir(), "test.bolt"), consts.DefaultFilePerm, bbolt.DefaultOptions)
	assert.Nil(t, err)
	defer db.Close()
	assert.Nil(t, InitDB(db))
	var n ipld.Node
	assert.Nil(t, db.Batch(func(tx *bbolt.Tx) error {
		n, err = addSingleFile(tx, "../../testdata/big_file.bin",
			io.LimitReader(bytes.NewReader(raw[start:]), int64(length)), int64(start), uint64(length))
		return err
	}))
	dag := New(db)

	r, err := ufsio.NewDagReader(context.TODO(), n, dag)
	assert.Nil(t, err)

	read, err := io.ReadAll(r)

	assert.Equal(t, length, len(read), "size should match")
	assert.Equal(t, start, bytes.Index(raw, read), "should start from index")
}

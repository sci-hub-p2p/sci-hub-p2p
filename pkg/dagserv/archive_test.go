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
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	posinfo "github.com/ipfs/go-ipfs-posinfo"
	"github.com/ipfs/go-merkledag"
	"github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"

	"sci_hub_p2p/pkg/constants"
	"sci_hub_p2p/pkg/dagserv"
	"sci_hub_p2p/pkg/variable"
)

func TestZipArchive(t *testing.T) {
	raw, err := os.ReadFile("./../../testdata/big_file.bin")
	t.Parallel()
	assert.Nil(t, err)

	db, err := bbolt.Open(filepath.Join(t.TempDir(), "test.bolt"), constants.DefaultFilePerm, bbolt.DefaultOptions)
	assert.Nil(t, err)
	defer db.Close()

	n, err := dagserv.Add(db, bytes.NewReader(raw), "../../testdata/big_file.bin", uint64(len(raw)), 0)
	assert.Nil(t, err)
	fmt.Println(n.Cid())
}

func TestDagServ_Add(t *testing.T) {
	t.Parallel()
	var length = 60
	var baseOffset = 18
	var blockOffset = 32
	var tmpDir = t.TempDir()
	var binary = filepath.Join(tmpDir, "test.bin")
	var dbPath = filepath.Join(tmpDir, "test.bolt")

	var raw = make([]byte, 3*256*1024) // 16*256K
	_, _ = rand.Read(raw)

	db, err := bbolt.Open(dbPath, constants.DefaultFilePerm, bbolt.DefaultOptions)
	assert.Nil(t, err)
	defer db.Close()

	dag := dagserv.New(db, 18)

	c, err := (cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   multihash.Names["blake2b-256"],
		MhLength: -1,
	}).Sum(raw)
	assert.Nil(t, err)

	block, _ := blocks.NewBlockWithCid(raw[baseOffset+blockOffset:baseOffset+blockOffset+length], c)

	assert.Nil(t,
		dag.Add(context.TODO(), &posinfo.FilestoreNode{
			Node: &merkledag.RawNode{Block: block},
			PosInfo: &posinfo.PosInfo{
				Offset:   32,
				FullPath: binary,
				Stat:     nil,
			}}),
	)

	assert.Nil(t,
		db.View(func(tx *bbolt.Tx) error {
			b := tx.Bucket(variable.NodeBucketName())
			assert.NotNil(t, b)
			data := b.Get(c.Bytes())
			assert.NotNil(t, data)
			var v = &dagserv.Record{}
			assert.Nil(t, proto.Unmarshal(data, v))
			assert.Equal(t, binary, v.Filename, "filename should be equal")
			assert.Equal(t, uint64(baseOffset+blockOffset), v.Offset, "offset should be equal")
			assert.Equal(t, uint64(length), v.Length, "offset should be equal")
			return nil
		}),
	)
	// test read
	raw[baseOffset+blockOffset+5] = raw[baseOffset+blockOffset+5] + 1
	assert.Nil(t, os.WriteFile(binary, raw, 0600))
	node, err := dag.Get(context.TODO(), c)
	assert.Nil(t, err)

	assert.True(t, bytes.Equal(node.RawData(), raw[baseOffset+blockOffset:baseOffset+blockOffset+length]),
		"file content should match", len(node.RawData()), length)
}

func TestDagServ_Add_Get(t *testing.T) {
	raw, err := os.ReadFile("./../../testdata/big_file.bin")
	t.Parallel()
	assert.Nil(t, err)

	db, err := bbolt.Open(filepath.Join(t.TempDir(), "test.bolt"), constants.DefaultFilePerm, bbolt.DefaultOptions)
	assert.Nil(t, err)
	defer db.Close()

	n, err := dagserv.Add(db, bytes.NewReader(raw), "../../testdata/big_file.bin", uint64(len(raw)), 0)
	assert.Nil(t, err)
	dag := dagserv.New(db, 0)
	m, err := dag.Get(context.TODO(), n.Cid())
	assert.Nil(t, err)
	assert.True(t, bytes.Equal(n.RawData(), m.RawData()))
}

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
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"sync"

	"github.com/ipfs/go-cid"
	chunker "github.com/ipfs/go-ipfs-chunker"
	format "github.com/ipfs/go-ipld-format"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-unixfs/importer/helpers"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

func NewZip() ZipArchive {
	return ZipArchive{
		M: make(map[string]ipld.Node),
		m: &sync.Mutex{},
	}
}

type ZipArchive struct {
	M  map[string]format.Node
	m  *sync.Mutex
	db *leveldb.DB
}

func (d ZipArchive) Get(ctx context.Context, cid cid.Cid) (format.Node, error) {
	d.m.Lock()
	defer d.m.Unlock()
	i, ok := d.M[cid.String()]
	if !ok {
		return nil, ErrNotFound
	}

	return i, nil
}

func (d ZipArchive) GetMany(ctx context.Context, cids []cid.Cid) <-chan *format.NodeOption {
	var c = make(chan *format.NodeOption)
	go func() {
		for _, cid := range cids {
			i, err := d.Get(ctx, cid)
			c <- &format.NodeOption{Node: i, Err: err}
		}
	}()

	return c
}

var dump = false

func (d ZipArchive) Add(ctx context.Context, node format.Node) error {
	d.m.Lock()
	defer d.m.Unlock()
	stat, _ := node.Stat()
	size, _ := node.Size()
	fmt.Println(node.Cid().String(), stat, size, len(node.RawData()))
	if !dump {

		fmt.Println(hex.Dump(node.RawData()[len(node.RawData())-128:]))
		fmt.Println()
		dump = true
	}
	// if
	// fmt.Println(hex.Dump(node.RawData()))
	d.M[node.Cid().String()] = node

	return nil
}

func (d ZipArchive) AddMany(ctx context.Context, nodes []format.Node) error {
	for _, node := range nodes {
		_ = d.Add(ctx, node)
	}

	return nil
}

func (d ZipArchive) Remove(ctx context.Context, cid cid.Cid) error {
	d.m.Lock()
	defer d.m.Unlock()
	delete(d.M, cid.String())

	return nil
}

func (d ZipArchive) RemoveMany(ctx context.Context, cids []cid.Cid) error {
	for _, c := range cids {
		_ = d.Remove(ctx, c)
	}

	return nil
}

func Build(r io.Reader) (ipld.Node, error) {
	prefix := cid.Prefix{
		Version:  0,
		Codec:    cid.DagProtobuf,
		MhType:   multihash.SHA2_256,
		MhLength: -1,
	}
	db, err := leveldb.OpenFile("./database/", nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	dbp := helpers.DagBuilderParams{
		Dagserv: ZipArchive{
			M:  make(map[string]ipld.Node),
			m:  &sync.Mutex{},
			db: db,
		},
		Maxlinks:   helpers.DefaultLinksPerBlock,
		CidBuilder: &prefix,
	}

	chunk, err := chunker.FromString(r, "default")
	if err != nil {
		return nil, errors.Wrapf(err, "can't create default chunker")
	}
	dbh, err := dbp.New(chunk)
	if err != nil {
		return nil, errors.Wrap(err, "can't create dag builder from chunker")
	}
	fmt.Println("start layout")
	n, err := BalanceLayout(dbh)
	if err != nil {
		return nil, errors.Wrapf(err, "can't layout all chunk")
	}

	return n, nil
}

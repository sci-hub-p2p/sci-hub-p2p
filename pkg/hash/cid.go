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
	"context"
	"io"
	"sync"

	"github.com/ipfs/go-cid"
	chunker "github.com/ipfs/go-ipfs-chunker"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-unixfs/importer/balanced"
	"github.com/ipfs/go-unixfs/importer/helpers"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
)

func Sha256CidBalanced(r io.Reader) ([]byte, error) {
	var c, err = Cid(r)
	if err != nil {
		return nil, errors.Wrap(err, "can't generate cid")
	}

	return c.Hash(), nil
}

func Cid(r io.Reader) (cid.Cid, error) {
	var n, err = addFile(r, &AddParams{
		Layout:    "balanced",
		Chunker:   "default",
		RawLeaves: false,
		NoCopy:    false,
		HashFun:   "sha2-256",
		Version:   0,
	})
	if err != nil {
		return cid.Cid{}, errors.Wrap(err, "can't generate cid")
	}

	return n.Cid(), nil
}

type DumpDagServ struct {
	M map[string]ipld.Node
	m *sync.Mutex
}

var errNotFound = errors.New("not found")

func (d DumpDagServ) Get(ctx context.Context, cid cid.Cid) (ipld.Node, error) {
	d.m.Lock()
	defer d.m.Unlock()
	i, ok := d.M[cid.String()]
	if !ok {
		return nil, errNotFound
	}

	return i, nil
}

func (d DumpDagServ) GetMany(ctx context.Context, cids []cid.Cid) <-chan *ipld.NodeOption {
	var c = make(chan *ipld.NodeOption)
	go func() {
		for _, cid := range cids {
			i, err := d.Get(ctx, cid)
			c <- &ipld.NodeOption{Node: i, Err: err}
		}
	}()

	return c
}

func (d DumpDagServ) Add(ctx context.Context, node ipld.Node) error {
	d.m.Lock()
	defer d.m.Unlock()
	d.M[node.Cid().String()] = node

	return nil
}

func (d DumpDagServ) AddMany(ctx context.Context, nodes []ipld.Node) error {
	for _, node := range nodes {
		err := d.Add(ctx, node)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d DumpDagServ) Remove(ctx context.Context, cid cid.Cid) error {
	d.m.Lock()
	defer d.m.Unlock()
	delete(d.M, cid.String())

	return nil
}

func (d DumpDagServ) RemoveMany(ctx context.Context, cids []cid.Cid) error {
	for _, c := range cids {
		_ = d.Remove(ctx, c)
	}

	return nil
}

// AddParams contains all of the configurable parameters needed to specify the
// importing process of a file.
type AddParams struct {
	Layout    string
	Chunker   string
	RawLeaves bool
	NoCopy    bool
	HashFun   string
	Version   int
}

func addFile(r io.Reader, params *AddParams) (ipld.Node, error) {
	if params == nil {
		params = &AddParams{}
	}

	prefix := cid.Prefix{
		Version:  0,
		Codec:    cid.DagProtobuf,
		MhType:   multihash.SHA2_256,
		MhLength: -1,
	}

	dbp := helpers.DagBuilderParams{
		Dagserv: DumpDagServ{
			M: make(map[string]ipld.Node),
			m: &sync.Mutex{},
		},
		RawLeaves:  params.RawLeaves,
		Maxlinks:   helpers.DefaultLinksPerBlock,
		NoCopy:     params.NoCopy,
		CidBuilder: &prefix,
	}

	chunk, err := chunker.FromString(r, params.Chunker)
	if err != nil {
		return nil, errors.Wrapf(err, "can't create chunker %s", params.Chunker)
	}
	dbh, err := dbp.New(chunk)
	if err != nil {
		return nil, errors.Wrap(err, "can't create dagbuilder from chunker")
	}

	n, err := balanced.Layout(dbh)
	if err != nil {
		return nil, errors.Wrapf(err, "can't layout all chunk")
	}

	return n, nil
}

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
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/ipfs/go-cid"
	chunker "github.com/ipfs/go-ipfs-chunker"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-unixfs/importer/balanced"
	"github.com/ipfs/go-unixfs/importer/helpers"
	"github.com/ipfs/go-unixfs/importer/trickle"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
)

func Sha256CidBalanced(r io.Reader) ([]byte, error) {
	var n, err = addFile(r, &AddParams{
		Layout:    "balanced",
		Chunker:   "default",
		RawLeaves: false,
		NoCopy:    false,
		HashFun:   "sha2-256",
		Version:   0,
	})
	if err != nil {
		return nil, err
	}

	return n.Cid().Hash(), nil
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
		d.Remove(ctx, c)
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
	if params.HashFun == "" {
		params.HashFun = "sha2-256"
	}

	prefix := cid.Prefix{
		Version:  0,
		Codec:    18,
		MhType:   multihash.SHA2_256,
		MhLength: -1,
	}

	hashFunCode, ok := multihash.Names[strings.ToLower(params.HashFun)]
	if !ok {
		return nil, fmt.Errorf("unrecognized hash function: %s", params.HashFun)
	}
	prefix.MhType = hashFunCode
	prefix.MhLength = -1
	prefix.Codec = cid.DagCBOR

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

	chnk, err := chunker.FromString(r, params.Chunker)
	if err != nil {
		return nil, err
	}
	dbh, err := dbp.New(chnk)
	if err != nil {
		return nil, err
	}

	var n ipld.Node
	switch params.Layout {
	case "trickle":
		n, err = trickle.Layout(dbh)
	case "balanced", "":
		n, err = balanced.Layout(dbh)
	default:
		return nil, errors.New("invalid Layout")
	}
	return n, err
}

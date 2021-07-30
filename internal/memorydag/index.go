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

package memorydag

import (
	"context"
	"sync"

	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
)

type DumpDagServ struct {
	M map[string]ipld.Node
	sync.Mutex
}

func (d *DumpDagServ) Get(ctx context.Context, cid cid.Cid) (ipld.Node, error) {
	d.Lock()
	defer d.Unlock()

	i, ok := d.M[cid.String()]
	if !ok {
		return nil, ipld.ErrNotFound
	}

	return i, nil
}

func (d *DumpDagServ) GetMany(ctx context.Context, cids []cid.Cid) <-chan *ipld.NodeOption {
	var c = make(chan *ipld.NodeOption)

	go func() {
		for _, cid := range cids {
			i, err := d.Get(ctx, cid)
			c <- &ipld.NodeOption{Node: i, Err: err}
		}
	}()

	return c
}

func (d *DumpDagServ) Add(ctx context.Context, node ipld.Node) error {
	d.Lock()
	defer d.Unlock()
	d.M[node.Cid().String()] = node

	return nil
}

func (d *DumpDagServ) AddMany(ctx context.Context, nodes []ipld.Node) error {
	d.Lock()
	defer d.Unlock()

	for _, node := range nodes {
		d.M[node.Cid().String()] = node
	}

	return nil
}

func (d *DumpDagServ) Remove(ctx context.Context, cid cid.Cid) error {
	d.Lock()
	defer d.Unlock()
	delete(d.M, cid.String())

	return nil
}

func (d *DumpDagServ) RemoveMany(ctx context.Context, cids []cid.Cid) error {
	d.Lock()
	defer d.Unlock()

	for _, c := range cids {
		delete(d.M, c.String())
	}

	return nil
}

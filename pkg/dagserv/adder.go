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

// this damon is for adding files to database.
// avoid sync on every change.

import (
	"context"
	"sync"

	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/vars"
)

var _ ipld.DAGService = (*Adder)(nil)

func NewAdder(tx *bbolt.Tx, baseOffset int64) *Adder {
	return &Adder{
		tx:         tx,
		baseOffset: baseOffset,
	}
}

type Adder struct {
	tx         *bbolt.Tx
	baseOffset int64
	sync.RWMutex
}

func (a *Adder) Commit() error {
	return errors.Wrap(a.tx.Commit(), "failed to commit change in database")
}

func (a *Adder) Add(ctx context.Context, node ipld.Node) error {
	a.Lock()
	defer a.Unlock()

	return errors.Wrap(add(a.tx, node, a.baseOffset), "can't save node to database")
}

func (a *Adder) Get(ctx context.Context, c cid.Cid) (ipld.Node, error) {
	logger.Info("Get", c)
	a.RLock()
	defer a.RUnlock()
	if c.Version() == 0 {
		return nil, ErrNotFound
	}
	switch c.Type() {
	case cid.DagProtobuf:
		b := a.tx.Bucket(vars.NodeBucketName())
		n, err := ReadProtoNode(b, c)

		return n, errors.Wrap(err, "can't read node from database")
	case cid.Raw:
		b := a.tx.Bucket(vars.NodeBucketName())
		n, err := ReadFileStoreNode(b, c)

		return n, errors.Wrap(err, "can't read node from database")
	}

	panic("un-supported cid data type")
}

func (a *Adder) AddMany(ctx context.Context, nodes []ipld.Node) error {
	for _, node := range nodes {
		err := add(a.tx, node, a.baseOffset)
		if err != nil {
			return errors.Wrap(err, "can't save node to database")
		}
	}

	return nil
}

// GetMany TODO: need to parallel this, but I'm lazy.
func (a *Adder) GetMany(ctx context.Context, cids []cid.Cid) <-chan *ipld.NodeOption {
	panic("get many")
}

func (a *Adder) Remove(ctx context.Context, c cid.Cid) error {
	panic("can't Remove node from 'Adder'")
}

func (a *Adder) RemoveMany(ctx context.Context, cids []cid.Cid) error {
	panic("can't RemoveMany node from 'adder'")
}

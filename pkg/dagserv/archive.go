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
	"fmt"
	"reflect"
	"sync"

	"github.com/ipfs/go-cid"
	posinfo "github.com/ipfs/go-ipfs-posinfo"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/variable"
)

func NewZip() ZipArchive {
	return ZipArchive{
		m: &sync.Mutex{},
	}
}

var _ ipld.DAGService = ZipArchive{}

type ZipArchive struct {
	m          *sync.Mutex
	db         *bbolt.DB
	baseOffset uint64
}

func (d ZipArchive) Get(ctx context.Context, c cid.Cid) (ipld.Node, error) {
	d.m.Lock()
	defer d.m.Unlock()
	if c.Version() == 0 {
		return nil, ErrNotFound
	}
	var n ipld.Node
	switch c.Type() {
	case cid.DagProtobuf:
		err := d.db.View(func(tx *bbolt.Tx) error {
			var err error
			b := tx.Bucket(variable.NodeBucketName())
			n, err = ReadProtoNode(b, c)

			return err
		})

		return n, errors.Wrap(err, "can't read node from database")
	case cid.Raw:
		err := d.db.View(func(tx *bbolt.Tx) error {
			var err error
			b := tx.Bucket(variable.NodeBucketName())
			n, err = ReadFileStoreNode(b, c)

			return err
		})

		return n, errors.Wrap(err, "can't read node from database")
	}

	panic("un-supported cid data type")
}

func (d ZipArchive) GetMany(ctx context.Context, cids []cid.Cid) <-chan *ipld.NodeOption {
	var c = make(chan *ipld.NodeOption)
	go func() {
		for _, cid := range cids {
			i, err := d.Get(ctx, cid)
			c <- &ipld.NodeOption{Node: i, Err: err}
		}
	}()

	return c
}

func (d ZipArchive) Add(ctx context.Context, node ipld.Node) error {
	d.m.Lock()
	defer d.m.Unlock()
	if v, ok := node.(*merkledag.ProtoNode); ok {
		err := d.db.Update(func(tx *bbolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists(variable.NodeBucketName())
			if err != nil {
				return errors.Wrap(err, "can't create bucket")
			}

			return SaveProtoNode(b, node.Cid(), v)
		})

		return errors.Wrap(err, "can't save node to database")
	}

	if v, ok := node.(*posinfo.FilestoreNode); ok {
		length, _ := v.Size()
		blockOffsetOfZip := v.PosInfo.Offset + d.baseOffset

		err := d.db.Update(func(tx *bbolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists(variable.NodeBucketName())
			if err != nil {
				return errors.Wrap(err, "can't create bucket")
			}

			return SaveFileStoreMeta(b, node.Cid(), v.PosInfo.FullPath, blockOffsetOfZip, length)
		})

		return errors.Wrap(err, "can't save node to database")
	}

	msg := fmt.Sprintf("unknown ipld.Node type %s", reflect.TypeOf(node))
	logger.Errorf(msg)
	panic(msg)
}

func (d ZipArchive) AddMany(ctx context.Context, nodes []ipld.Node) error {
	for _, node := range nodes {
		err := d.Add(ctx, node)
		if err != nil {
			return errors.Wrap(err, "can't save many nodes to database")
		}
	}

	return nil
}

func (d ZipArchive) Remove(ctx context.Context, cid cid.Cid) error {
	panic("can't remove node from disk")
	// d.m.Lock()
	// defer d.m.Unlock()
	// delete(d.m, cid.String())
	//
	// return nil
}

func (d ZipArchive) RemoveMany(ctx context.Context, cids []cid.Cid) error {
	panic("can't remove node from disk")
	// for _, c := range cids {
	// 	_ = d.Remove(ctx, c)
	// }

	// return nil
}

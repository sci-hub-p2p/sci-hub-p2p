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
	"context"
	"sync"

	"github.com/ipfs/go-cid"
	posinfo "github.com/ipfs/go-ipfs-posinfo"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"

	"sci_hub_p2p/pkg/consts"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/storage"
)

var _ ipld.DAGService = (*Archive)(nil)

func New(db *bbolt.DB) *Archive {
	return &Archive{
		db:  db,
		log: logger.WithLogger("DAGService").Named("Archive"),
	}
}

func InitDB(db *bbolt.DB) error {
	return errors.Wrap(db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(consts.NodeBucketName())
		if err != nil {
			return errors.Wrap(err, "can't create node bucket")
		}
		_, err = tx.CreateBucketIfNotExists(consts.BlockBucketName())
		if err != nil {
			return errors.Wrap(err, "can't create block bucket")
		}

		return nil
	}), "failed to init bolt database")
}

type Archive struct {
	db  *bbolt.DB
	log *zap.Logger
	sync.RWMutex
}

func (d *Archive) Get(ctx context.Context, c cid.Cid) (ipld.Node, error) {
	d.log.Debug("Get Node", zap.String("CID", c.String()))
	d.RLock()
	defer d.RUnlock()
	if c.Version() == 0 {
		return nil, ipld.ErrNotFound
	}
	var n ipld.Node
	switch c.Type() {
	case cid.DagProtobuf:
		err := d.db.View(func(tx *bbolt.Tx) error {
			var err error
			b := tx.Bucket(consts.NodeBucketName())
			n, err = storage.ReadProtoNode(b, c)

			return errors.Wrap(err, "failed to read Protobuf Node from storage")
		})

		return n, err
	case cid.Raw:
		err := d.db.View(func(tx *bbolt.Tx) error {
			var err error
			b := tx.Bucket(consts.NodeBucketName())
			n, err = storage.ReadFileStoreNode(b, c)

			return errors.Wrap(err, "failed to read FileStore Node from storage")
		})

		return n, err
	}

	panic("un-supported cid data type")
}

// GetMany TODO: need to parallel this, but I'm lazy.
func (d *Archive) GetMany(ctx context.Context, cids []cid.Cid) <-chan *ipld.NodeOption {
	var c = make(chan *ipld.NodeOption)
	go func() {
		for _, cid := range cids {
			i, err := d.Get(ctx, cid)
			c <- &ipld.NodeOption{Node: i, Err: err}
		}
	}()

	return c
}

func (d *Archive) Add(_ context.Context, node ipld.Node) error {
	panic("should not add node to Archive dag")
}

func (d *Archive) AddMany(_ context.Context, nodes []ipld.Node) error {
	panic("should not add node to Archive dag")
}

func (d *Archive) Remove(_ context.Context, c cid.Cid) error {
	err := d.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(consts.NodeBucketName())
		if b == nil {
			return nil
		}

		return b.Delete(c.Bytes())
	})

	return errors.Wrap(err, "can't delete node from database")
}

func (d *Archive) RemoveMany(_ context.Context, cids []cid.Cid) error {
	err := d.db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket(consts.NodeBucketName())
		if b == nil {
			return nil
		}
		for _, c := range cids {
			if err := b.Delete(c.Bytes()); err != nil {
				return err
			}
		}

		return nil
	})

	return errors.Wrap(err, "can't delete node from database")
}

func add(tx *bbolt.Tx, node ipld.Node, baseOffset int64) error {
	switch n := node.(type) {
	case *merkledag.ProtoNode:
		return errors.Wrap(storage.SaveProtoNode(tx, node.Cid(), n), "can't save node to database")
	case *posinfo.FilestoreNode:
		length, _ := n.Size()
		blockOffsetOfZip := baseOffset + int64(n.PosInfo.Offset)

		return errors.Wrap(storage.SaveFileStoreMeta(tx, node.Cid(),
			n.PosInfo.FullPath, blockOffsetOfZip, int64(length)),
			"can't save node to database")
	}

	return storage.ErrNotSupportNode
}

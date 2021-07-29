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

// Package storage is the common storage layout for dag and data store
package storage

import (
	"fmt"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	merkledag_pb "github.com/ipfs/go-merkledag/pb"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"

	"sci_hub_p2p/internal/utils"
	"sci_hub_p2p/pkg/consts"
	"sci_hub_p2p/pkg/pb"
)

func ReadFileStoreNode(b *bbolt.Bucket, c cid.Cid) (ipld.Node, error) {
	var v = &pb.Block{}
	data := b.Get(c.Bytes())
	if data == nil {
		return nil, ipld.ErrNotFound
	}

	err := proto.Unmarshal(data, v)
	if err != nil {
		return nil, errors.Wrap(err, "can't marshal persist.Record to binary")
	}

	p, err := utils.ReadFileAt(v.Filename, v.Offset, v.Size)
	if err != nil {
		return nil, errors.Wrap(err, "filed to read from disk")
	}

	block, err := blocks.NewBlockWithCid(p, c)

	return &merkledag.RawNode{Block: block}, errors.Wrap(err, "failed to create block")
}

func SaveFileStoreMeta(tx *bbolt.Tx, c cid.Cid, name string, offset, size int64) error {
	nb := tx.Bucket(consts.NodeBucketName())
	bb := tx.Bucket(consts.BlockBucketName())

	var block = pb.Block{
		Type:     pb.BlockType_file,
		CID:      c.Bytes(),
		Offset:   offset,
		Size:     size,
		Filename: name,
	}

	value, err := proto.Marshal(&block)
	if err != nil {
		return errors.Wrap(err, "failed to marshal block record to bytes")
	}

	err = bb.Put(c.Hash(), value)
	if err != nil {
		return errors.Wrap(err, "failed to save block record to database")
	}

	return errors.Wrap(nb.Put(c.Bytes(), value), "failed to save data to database")
}

func SaveProtoNode(tx *bbolt.Tx, c cid.Cid, n *merkledag.ProtoNode) error {
	nb := tx.Bucket(consts.NodeBucketName())
	bb := tx.Bucket(consts.BlockBucketName())

	var v = pb.Block{Type: pb.BlockType_proto, CID: c.Bytes(), Size: int64(len(n.RawData()))}
	value, err := proto.Marshal(&v)
	if err != nil {
		return errors.Wrap(err, "failed to marshal block record to bytes")
	}

	err = bb.Put(c.Hash(), value)
	if err != nil {
		return errors.Wrap(err, "failed to save block record to database")
	}

	return errors.Wrap(nb.Put(c.Bytes(), n.RawData()), "failed to save node record to database")
}

func ReadProtoNode(nb *bbolt.Bucket, c cid.Cid) (ipld.Node, error) {
	data := nb.Get(c.Bytes())
	if data == nil {
		return nil, ipld.ErrNotFound
	}
	v, err := unmarshal(data, c)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal data to `merkledag.ProtoNode`")
	}

	return v, nil
}

// from https://github.com/ipfs/go-merkledag/blob/v0.3.2/coding.go#L25-L46
func unmarshal(encoded []byte, c cid.Cid) (*merkledag.ProtoNode, error) {
	var n = &merkledag.ProtoNode{}

	n.SetCidBuilder(cid.V1Builder{
		Codec:    c.Prefix().Codec,
		MhType:   c.Prefix().MhType,
		MhLength: -1,
	})

	var pbn merkledag_pb.PBNode

	if err := pbn.Unmarshal(encoded); err != nil {
		return nil, errors.Wrap(err, "unmarshal failed")
	}

	pbnl := pbn.GetLinks()
	for i, l := range pbnl {
		c, err := cid.Cast(l.GetHash())
		if err != nil {
			return nil, fmt.Errorf("link hash #%d is not valid multihash. %w", i, err)
		}
		err = n.AddRawLink(l.GetName(), &ipld.Link{
			Name: l.GetName(),
			Size: l.GetTsize(),
			Cid:  c,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to add link to node")
		}
	}

	n.SetData(pbn.GetData())

	return n, nil
}

var ErrNotSupportNode = errors.New("not supported error")

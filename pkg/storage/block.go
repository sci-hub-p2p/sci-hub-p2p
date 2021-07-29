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

package storage

import (
	ds "github.com/ipfs/go-datastore"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"sci_hub_p2p/internal/utils"
	"sci_hub_p2p/pkg/consts"
	"sci_hub_p2p/pkg/pb"
)

func ReadLen(tx *bbolt.Tx, log *zap.Logger, mh []byte) (int, error) {
	bb := tx.Bucket(consts.BlockBucketName())
	nb := tx.Bucket(consts.NodeBucketName())

	v := bb.Get(mh)
	if v == nil {
		return -1, ds.ErrNotFound
	}

	var r = &pb.Block{}

	if err := proto.Unmarshal(v, r); err != nil {
		return -1, errors.Wrap(err, "failed to decode block Record from database raw value")
	}
	log.Debug("find block in KV, type", zap.String("type", r.Type.String()))
	switch r.Type {
	case pb.BlockType_proto:
		n := nb.Get(r.CID)
		if n == nil {
			return -1, ds.ErrNotFound
		}

		return len(n), nil
	case pb.BlockType_file:
		return int(r.Size), nil
	}

	return -1, ErrNotValidBlock
}

func ReadBlock(tx *bbolt.Tx, mh []byte) ([]byte, error) {
	bb := tx.Bucket(consts.BlockBucketName())
	nb := tx.Bucket(consts.NodeBucketName())

	v := bb.Get(mh)
	if v == nil {
		return nil, ds.ErrNotFound
	}

	var r = &pb.Block{}
	if err := proto.Unmarshal(v, r); err != nil {
		return nil, errors.Wrap(err, "failed to decode block Record from database raw value")
	}
	switch r.Type {
	case pb.BlockType_proto:
		p := nb.Get(r.CID)
		if p == nil {
			return nil, errors.Wrap(ds.ErrNotFound, "can't read proto node from node bucket")
		}

		return p, nil
	case pb.BlockType_file:
		var p, err = utils.ReadFileAt(r.Filename, r.Offset, r.Size)

		return p, errors.Wrap(err, "can't read file block from disk")
	}

	return nil, ErrNotValidBlock
}

var ErrNotValidBlock = errors.New("not valid record in block bucket")

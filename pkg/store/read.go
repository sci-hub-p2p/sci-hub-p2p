// Copyright 2021 Trim21 <trim21.me@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.

package store

import (
	"github.com/dgraph-io/ristretto"
	ds "github.com/ipfs/go-datastore"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"sci_hub_p2p/internal/utils"
	"sci_hub_p2p/pkg/consts"
	"sci_hub_p2p/pkg/pb"
)

var ErrNotValidBlock = errors.New("not valid record in block bucket")

func readBlock(tx *bbolt.Tx, mh []byte) ([]byte, bool, error) {
	bb := tx.Bucket(consts.BlockBucketName())
	nb := tx.Bucket(consts.NodeBucketName())

	v := bb.Get(mh)
	if v == nil {
		return nil, false, ds.ErrNotFound
	}

	var r = &pb.Block{}
	if err := proto.Unmarshal(v, r); err != nil {
		return nil, false, errors.Wrap(err, "failed to decode block Record from database raw value")
	}

	switch r.Type {
	case pb.BlockType_proto:
		p := nb.Get(r.CID)
		if p == nil {
			return nil, false, errors.Wrap(ds.ErrNotFound, "can't read proto node from node bucket")
		}

		return p, false, nil
	case pb.BlockType_file:
		var p, err = utils.ReadFileAt(r.Filename, r.Offset, r.Size)

		return p, true, errors.Wrap(err, "can't read file block from disk")
	}

	return nil, false, ErrNotValidBlock
}

func CachedReadBlockW(db *bbolt.DB, cache *ristretto.Cache, mh []byte) ([]byte, error) {
	value, found := cache.Get(mh)
	if found {
		v, ok := value.([]byte)
		if ok {
			return v, nil
		}

		cache.Del(mh)
	}

	var out []byte
	var shouldCache bool
	err := db.View(func(tx *bbolt.Tx) error {
		v, isFile, err := readBlock(tx, mh)
		if err != nil {
			return err
		}
		if isFile {
			out = v
			shouldCache = true
		} else {
			out = make([]byte, len(v))
			copy(out, v)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if shouldCache {
		cache.Set(mh, out, int64(len(out)))
		cache.Wait()
	}

	return out, errors.Wrap(err, "can't read file block from disk")
}

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

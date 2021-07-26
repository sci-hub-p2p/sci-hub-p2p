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

package store

import (
	"sync"

	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"sci_hub_p2p/internal/utils"
	"sci_hub_p2p/pkg/dagserv"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/variable"
)

// Here are some basic store implementations.

var _ ds.Datastore = (*MapDataStore)(nil)

// MapDataStore uses a standard Go map for internal storage.
type MapDataStore struct {
	dag       ipld.DAGService
	db        *bbolt.DB
	values    map[ds.Key][]byte
	logger    *zap.Logger
	keysCache sync.Map
	sync.RWMutex
}

// NewMapDatastore constructs a MapDataStore. It is _not_ thread-safe by
// default, wrap using sync.MutexWrap if you need thread safety (the answer here
// is usually yes).
func NewMapDatastore(db *bbolt.DB) (d *MapDataStore) {
	return &MapDataStore{
		values: make(map[ds.Key][]byte),
		db:     db,
		dag:    dagserv.New(db),
		logger: logger.WithLogger("MapDataStore"),
	}
}

// Put implements Datastore.Put.
func (d *MapDataStore) Put(key ds.Key, value []byte) (err error) {
	if key.IsDescendantOf(topLevelBlockKey) {
		logger.Debug("try to put block, put it in memory")
	}
	d.Lock()
	d.values[key] = value
	d.Unlock()

	return nil
}

// Sync implements Datastore.Sync.
func (d *MapDataStore) Sync(prefix ds.Key) error {
	return errors.Wrap(d.db.Sync(), "failed to sync bbolt DB")
}

func (d *MapDataStore) Get(key ds.Key) ([]byte, error) {
	var log = d.logger.With(logger.Key(key))
	log.Debug("try to get block, check it in memory first")

	if !key.IsDescendantOf(topLevelBlockKey) {
		d.RLock()
		val, found := d.values[key]
		d.RUnlock()
		if !found {
			return nil, ds.ErrNotFound
		}

		return val, nil
	}

	// /blocks/{multi hash}

	log.Debug("didn't find in memory, now check it in KV database")
	mh, err := dshelp.DsKeyToMultihash(ds.NewKey(key.BaseNamespace()))
	if err != nil {
		d.logger.Error("block key is not a valid multi hash", zap.Error(err))

		return nil, errors.Wrapf(err, "failed to decode key to multihash for key %s", key)
	}

	var p []byte

	err = d.db.View(func(tx *bbolt.Tx) error {
		p, err = readBlock(tx, mh)
		if p != nil {
			log.Debug("find in KV")
		}

		return err
	})
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return nil, err
		}
		log.Debug("read block got", zap.Error(err))

		return nil, errors.Wrap(err, "can't read from disk")
	}

	return p, nil
}

// Has implements Datastore.Has.
func (d *MapDataStore) Has(key ds.Key) (exists bool, err error) {
	d.RLock()
	_, found := d.values[key]
	d.RUnlock()
	if found {
		return found, nil
	}

	_, found = d.keysCache.Load(key)

	if !found {
		mh, err := dshelp.DsKeyToMultihash(ds.NewKey(key.BaseNamespace()))
		if err != nil {
			return false, errors.Wrap(err, "failed to decode key to multi HASH")
		}
		_ = d.db.View(func(tx *bbolt.Tx) error {
			b := tx.Bucket(variable.BlockBucketName())
			if b.Get(mh) != nil {
				found = true
			}

			return nil
		})
	}

	return found, nil
}

// GetSize implements Datastore.GetSize.
func (d *MapDataStore) GetSize(key ds.Key) (size int, err error) {
	d.RLock()

	v, found := d.values[key]
	if found {
		d.RUnlock()

		return len(v), nil
	}
	d.RUnlock()

	d.logger.Debug("get size of from kV", logger.Key(key))
	var l = -1
	var mh []byte
	if !found {
		mh, err = dshelp.DsKeyToMultihash(ds.NewKey(key.BaseNamespace()))
		if err != nil {
			return 0, errors.Wrap(err, "failed to decode key to multi HASH")
		}

		err = d.db.View(func(tx *bbolt.Tx) error {
			l, err = readLen(tx, d.logger, mh)

			return err
		})
	}

	return l, err
}

// Delete implements Datastore.Delete.
func (d *MapDataStore) Delete(key ds.Key) (err error) {
	d.Lock()
	delete(d.values, key)
	d.Unlock()

	return nil
}

// Query is copied from go-ds-bolt and modified.
func (d *MapDataStore) Query(q dsq.Query) (dsq.Results, error) {
	var log = d.logger.Named("Query").With(zap.String("prefix", q.Prefix))
	if q.Prefix != "/blocks" {
		log.Debug("none `/blocks` query, only search in memory")
		d.RLock()
		defer d.RUnlock()
		re := make([]dsq.Entry, 0, len(d.values))
		for k, v := range d.values {
			e := dsq.Entry{Key: k.String(), Size: len(v)}
			if !q.KeysOnly {
				e.Value = v
			}
			re = append(re, e)
		}
		r := dsq.ResultsWithEntries(q, re)
		r = dsq.NaiveQueryApply(q, r)

		return r, nil
	}

	log.Debug("try to query from KV")

	return queryBolt(d, q, log)
}

func (d *MapDataStore) Batch() (ds.Batch, error) {
	return ds.NewBasicBatch(d), nil
}

func (d *MapDataStore) Close() error {
	return nil
}

func readLen(tx *bbolt.Tx, log *zap.Logger, mh []byte) (int, error) {
	bb := tx.Bucket(variable.BlockBucketName())
	nb := tx.Bucket(variable.NodeBucketName())

	v := bb.Get(mh)
	if v == nil {
		return -1, ds.ErrNotFound
	}

	var r = &dagserv.Block{}

	if err := proto.Unmarshal(v, r); err != nil {
		return -1, errors.Wrap(err, "failed to decode block Record from database raw value")
	}
	log.Debug("find block in KV, type", zap.String("type", r.Type.String()))
	switch r.Type {
	case dagserv.BlockType_proto:
		n := nb.Get(r.CID)
		if n == nil {
			return -1, ds.ErrNotFound
		}

		return len(n), nil
	case dagserv.BlockType_file:
		return int(r.Size), nil
	}

	return -1, errNotValidBlock
}

func readBlock(tx *bbolt.Tx, mh []byte) ([]byte, error) {
	bb := tx.Bucket(variable.BlockBucketName())
	nb := tx.Bucket(variable.NodeBucketName())

	v := bb.Get(mh)
	if v == nil {
		return nil, ds.ErrNotFound
	}

	var r = &dagserv.Block{}
	if err := proto.Unmarshal(v, r); err != nil {
		return nil, errors.Wrap(err, "failed to decode block Record from database raw value")
	}
	switch r.Type {
	case dagserv.BlockType_proto:
		p := nb.Get(r.CID)
		if p == nil {
			return nil, errors.Wrap(ds.ErrNotFound, "can't read proto node from node bucket")
		}

		return p, nil
	case dagserv.BlockType_file:
		var p, err = utils.ReadFileAt(r.Filename, r.Offset, r.Size)

		return p, errors.Wrap(err, "can't read file block from disk")
	}

	return nil, errNotValidBlock
}

var errNotValidBlock = errors.New("not valid record in block bucket")

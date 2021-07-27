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
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"

	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/storage"
	"sci_hub_p2p/pkg/variable"
)

// Here are some basic store implementations.

var _ ds.Datastore = (*MapDataStore)(nil)

// MapDataStore uses a standard Go map for internal storage.
type MapDataStore struct {
	db            *bbolt.DB
	values        map[ds.Key][]byte
	logger        *zap.Logger
	keysSizeCache sync.Map
	sync.RWMutex
}

// NewArchiveFallbackDatastore constructs a MapDataStore. It is _not_ thread-safe by
// default, wrap using sync.MutexWrap if you need thread safety (the answer here
// is usually yes).
func NewArchiveFallbackDatastore(db *bbolt.DB) (d *MapDataStore) {
	return &MapDataStore{
		values: make(map[ds.Key][]byte),
		db:     db,
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
func (d *MapDataStore) Sync(_ ds.Key) error {
	return errors.Wrap(d.db.Sync(), "failed to sync bbolt DB")
}

func (d *MapDataStore) Get(key ds.Key) ([]byte, error) {
	var log = d.logger.With(logger.Key(key))
	log.Debug("try to get block, check it in memory first")

	if !isBlockKey(key) {
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
		var e error
		p, e = storage.ReadBlock(tx, mh)
		if p != nil {
			log.Debug("find in KV")
		}

		return errors.Wrap(e, "failed to read block")
	})

	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return nil, ds.ErrNotFound
		}
		log.Debug("read block got", zap.Error(err))

		return nil, err
	}
	d.keysSizeCache.Store(key, len(p))

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

	_, found = d.keysSizeCache.Load(key)

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
func (d *MapDataStore) GetSize(key ds.Key) (int, error) {
	var log = d.logger.With(logger.Key(key))
	if !isBlockKey(key) {
		log.Debug("non /blocks key, lookup in map")
		d.RLock()
		v, found := d.values[key]
		d.RUnlock()
		if found {
			return len(v), nil
		}

		return 0, ds.ErrNotFound
	}

	log.Debug("didn't find key in kv, try get size from cache")
	v, ok := d.keysSizeCache.Load(key.String())
	if ok {
		return v.(int), nil
	}

	log.Debug("lookup size of from kV")
	var l = -1

	mh, err := dshelp.DsKeyToMultihash(ds.NewKey(key.BaseNamespace()))
	if err != nil {
		return 0, errors.Wrap(err, "failed to decode key to multi HASH")
	}

	err = d.db.View(func(tx *bbolt.Tx) error {
		var e error
		l, e = storage.ReadLen(tx, d.logger, mh)

		return errors.Wrap(e, "failed to get size from database")
	})

	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return 0, ds.ErrNotFound
		}
		log.Debug("read block got", zap.Error(err))

		return 0, err
	}
	d.keysSizeCache.Store(key, l)

	return l, nil
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

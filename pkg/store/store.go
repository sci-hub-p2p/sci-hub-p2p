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
	"sync"

	"github.com/dgraph-io/ristretto"
	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"

	"sci_hub_p2p/pkg/consts"
	"sci_hub_p2p/pkg/logger"
)

var _ ds.Datastore = (*MapDataStore)(nil)

type MapDataStore struct {
	db            *bbolt.DB
	cache         *ristretto.Cache
	values        map[ds.Key][]byte
	logger        *zap.Logger
	keysSizeCache sync.Map // cache block key content size
	sync.RWMutex
}

const KB256 = 256 * 1024
const defaultBufferItems = 64 // number of keys per Get buffer.

func NewArchiveFallbackDatastore(db *bbolt.DB, cacheSize int64) (d *MapDataStore) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: cacheSize/KB256 + 1,
		MaxCost:     cacheSize,
		BufferItems: defaultBufferItems,
	})
	if err != nil {
		panic(err)
	}

	return &MapDataStore{
		values: make(map[ds.Key][]byte),
		db:     db,
		logger: logger.WithLogger("MapDataStore"),
		cache:  cache,
	}
}

// Put implements Datastore.Put.
func (d *MapDataStore) Put(key ds.Key, value []byte) error {
	if isBlockKey(topLevelBlockKey) {
		logger.Debug("try to put block, just skip")

		return nil
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
	var log = d.logger.Named("Get").With(logger.Key(key))
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

	var out []byte

	out, err = CachedReadBlockW(d.db, d.cache, mh)

	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			return nil, ds.ErrNotFound
		}

		log.Debug("read block got", zap.Error(err))

		return nil, err
	}

	d.keysSizeCache.Store(key, len(out))

	return out, nil
}

// Has returns whether the `key` is mapped to a `value`.
// In some contexts, it may be much cheaper only to check for existence of
// a value, rather than retrieving the value itself. (e.g. HTTP HEAD).
// The default implementation is found in `GetBackedHas`.
func (d *MapDataStore) Has(key ds.Key) (exists bool, err error) {
	if !isBlockKey(key) {
		// lookup in map for non blocks key
		d.RLock()
		_, found := d.values[key]
		d.RUnlock()

		if !found {
			return false, nil
		}

		return true, nil
	}

	if _, found := d.keysSizeCache.Load(key); found {
		return true, nil
	}

	var found bool

	mh, err := dshelp.DsKeyToMultihash(ds.NewKey(key.BaseNamespace()))
	if err != nil {
		return false, errors.Wrap(err, "failed to decode key to multi HASH")
	}

	_ = d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(consts.BlockBucketName())
		if b.Get(mh) != nil {
			found = true
		}

		return nil
	})

	return found, nil
}

func (d *MapDataStore) cacheGet(key interface{}) []byte {
	v, ok := d.cache.Get(key)
	if !ok {
		return nil
	}

	if b, ok := v.([]byte); ok {
		return b
	}

	d.cache.Del(key)

	return nil
}

// GetSize implements Datastore.GetSize.
func (d *MapDataStore) GetSize(key ds.Key) (int, error) {
	var log = d.logger.Named("GetSize").With(logger.Key(key))
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

	log.Debug("didn't find key in map, try get size from cache")

	if v, ok := d.keysSizeCache.Load(key.String()); ok {
		return v.(int), nil
	}

	mh, err := dshelp.DsKeyToMultihash(ds.NewKey(key.BaseNamespace()))
	if err != nil {
		return 0, errors.Wrap(err, "failed to decode key to multi HASH")
	}

	log.Debug("lookup size in from Cache")

	if v := d.cacheGet([]byte(mh)); v != nil {
		return len(v), nil
	}

	log.Debug("lookup size of from kV")
	var l = -1

	err = d.db.View(func(tx *bbolt.Tx) error {
		var e error
		l, e = ReadLen(tx, d.logger, mh)

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

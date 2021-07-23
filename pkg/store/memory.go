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
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/jbenet/goprocess"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"

	"sci_hub_p2p/internal/utils"
	"sci_hub_p2p/pkg/dagserv"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/variable"
)

// Here are some basic store implementations.

var _ ds.Datastore = &MapDatastore{}

// MapDatastore uses a standard Go map for internal storage.
type MapDatastore struct {
	values    map[ds.Key][]byte
	db        *bbolt.DB
	dag       ipld.DAGService
	keysCache sync.Map
	keyCached bool
	sync.RWMutex
}

func showFirst32(p []byte) {
	l := len(p)
	if l == 0 {
		return
	}
	if l >= 32 {
		fmt.Println(hex.Dump(p[:32]))
	} else {
		fmt.Println(hex.Dump(p))
	}
}

func List(s *MapDatastore) {
	fmt.Println("start list key")
	for key, value := range s.values {
		fmt.Println(key)
		showFirst32(value)
	}
	fmt.Println("stop list key")
}

// NewMapDatastore constructs a MapDatastore. It is _not_ thread-safe by
// default, wrap using sync.MutexWrap if you need thread safety (the answer here
// is usually yes).
func NewMapDatastore(db *bbolt.DB) (d *MapDatastore) {
	return &MapDatastore{
		values: make(map[ds.Key][]byte),
		db:     db,
		dag:    dagserv.New(db, 0),
	}
}

// Put implements Datastore.Put.
func (d *MapDatastore) Put(key ds.Key, value []byte) (err error) {
	if key.IsDescendantOf(topLevelBlockKey) {
		fmt.Println("try to put block, put it in memory")
	}
	d.Lock()
	d.values[key] = value
	d.Unlock()

	return nil
}

// Sync implements Datastore.Sync.
func (d *MapDatastore) Sync(prefix ds.Key) error {
	return errors.Wrap(d.db.Sync(), "failed to sync bbolt DB")
}

func (d *MapDatastore) Get(key ds.Key) ([]byte, error) {
	if key.IsDescendantOf(topLevelBlockKey) {
		fmt.Println("try to get block, check it in memory first")
		d.RLock()
		if val, ok := d.values[key]; ok {
			d.RUnlock()

			return val, nil
		}
		d.RUnlock()
		fmt.Println("didn't find in memory, now check it in KV database")
		mh, err := dshelp.DsKeyToMultihash(ds.NewKey(key.BaseNamespace()))
		if err != nil {
			return nil, errors.Wrap(err, "failed to decode key to multihash")
		}

		var p []byte

		err = d.db.View(func(tx *bbolt.Tx) error {
			b := tx.Bucket(variable.BlockBucketName())
			p, err = readBlock(b, mh)

			return err
		})
		if err != nil {
			if errors.Is(err, ds.ErrNotFound) {
				fmt.Println("can't find", key)
				return nil, err
			}
			fmt.Println("read block got", err)
			return nil, errors.Wrap(err, "can't read from disk")
		}

		return p, nil
	}
	// non block keysCache
	d.RLock()
	val, found := d.values[key]
	d.RUnlock()

	if !found {
		return nil, ds.ErrNotFound
	}

	return val, nil

	// c, err := dshelp.DsKeyToMultihash(key)
	// if err != nil {
	// 	return nil, err
	// }
	// var val []byte
	// err = d.db.View(func(tx *bbolt.Tx) error {
	// 	b := tx.Bucket(variable.NodeBucketName())
	// 	val = b.Get(c)
	// 	return nil
	// })
	// if err != nil {
	// 	return nil, err
	// }
	// if val == nil {
	//
	// 	return nil, ds.ErrNotFound
	// }
	// return val, nil
}

// Has implements Datastore.Has.
func (d *MapDatastore) Has(key ds.Key) (exists bool, err error) {
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
func (d *MapDatastore) GetSize(key ds.Key) (size int, err error) {
	d.RLock()
	defer d.RUnlock()

	v, found := d.values[key]
	if found {
		return len(v), nil
	}

	var l int = -1
	if !found {
		mh, err := dshelp.DsKeyToMultihash(ds.NewKey(key.BaseNamespace()))
		if err != nil {
			return 0, errors.Wrap(err, "failed to decode key to multi HASH")
		}
		err = d.db.View(func(tx *bbolt.Tx) error {
			logger.WithField("func", "readLen").Infoln("start read len", key)
			l, err = readLen(tx.Bucket(variable.BlockBucketName()), mh)
			logger.WithField("func", "readLen").Infoln("end read len", key)
			return err
		})
	}

	return l, err
}

// Delete implements Datastore.Delete.
func (d *MapDatastore) Delete(key ds.Key) (err error) {
	d.Lock()
	delete(d.values, key)
	d.Unlock()

	return nil
}

var errStop = errors.New("stop iter")

// Query implements Datastore.Query
// func (d *MapDatastore) Query(q dsq.Query) (dsq.Results, error) {
// 	d.RLock()
//
// 	re := make([]dsq.Entry, 0, len(d.values))
// 	for k, v := range d.values {
// 		e := dsq.Entry{Key: k.String(), Size: len(v)}
// 		if !q.KeysOnly {
// 			e.Value = v
// 		}
// 		re = append(re, e)
// 	}
// 	r := dsq.ResultsWithEntries(q, re)
// 	r = dsq.NaiveQueryApply(q, r)
// 	return r, nil
// }

// Query implements Datastore.Query.
func (d *MapDatastore) Query(q dsq.Query) (dsq.Results, error) {
	d.RLock()
	re := make([]dsq.Entry, 0, len(d.values))
	for k, v := range d.values {
		e := dsq.Entry{Key: k.String(), Size: len(v)}
		if !q.KeysOnly {
			e.Value = v
		}
		re = append(re, e)
	}
	d.RUnlock()
	var c = make(chan dsq.Result, 10)
	r := result{
		closed: false,
		c:      c,
		q:      q,
	}

	go func() {
		if strings.HasPrefix(q.Prefix, "/blocks") {
			if d.keyCached && q.KeysOnly {
				d.keysCache.Range(func(key, value interface{}) bool {
					if r.closed {
						return true
					}
					e := dsq.Entry{Key: key.(string), Size: value.(int)}
					c <- dsq.Result{Entry: e}
					return false
				})
			} else {
				_ = d.db.View(func(tx *bbolt.Tx) error {
					_ = tx.Bucket(variable.BlockBucketName()).ForEach(func(k, v []byte) error {
						if r.closed {
							return errStop
						}
						e := dsq.Entry{Key: topLevelBlockKey.Child(dshelp.MultihashToDsKey(k)).String(), Size: len(v)}
						d.keysCache.Store(e.Key, len(v))
						if !q.KeysOnly {
							e.Value = v
						}
						c <- dsq.Result{Entry: e}
						return nil
					})
					return nil
				})
				d.keyCached = true
			}
		}
	}()

	return r, nil
}

type result struct {
	closed bool
	c      chan dsq.Result
	q      dsq.Query
}

func (r result) Query() dsq.Query {
	return r.q
}

func (r result) Next() <-chan dsq.Result {
	return r.c
}

func (r result) NextSync() (dsq.Result, bool) {
	if r.closed {
		return dsq.Result{}, true
	}
	for i := range r.c {
		return i, false
	}
	return dsq.Result{}, true
}

func (r result) Rest() ([]dsq.Entry, error) {
	var re []dsq.Entry
	for i := range r.c {
		if i.Error != nil {
			return nil, i.Error
		}
		re = append(re, i.Entry)
	}
	return re, nil
}

func (r result) Close() error {
	r.closed = true
	return nil
}

func (r result) Process() goprocess.Process {
	return nil
}

func (d *MapDatastore) Batch() (ds.Batch, error) {
	return ds.NewBasicBatch(d), nil
}

func (d *MapDatastore) Close() error {
	return nil
}

func readLen(b *bbolt.Bucket, mh []byte) (int, error) {
	fmt.Println("readLen: try to read metadata from KV")
	defer fmt.Println("exit readLen function")
	v := b.Get(mh)
	if v == nil {
		return -1, ds.ErrNotFound
	}
	var r = &dagserv.Block{}
	err := proto.Unmarshal(v, r)
	if err != nil {
		return -1, errors.Wrap(err, "failed to decode block Record from database raw value")
	}
	switch r.Type {
	case dagserv.BlockType_proto:
		return len(v), nil
	case dagserv.BlockType_file:
		return int(r.Length), nil
	}

	return -1, errNotValidBlock
}

func readBlock(b *bbolt.Bucket, mh []byte) ([]byte, error) {
	fmt.Println("readBlock: try to read metadata from KV")
	defer fmt.Println("exit readBlock function")
	v := b.Get(mh)
	if v == nil {
		return nil, ds.ErrNotFound
	}
	var r = &dagserv.Block{}
	err := proto.Unmarshal(v, r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode block Record from database raw value")
	}
	switch r.Type {
	case dagserv.BlockType_proto:
		return v, nil
	case dagserv.BlockType_file:
		var p, err = utils.ReadFileAt(r.Filename, int64(r.Offset), int64(r.Length))

		return p, errors.Wrap(err, "can't read file block from disk")
	}

	return nil, errNotValidBlock
}

var errNotValidBlock = errors.New("not valid record")

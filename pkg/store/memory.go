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
	"context"
	"fmt"
	"log"

	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/pkg/dagserv"
	"sci_hub_p2p/pkg/variable"
)

// Here are some basic store implementations.

// MapDatastore uses a standard Go map for internal storage.
type MapDatastore struct {
	values map[ds.Key][]byte
	db     *bbolt.DB
	dag    ipld.DAGService
}

var ErrWriteNotAllowed = errors.New("write data not allowed")

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

// Put implements Datastore.Put
func (d *MapDatastore) Put(key ds.Key, value []byte) (err error) {
	return ErrWriteNotAllowed
}

// Sync implements Datastore.Sync
func (d *MapDatastore) Sync(prefix ds.Key) error {
	return nil
}

// Get implements Datastore.Get
func (d *MapDatastore) Get(key ds.Key) (value []byte, err error) {
	c, err := dshelp.DsKeyToMultihash(key)
	if err != nil {
		return nil, err
	}
	d.dag.Get(context.TODO(), c)
	var val []byte
	err = d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(variable.NodeBucketName())
		val = b.Get(c.Bytes())
		return nil
	})
	if err != nil {
		return nil, err
	}
	if val == nil {

		return nil, ds.ErrNotFound
	}
	return val, nil
}

// Has implements Datastore.Has
func (d *MapDatastore) Has(key ds.Key) (exists bool, err error) {
	_, found := d.values[key]
	return found, nil
}

// GetSize implements Datastore.GetSize
func (d *MapDatastore) GetSize(key ds.Key) (size int, err error) {
	if v, found := d.values[key]; found {
		return len(v), nil
	}
	return -1, ds.ErrNotFound
}

// Delete implements Datastore.Delete
func (d *MapDatastore) Delete(key ds.Key) (err error) {
	delete(d.values, key)
	return nil
}

// Query implements Datastore.Query
func (d *MapDatastore) Query(q dsq.Query) (dsq.Results, error) {
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

func (d *MapDatastore) Batch() (ds.Batch, error) {
	return ds.NewBasicBatch(d), nil
}

func (d *MapDatastore) Close() error {
	return nil
}

// LogDatastore logs all accesses through the store.
type LogDatastore struct {
	Name  string
	child ds.Datastore
}

// Shim is a store which has a child.
type Shim interface {
	ds.Datastore

	Children() []ds.Datastore
}

// NewLogDatastore constructs a log store.
func NewLogDatastore(ds ds.Datastore, name string) *LogDatastore {
	if len(name) < 1 {
		name = "LogDatastore"
	}
	return &LogDatastore{Name: name, child: ds}
}

// Children implements Shim
func (d *LogDatastore) Children() []ds.Datastore {
	return []ds.Datastore{d.child}
}

// Put implements Datastore.Put
func (d *LogDatastore) Put(key ds.Key, value []byte) (err error) {
	log.Printf("%s: Put %s\n", d.Name, key)
	// log.Printf("%s: Put %s ```%s```", d.Name, key, value)
	return d.child.Put(key, value)
}

// Sync implements Datastore.Sync
func (d *LogDatastore) Sync(prefix ds.Key) error {
	log.Printf("%s: Sync %s\n", d.Name, prefix)
	return d.child.Sync(prefix)
}

// Get implements Datastore.Get
func (d *LogDatastore) Get(key ds.Key) (value []byte, err error) {
	log.Printf("%s: Get %s\n", d.Name, key)
	if mh, err := dshelp.DsKeyToMultihash(key); err == nil {
		fmt.Println(mh)
	}

	return d.child.Get(key)
}

// Has implements Datastore.Has
func (d *LogDatastore) Has(key ds.Key) (exists bool, err error) {
	log.Printf("%s: Has %s\n", d.Name, key)
	return d.child.Has(key)
}

// GetSize implements Datastore.GetSize
func (d *LogDatastore) GetSize(key ds.Key) (size int, err error) {
	log.Printf("%s: GetSize %s\n", d.Name, key)
	return d.child.GetSize(key)
}

// Delete implements Datastore.Delete
func (d *LogDatastore) Delete(key ds.Key) (err error) {
	log.Printf("%s: Delete %s\n", d.Name, key)
	return d.child.Delete(key)
}

// DiskUsage implements the PersistentDatastore interface.
func (d *LogDatastore) DiskUsage() (uint64, error) {
	log.Printf("%s: DiskUsage\n", d.Name)
	return ds.DiskUsage(d.child)
}

// Query implements Datastore.Query
func (d *LogDatastore) Query(q dsq.Query) (dsq.Results, error) {
	log.Printf("%s: Query\n", d.Name)
	log.Printf("%s: q.Prefix: %s\n", d.Name, q.Prefix)
	log.Printf("%s: q.KeysOnly: %v\n", d.Name, q.KeysOnly)
	log.Printf("%s: q.Filters: %d\n", d.Name, len(q.Filters))
	log.Printf("%s: q.Orders: %d\n", d.Name, len(q.Orders))
	log.Printf("%s: q.Offset: %d\n", d.Name, q.Offset)

	return d.child.Query(q)
}

// LogBatch logs all accesses through the batch.
type LogBatch struct {
	Name  string
	child ds.Batch
}

func (d *LogDatastore) Batch() (ds.Batch, error) {
	log.Printf("%s: Batch\n", d.Name)
	if bds, ok := d.child.(ds.Batching); ok {
		b, err := bds.Batch()

		if err != nil {
			return nil, err
		}
		return &LogBatch{
			Name:  d.Name,
			child: b,
		}, nil
	}
	return nil, ds.ErrBatchUnsupported
}

// Put implements Batch.Put
func (d *LogBatch) Put(key ds.Key, value []byte) (err error) {
	log.Printf("%s: BatchPut %s\n", d.Name, key)
	// log.Printf("%s: Put %s ```%s```", d.Name, key, value)
	return d.child.Put(key, value)
}

// Delete implements Batch.Delete
func (d *LogBatch) Delete(key ds.Key) (err error) {
	log.Printf("%s: BatchDelete %s\n", d.Name, key)
	return d.child.Delete(key)
}

// Commit implements Batch.Commit
func (d *LogBatch) Commit() (err error) {
	log.Printf("%s: BatchCommit\n", d.Name)
	return d.child.Commit()
}

func (d *LogDatastore) Close() error {
	log.Printf("%s: Close\n", d.Name)
	return d.child.Close()
}

func (d *LogDatastore) Check() error {
	if c, ok := d.child.(ds.CheckedDatastore); ok {
		return c.Check()
	}
	return nil
}

func (d *LogDatastore) Scrub() error {
	if c, ok := d.child.(ds.ScrubbedDatastore); ok {
		return c.Scrub()
	}
	return nil
}

func (d *LogDatastore) CollectGarbage() error {
	if c, ok := d.child.(ds.GCDatastore); ok {
		return c.CollectGarbage()
	}
	return nil
}

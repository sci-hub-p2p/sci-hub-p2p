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

	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/pkg/dagserv"
)

// Here are some basic store implementations.

var _ ds.Datastore = &MapDatastore{}

// MapDatastore uses a standard Go map for internal storage.
type MapDatastore struct {
	values map[ds.Key][]byte
	db     *bbolt.DB
	dag    ipld.DAGService
}

var ErrWriteNotAllowed = errors.New("write data not allowed")

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
	d.values[key] = value

	return nil
}

// Sync implements Datastore.Sync.
func (d *MapDatastore) Sync(prefix ds.Key) error {
	return nil
}

func (d *MapDatastore) Get(key ds.Key) (value []byte, err error) {
	val, found := d.values[key]
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
	_, found := d.values[key]

	return found, nil
}

// GetSize implements Datastore.GetSize.
func (d *MapDatastore) GetSize(key ds.Key) (size int, err error) {
	if v, found := d.values[key]; found {
		return len(v), nil
	}

	return -1, ds.ErrNotFound
}

// Delete implements Datastore.Delete.
func (d *MapDatastore) Delete(key ds.Key) (err error) {
	delete(d.values, key)

	return nil
}

// Query implements Datastore.Query.
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

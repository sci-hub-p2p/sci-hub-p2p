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

//nolint:wrapcheck
package store

import (
	"fmt"

	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
)

var _ ds.Datastore = &LogDatastore{}

// LogDatastore logs all accesses through the store.
type LogDatastore struct {
	child ds.Datastore
	Name  string
}

// Shim is a store which has a child.
type Shim interface {
	ds.Datastore
	Children() []ds.Datastore
}

// NewLogDatastore constructs a fmt store.
func NewLogDatastore(ds ds.Datastore, name string) *LogDatastore {
	if len(name) < 1 {
		name = "LogDatastore"
	}

	return &LogDatastore{Name: name, child: ds}
}

// Children implements Shim.
func (d *LogDatastore) Children() []ds.Datastore {
	return []ds.Datastore{d.child}
}

// Put implements Datastore.Put.
func (d *LogDatastore) Put(key ds.Key, value []byte) (err error) {
	fmt.Printf("%s: Put %s\n", d.Name, key)
	// fmt.Printf("%s: Put %s ```%s```", d.Name, key, value)
	return d.child.Put(key, value)
}

// Sync implements Datastore.Sync.
func (d *LogDatastore) Sync(prefix ds.Key) error {
	fmt.Printf("%s: Sync %s\n", d.Name, prefix)

	return d.child.Sync(prefix)
}

// Get implements Datastore.Get.
func (d *LogDatastore) Get(key ds.Key) (value []byte, err error) {
	fmt.Printf("%s: Get %s\n", d.Name, key)
	if mh, err := dshelp.DsKeyToMultihash(key); err == nil {
		fmt.Println(mh)
	}

	return d.child.Get(key)
}

// Has implements Datastore.Has.
func (d *LogDatastore) Has(key ds.Key) (exists bool, err error) {
	fmt.Printf("%s: Has %s\n", d.Name, key)

	return d.child.Has(key)
}

// GetSize implements Datastore.GetSize.
func (d *LogDatastore) GetSize(key ds.Key) (size int, err error) {
	fmt.Printf("%s: GetSize %s\n", d.Name, key)

	return d.child.GetSize(key)
}

// Delete implements Datastore.Delete.
func (d *LogDatastore) Delete(key ds.Key) (err error) {
	fmt.Printf("%s: Delete %s\n", d.Name, key)

	return d.child.Delete(key)
}

// DiskUsage implements the PersistentDatastore interface.
func (d *LogDatastore) DiskUsage() (uint64, error) {
	fmt.Printf("%s: DiskUsage\n", d.Name)

	return ds.DiskUsage(d.child)
}

// Query implements Datastore.Query.
func (d *LogDatastore) Query(q dsq.Query) (dsq.Results, error) {
	fmt.Printf("%s: Query\n", d.Name)
	fmt.Printf("%s: q.Prefix: %s\n", d.Name, q.Prefix)
	fmt.Printf("%s: q.KeysOnly: %v\n", d.Name, q.KeysOnly)
	fmt.Printf("%s: q.Filters: %d\n", d.Name, len(q.Filters))
	fmt.Printf("%s: q.Orders: %d\n", d.Name, len(q.Orders))
	fmt.Printf("%s: q.Offset: %d\n", d.Name, q.Offset)

	return d.child.Query(q)
}

// LogBatch logs all accesses through the batch.
type LogBatch struct {
	child ds.Batch
	Name  string
}

func (d *LogDatastore) Batch() (ds.Batch, error) {
	fmt.Printf("%s: Batch\n", d.Name)
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

// Put implements Batch.Put.
func (d *LogBatch) Put(key ds.Key, value []byte) (err error) {
	fmt.Printf("%s: BatchPut %s\n", d.Name, key)
	// fmt.Printf("%s: Put %s ```%s```", d.Name, key, value)
	return d.child.Put(key, value)
}

// Delete implements Batch.Delete.
func (d *LogBatch) Delete(key ds.Key) (err error) {
	fmt.Printf("%s: BatchDelete %s\n", d.Name, key)

	return d.child.Delete(key)
}

// Commit implements Batch.Commit.
func (d *LogBatch) Commit() (err error) {
	fmt.Printf("%s: BatchCommit\n", d.Name)

	return d.child.Commit()
}

func (d *LogDatastore) Close() error {
	fmt.Printf("%s: Close\n", d.Name)

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

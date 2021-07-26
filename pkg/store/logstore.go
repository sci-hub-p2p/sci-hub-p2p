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
	ds "github.com/ipfs/go-datastore"
	dsq "github.com/ipfs/go-datastore/query"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"sci_hub_p2p/pkg/logger"
)

var _ ds.Datastore = (*LogDatastore)(nil)

// LogDatastore logs all accesses through the store.
type LogDatastore struct {
	logger *zap.Logger
	child  ds.Datastore
	Name   string
}

// Shim is a store which has a child.
type Shim interface {
	ds.Datastore
	Children() []ds.Datastore
}

// NewLogDatastore constructs a fmt store.
func NewLogDatastore(ds ds.Datastore, name string) *LogDatastore {
	if name == "" {
		name = "LogDatastore"
	}

	return &LogDatastore{Name: name, child: ds, logger: logger.WithLogger(name)}
}

// Children implements Shim.
func (d *LogDatastore) Children() []ds.Datastore {
	return []ds.Datastore{d.child}
}

// Put implements Datastore.Put.
func (d *LogDatastore) Put(key ds.Key, value []byte) (err error) {
	d.logger.Debug("Put", zap.String("key", key.String()))

	return d.child.Put(key, value)
}

// Sync implements Datastore.Sync.
func (d *LogDatastore) Sync(prefix ds.Key) error {
	d.logger.Debug("Sync", zap.String("prefix", prefix.String()))

	return d.child.Sync(prefix)
}

// Get implements Datastore.Get.
func (d *LogDatastore) Get(key ds.Key) (value []byte, err error) {
	d.logger.Debug("Get", zap.String("key", key.String()))

	value, err = d.child.Get(key)

	if errors.Is(err, ds.ErrNotFound) {
		d.logger.Debug("Get missing block", zap.String("key", key.String()))
	}

	return
}

// Has implements Datastore.Has.
func (d *LogDatastore) Has(key ds.Key) (exists bool, err error) {
	exists, err = d.child.Has(key)
	if err != nil {
		d.logger.Error("Has", zap.String("key", key.String()), zap.Bool("return", exists), zap.Error(err))
	} else {
		d.logger.Debug("Has", zap.String("key", key.String()), zap.Bool("return", exists))
	}

	return
}

// GetSize implements Datastore.GetSize.
func (d *LogDatastore) GetSize(key ds.Key) (size int, err error) {
	d.logger.Debug("GetSize", zap.String("key", key.String()))

	size, err = d.child.GetSize(key)

	if err != nil {
		if !errors.Is(err, ds.ErrNotFound) {
			d.logger.Error("GetSize", zap.String("key", key.String()), zap.Error(err))
		}
	}

	return
}

// Delete implements Datastore.Delete.
func (d *LogDatastore) Delete(key ds.Key) (err error) {
	d.logWithKey(key).Debug("Delete")

	return d.child.Delete(key)
}

// DiskUsage implements the PersistentDatastore interface.
func (d *LogDatastore) DiskUsage() (uint64, error) {
	d.logger.Debug("DiskUsage")

	return ds.DiskUsage(d.child)
}

// Query implements Datastore.Query.
func (d *LogDatastore) Query(q dsq.Query) (dsq.Results, error) {
	d.logger.Debug("Query", zap.String("prefix", q.Prefix), zap.Bool("keysOnly", q.KeysOnly))

	return d.child.Query(q)
}

// LogBatch logs all accesses through the batch.
type LogBatch struct {
	logger *zap.Logger
	child  ds.Batch
	Name   string
}

func (d *LogDatastore) Batch() (ds.Batch, error) {
	d.logger.Debug("Batch")
	if bds, ok := d.child.(ds.Batching); ok {
		b, err := bds.Batch()

		if err != nil {
			return nil, err
		}

		return &LogBatch{
			Name:   d.Name,
			child:  b,
			logger: d.logger.Named("LogBatch"),
		}, nil
	}

	return nil, ds.ErrBatchUnsupported
}

// Put implements Batch.Put.
func (d *LogBatch) Put(key ds.Key, value []byte) (err error) {
	d.logger.Debug("BatchPut", zap.String("key", key.String()))

	return d.child.Put(key, value)
}

// Delete implements Batch.Delete.
func (d *LogBatch) Delete(key ds.Key) (err error) {
	d.logger.Debug("BatchDelete", zap.String("key", key.String()))

	return d.child.Delete(key)
}

// Commit implements Batch.Commit.
func (d *LogBatch) Commit() (err error) {
	d.logger.Debug("BatchCommit")

	return d.child.Commit()
}

func (d *LogDatastore) Close() error {
	d.logger.Debug("Close")

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

func (d *LogDatastore) logWithKey(key ds.Key) *zap.Logger {
	return d.logger.With(zap.String("key", key.String()))
}

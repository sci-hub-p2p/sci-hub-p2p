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

package persist

import (
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/internal/torrent"
	"sci_hub_p2p/pkg/consts"
	"sci_hub_p2p/pkg/indexes"
	"sci_hub_p2p/pkg/vars"
)

var ErrNotFound = errors.New("Not found in database")

func GetIndexRecordDB(iDB *bbolt.DB, doi []byte) (*indexes.Record, error) {
	var r *indexes.Record

	err := iDB.View(func(tx *bbolt.Tx) error {
		if v := tx.Bucket(consts.IndexBucketName()).Get(doi); v != nil {
			r = indexes.LoadRecordV0(v)
		}

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to read from Database")
	}

	if r == nil {
		return nil, errors.Wrap(ErrNotFound, "failed to read doi in DB")
	}

	return r, nil
}

func GetIndexRecord(doi []byte) (*indexes.Record, error) {
	iDB, err := bbolt.Open(vars.IndexesBoltPath(), consts.DefaultFilePerm, bbolt.DefaultOptions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open indexes database")
	}
	defer iDB.Close()

	var r *indexes.Record

	err = iDB.View(func(tx *bbolt.Tx) error {
		if v := tx.Bucket(consts.IndexBucketName()).Get(doi); v != nil {
			r = indexes.LoadRecordV0(v)
		}

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to read from Database")
	}

	if r == nil {
		return nil, errors.Wrap(ErrNotFound, "failed to read doi in DB")
	}

	return r, nil
}

// GetTorrent accept a raw sha1 hash, return a parsed torrent.
func GetTorrent(hash []byte) (*torrent.Torrent, error) {
	tDB, err := bbolt.Open(vars.TorrentDBPath(), consts.DefaultFilePerm, bbolt.DefaultOptions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open torrent database")
	}
	defer tDB.Close()

	return GetTorrentDB(tDB, hash)
}

func GetTorrentDB(tDB *bbolt.DB, hash []byte) (*torrent.Torrent, error) {
	var raw []byte

	err := tDB.View(func(tx *bbolt.Tx) error {
		value := tx.Bucket(consts.TorrentBucket()).Get(hash)
		if value == nil {
			return errors.Wrap(ErrNotFound, "failed to find torrent in DB")
		}
		raw = make([]byte, len(value))
		copy(raw, value)

		return nil
	})
	if err != nil {
		return nil, err
	}

	t, err := torrent.ParseRaw(raw)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse torrent in database")
	}

	return t, nil
}

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
package dag

import (
	"archive/zip"
	"io"

	ipld "github.com/ipfs/go-ipld-format"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/pkg/storage"
)

func AddZip(db *bbolt.DB, abs string) error {
	return db.Batch(func(tx *bbolt.Tx) error {
		r, err := zip.OpenReader(abs)
		if err != nil {
			return errors.Wrap(err, "failed to open zip file")
		}
		defer r.Close()
		for _, f := range r.File {
			err := addZipContentFile(tx, abs, f)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func addZipContentFile(tx *bbolt.Tx, zipPath string, f *zip.File) error {
	offset, err := f.DataOffset()
	if err != nil {
		return errors.Wrap(err, "failed to get decompress file from zip")
	}

	r, err := f.Open()
	if err != nil {
		return errors.Wrap(err, "failed to read compressed file")
	}
	defer r.Close()
	_, err = addSingleFile(tx, zipPath, r, offset, f.CompressedSize64)

	return err
}

func addSingleFile(tx *bbolt.Tx, zipPath string, r io.Reader, offset int64, size uint64) (ipld.Node, error) {
	cf := wrapZipFile(r, zipPath, size)
	n, err := storage.Add(NewAdder(tx, offset), cf)

	return n, errors.Wrap(err, "failed to add generate DAG from reader")
}

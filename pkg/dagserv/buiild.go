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

package dagserv

import (
	"archive/zip"
	"io"

	"github.com/ipfs/go-cid"
	chunker "github.com/ipfs/go-ipfs-chunker"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-unixfs/importer/balanced"
	"github.com/ipfs/go-unixfs/importer/helpers"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

const DefaultChunkSize = 256 * 1024

func DefaultPrefix() cid.Prefix {
	return cid.Prefix{
		Version:  1,
		Codec:    cid.DagProtobuf,
		MhType:   multihash.Names["blake2b-256"],
		MhLength: -1,
	}
}

func AddZip(db *bbolt.DB, abs string) error {
	return errors.Wrap(db.Batch(func(tx *bbolt.Tx) error {
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
	}), "failed to add contents from files.")
}

func addZipContentFile(tx *bbolt.Tx, abs string, f *zip.File) error {
	offset, err := f.DataOffset()
	if err != nil {
		return errors.Wrap(err, "failed to get decompress file from zip")
	}

	r, err := f.Open()
	if err != nil {
		return errors.Wrap(err, "failed to read compressed file")
	}
	defer r.Close()
	_, err = addSingleFile(tx, abs, r, offset, f.CompressedSize64)

	return err
}

func addSingleFile(tx *bbolt.Tx, abs string, r io.Reader, offset int64, size uint64) (ipld.Node, error) {
	dbp := helpers.DagBuilderParams{
		Dagserv:    NewAdder(tx, offset),
		NoCopy:     true,
		RawLeaves:  true,
		Maxlinks:   helpers.DefaultLinksPerBlock,
		CidBuilder: DefaultPrefix(),
	}

	// NoCopy require a `FileInfo` on chunker
	cf := wrapZipFile(r, abs, size)

	chunk := chunker.NewSizeSplitter(cf, DefaultChunkSize)
	dbh, err := dbp.New(chunk)
	if err != nil {
		return nil, errors.Wrap(err, "can't create dag builder from chunker")
	}
	n, err := balanced.Layout(dbh)
	if err != nil {
		return nil, errors.Wrapf(err, "can't layout all chunk")
	}

	return n, nil
}

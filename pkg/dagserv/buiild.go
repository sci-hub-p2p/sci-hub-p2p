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
	"io"
	"os"

	"github.com/ipfs/go-cid"
	chunker "github.com/ipfs/go-ipfs-chunker"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-unixfs/importer/balanced"
	"github.com/ipfs/go-unixfs/importer/helpers"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

func Add(db *bbolt.DB, r io.Reader, abs string, size, baseOffset int64) (ipld.Node, error) {
	dbp := helpers.DagBuilderParams{
		Dagserv:    New(db, baseOffset),
		NoCopy:     true,
		RawLeaves:  true,
		Maxlinks:   helpers.DefaultLinksPerBlock,
		CidBuilder: DefaultPrefix(),
	}
	// NoCopy require a `FileInfo` on chunker
	f := CompressedFile{
		reader:             r,
		zipPath:            abs,
		compressedFilePath: "path/in/zip/article.pdf",
		size:               uint64(size),
	}

	chunk, err := chunker.FromString(f, "default")
	if err != nil {
		return nil, errors.Wrapf(err, "can't create default chunker")
	}
	dbh, err := dbp.New(chunk)
	if err != nil {
		return nil, errors.Wrap(err, "can't create dag builder from chunker")
	}
	n, err := balanced.Layout(dbh)
	if err != nil {
		return nil, errors.Wrapf(err, "can't layout all chunk")
	}

	return n, errors.Wrap(db.Sync(), "failed to flush data to disk")
}

func DefaultPrefix() cid.Prefix {
	return cid.Prefix{
		Version:  1,
		Codec:    cid.DagProtobuf,
		MhType:   multihash.Names["blake2b-256"],
		MhLength: -1,
	}
}

func AddFile(db *bbolt.DB, abs string) (ipld.Node, error) {
	s, err := os.Stat(abs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get file stat")
	}
	r, err := os.Open(abs)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file %s", abs)
	}
	defer r.Close()

	size := uint64(s.Size())
	dbp := helpers.DagBuilderParams{
		Dagserv:    New(db, 0),
		NoCopy:     true,
		RawLeaves:  true,
		Maxlinks:   helpers.DefaultLinksPerBlock,
		CidBuilder: DefaultPrefix(),
	}
	// NoCopy require a `FileInfo` on chunker
	f := CompressedFile{
		reader:             r,
		zipPath:            abs,
		compressedFilePath: "path/in/zip/article.pdf",
		size:               size,
	}

	chunk, err := chunker.FromString(f, "default")
	if err != nil {
		return nil, errors.Wrapf(err, "can't create default chunker")
	}
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

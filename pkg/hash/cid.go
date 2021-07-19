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

package hash

import (
	"io"
	"strings"

	"github.com/ipfs/go-cid"
	chunker "github.com/ipfs/go-ipfs-chunker"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-unixfs/importer/balanced"
	"github.com/ipfs/go-unixfs/importer/helpers"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"

	"sci_hub_p2p/pkg/dagServ"
)

func Black2dBalancedSized256K(r io.Reader) ([]byte, error) {
	var c, err = Cid(r)
	if err != nil {
		return nil, errors.Wrap(err, "can't generate cid")
	}

	return c.Hash(), nil
}

func Cid(r io.Reader) (cid.Cid, error) {
	var n, err = addFile(r, "blake2b-256", "size-262144", 1, true)
	if err != nil {
		return cid.Cid{}, errors.Wrap(err, "can't generate cid")
	}

	return n.Cid(), nil
}

var ErrMissingHashFunc = errors.New("missing hash function")

func addFile(
	r io.Reader,
	hashFun string,
	chunkMethod string,
	version uint64,
	rawLeaves bool,
) (ipld.Node, error) {
	hashFunCode, ok := multihash.Names[strings.ToLower(hashFun)]
	if !ok {
		return nil, errors.Wrapf(ErrMissingHashFunc, "unrecognized hash with %s", hashFun)
	}

	prefix := cid.Prefix{
		Version:  version,
		Codec:    cid.DagProtobuf,
		MhType:   hashFunCode,
		MhLength: -1,
	}

	dbp := helpers.DagBuilderParams{
		Dagserv:    dagServ.NewMemory(),
		RawLeaves:  true,
		Maxlinks:   helpers.DefaultLinksPerBlock,
		CidBuilder: &prefix,
		NoCopy:     false,
	}

	chunk, err := chunker.FromString(r, chunkMethod)
	if err != nil {
		return nil, errors.Wrapf(err, "can't create chunker %s", chunkMethod)
	}
	dbh, err := dbp.New(chunk)
	if err != nil {
		return nil, errors.Wrap(err, "can't create DAG builder from chunker")
	}

	n, err := balanced.Layout(dbh)
	if err != nil {
		return nil, errors.Wrapf(err, "can't layout all chunk")
	}

	return n, nil
}

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
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/ipfs/go-cid"
	chunker "github.com/ipfs/go-ipfs-chunker"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-unixfs/importer/balanced"
	"github.com/ipfs/go-unixfs/importer/helpers"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/pkg/constants"
)

func Build(raw []byte, abs string, baseOffset uint64) (ipld.Node, error) {
	prefix := cid.Prefix{
		Version:  1,
		Codec:    cid.DagProtobuf,
		MhType:   multihash.Names["blake2b-256"],
		MhLength: -1,
	}
	dbPath := "../../test.bolt"
	db, err := bbolt.Open(dbPath, constants.DefaultFilePerm, &bbolt.Options{
		FreelistType: bbolt.FreelistMapType,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "can't open database %s", dbPath)
	}
	defer db.Close()

	dagServ := ZipArchive{
		m:          &sync.Mutex{},
		db:         db,
		baseOffset: baseOffset,
	}

	dbp := helpers.DagBuilderParams{
		Dagserv:    dagServ,
		NoCopy:     true,
		RawLeaves:  true,
		Maxlinks:   helpers.DefaultLinksPerBlock,
		CidBuilder: &prefix,
	}
	// NoCopy require a `FileInfo` on chunker
	f := CompressedFile{
		reader:             bytes.NewReader(raw),
		zipPath:            abs,
		compressedFilePath: "path/in/zip/article.pdf",
		size:               int64(len(raw)),
	}

	chunk, err := chunker.FromString(f, "default")
	if err != nil {
		return nil, errors.Wrapf(err, "can't create default chunker")
	}
	dbh, err := dbp.New(chunk)
	if err != nil {
		return nil, errors.Wrap(err, "can't create dag builder from chunker")
	}
	fmt.Println("start layout")
	n, err := balanced.Layout(dbh)
	if err != nil {
		return nil, errors.Wrapf(err, "can't layout all chunk")
	}

	fmt.Println("try to get node", n.Cid())
	node, err := dagServ.Get(context.TODO(), n.Cid())
	if err != nil {
		return nil, err
	}
	fmt.Println(node.Cid(), Sha256SumHex(node.RawData()))
	fmt.Println(n.Cid(), Sha256SumHex(n.RawData()))

	return n, nil
}

func Sha256SumHex(b []byte) string {
	h := sha256.New()
	_, _ = h.Write(b)
	sum := h.Sum(nil)

	return hex.EncodeToString(sum)
}

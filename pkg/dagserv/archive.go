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
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/ipfs/go-cid"
	chunker "github.com/ipfs/go-ipfs-chunker"
	posinfo "github.com/ipfs/go-ipfs-posinfo"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-unixfs"
	"github.com/ipfs/go-unixfs/importer/helpers"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/pkg/constants"
	"sci_hub_p2p/pkg/logger"
)

func NewZip() ZipArchive {
	return ZipArchive{
		M: make(map[string]ipld.Node),
		m: &sync.Mutex{},
	}
}

var _ ipld.DAGService = ZipArchive{}

type ZipArchive struct {
	M          map[string]ipld.Node
	m          *sync.Mutex
	db         *bbolt.DB
	raw        []byte // raw content, determine block offset
	baseOffset uint64
}

func (d ZipArchive) Get(ctx context.Context, cid cid.Cid) (ipld.Node, error) {
	d.m.Lock()
	defer d.m.Unlock()
	i, ok := d.M[cid.String()]
	if !ok {
		return nil, ErrNotFound
	}

	return i, nil
}

func (d ZipArchive) GetMany(ctx context.Context, cids []cid.Cid) <-chan *ipld.NodeOption {
	var c = make(chan *ipld.NodeOption)
	go func() {
		for _, cid := range cids {
			i, err := d.Get(ctx, cid)
			c <- &ipld.NodeOption{Node: i, Err: err}
		}
	}()

	return c
}

func showFirst32(b []byte) {
	l := len(b)
	if l == 0 {
		return
	}

	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	if l > 32 {
		fmt.Print(hex.Dump(b[:32]))
	} else {
		fmt.Print(hex.Dump(b))
	}
	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
}

func (d ZipArchive) Add(ctx context.Context, node ipld.Node) error {
	d.m.Lock()
	defer d.m.Unlock()

	stat, _ := node.Stat()

	fmt.Println(node.Cid().String(), stat)

	v, ok := node.(*merkledag.ProtoNode)
	if ok {
		fmt.Println("is ProtoNode")
		n, err := unixfs.FSNodeFromBytes(v.Data())
		if err != nil {
			logger.Error(err)
		} else {
			// fmt.Println(n)
			if n.Data() != nil {
				fmt.Println("fsnode with data")
			} else {
				fmt.Println("n without data, should save pure node data")
			}
		}
	} else {
		fmt.Println("not ProtoNode")
		// is pure data node
		if v, ok := node.(*merkledag.RawNode); ok {
			fmt.Println("Node: merkledag.RawNode")
			fmt.Println(v.Size())
		}
		if v, ok := node.(*posinfo.FilestoreNode); ok {
			fmt.Println("Node: posinfo.FilestoreNode")
			fmt.Println(v.PosInfo.FullPath)
			blockOffsetOfZip := v.PosInfo.Offset + d.baseOffset
			length, _ := v.Size()
			fmt.Println("this block is", blockOffsetOfZip, length)
		}
	}

	fmt.Println()
	d.M[node.Cid().String()] = node

	return nil
}

func (d ZipArchive) AddMany(ctx context.Context, nodes []ipld.Node) error {
	for _, node := range nodes {
		_ = d.Add(ctx, node)
	}

	return nil
}

func (d ZipArchive) Remove(ctx context.Context, cid cid.Cid) error {
	d.m.Lock()
	defer d.m.Unlock()
	delete(d.M, cid.String())

	return nil
}

func (d ZipArchive) RemoveMany(ctx context.Context, cids []cid.Cid) error {
	for _, c := range cids {
		_ = d.Remove(ctx, c)
	}

	return nil
}

func Build(raw []byte, baseOffset uint64) (ipld.Node, error) {
	prefix := cid.Prefix{
		Version:  0,
		Codec:    cid.DagProtobuf,
		MhType:   multihash.SHA2_256,
		MhLength: -1,
	}
	db, err := bbolt.Open("../../test.bolt", constants.DefaultFilePerm, &bbolt.Options{
		Timeout:         1,
		FreelistType:    bbolt.FreelistMapType,
		InitialMmapSize: 60 * 1024 * 1024,
	})
	if err != nil {
		return nil, err
	}
	defer db.Close()
	dbp := helpers.DagBuilderParams{
		Dagserv: ZipArchive{
			M:          make(map[string]ipld.Node),
			m:          &sync.Mutex{},
			db:         db,
			raw:        raw,
			baseOffset: baseOffset,
		},
		NoCopy:    true,
		RawLeaves: true,
		// TODO: NoCopy require a `FileInfo` on chunker
		Maxlinks:   helpers.DefaultLinksPerBlock,
		CidBuilder: &prefix,
	}
	f := CompressedFile{
		reader:             bytes.NewReader(raw),
		zipPath:            "path/to/archive.zip",
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
	n, err := BalanceLayout(dbh)
	if err != nil {
		return nil, errors.Wrapf(err, "can't layout all chunk")
	}

	return n, nil
}

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
	"fmt"
	"io"
	"os"

	blocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	posinfo "github.com/ipfs/go-ipfs-posinfo"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	pb "github.com/ipfs/go-merkledag/pb"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

func ReadFileStoreNode(b *bbolt.Bucket, c cid.Cid) (ipld.Node, error) {
	var v = &Record{}
	data := b.Get(c.Bytes())
	if data == nil {
		return nil, ErrNotFound
	}

	err := proto.Unmarshal(data, v)
	if err != nil {
		return nil, errors.Wrap(err, "can't marshal persist.Record to binary")
	}
	f, err := os.Open(v.Filename)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open file %s", v.Filename)
	}
	defer f.Close()

	_, err = f.Seek(int64(v.Offset), io.SeekStart)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to seek file %s", v.Filename)
	}

	var p = make([]byte, v.Length)

	_, err = io.ReadFull(f, p)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read file %s", v.Filename)
	}

	block, err := blocks.NewBlockWithCid(p, c)

	return &posinfo.FilestoreNode{Node: &merkledag.RawNode{Block: block}}, errors.Wrap(err, "failed to create block")
}

func SaveFileStoreMeta(b *bbolt.Bucket, c cid.Cid, name string, offset, size uint64) error {
	v := &Record{
		Offset:   offset,
		Length:   size,
		Filename: name,
	}
	raw, err := proto.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "can't marshal persist.Record to binary")
	}

	return errors.Wrap(b.Put(c.Bytes(), raw), "failed to save data to database")
}

func SaveProtoNode(b *bbolt.Bucket, c cid.Cid, n *merkledag.ProtoNode) error {
	return errors.Wrap(b.Put(c.Bytes(), n.RawData()), "failed to save data to database")
}

func ReadProtoNode(b *bbolt.Bucket, c cid.Cid) (ipld.Node, error) {
	data := b.Get(c.Bytes())
	if data == nil {
		return nil, ErrNotFound
	}
	v, err := unmarshal(data, c)

	return v, errors.Wrap(err, "failed to unmarshal data to `merkledag.ProtoNode`")
}

// from https://github.com/ipfs/go-merkledag/blob/v0.3.2/coding.go#L25-L46
func unmarshal(encoded []byte, c cid.Cid) (*merkledag.ProtoNode, error) {
	var n = &merkledag.ProtoNode{}

	n.SetCidBuilder(cid.V1Builder{
		Codec:    c.Prefix().Codec,
		MhType:   c.Prefix().MhType,
		MhLength: -1,
	})

	var pbn pb.PBNode

	if err := pbn.Unmarshal(encoded); err != nil {
		return nil, errors.Wrap(err, "unmarshal failed")
	}

	pbnl := pbn.GetLinks()
	for i, l := range pbnl {
		c, err := cid.Cast(l.GetHash())
		if err != nil {
			return nil, fmt.Errorf("link hash #%d is not valid multihash. %w", i, err)
		}
		err = n.AddRawLink(l.GetName(), &ipld.Link{
			Name: l.GetName(),
			Size: l.GetTsize(),
			Cid:  c,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to add link to node")
		}
	}

	n.SetData(pbn.GetData())

	return n, nil
}

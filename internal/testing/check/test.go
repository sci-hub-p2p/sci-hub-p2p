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

// nolint
package main

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"

	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/pb"
	"sci_hub_p2p/pkg/variable"
)

func main() {
	LoadTestData()
}

func LoadTestData() {
	db, err := bbolt.Open("./test.bolt", 0644, bbolt.DefaultOptions)
	if err != nil {
		logger.Fatal("", logger.Err(err))
	}
	defer db.Close()

	err = db.View(func(tx *bbolt.Tx) error {
		nb := tx.Bucket(variable.NodeBucketName())
		bb := tx.Bucket(variable.BlockBucketName())
		return bb.ForEach(func(k, v []byte) error {
			fmt.Println()
			fmt.Println("checking", hex.EncodeToString(k))
			mh, err := multihash.Decode(k)
			if err != nil {
				return errors.Wrap(err, "failed to decode multihash")
			}

			fmt.Println("multi hash:", mh.Name, mh.Length, mh.Code, len(mh.Digest))

			var r = &pb.Block{}
			err = proto.Unmarshal(v, r)
			if err != nil {
				return errors.Wrap(err, "failed to decode block Record from database raw value")
			}
			fmt.Println(r.Type)
			c, err := cid.Parse(r.CID)
			if err != nil {
				if bytes.Equal(r.CID, k) {
					fmt.Println("!! why block key and it's record's CID are equal? !!")
				}
				fmt.Println(hex.EncodeToString(r.CID))
				return err
			}
			if !bytes.Equal(c.Bytes(), r.CID) {
				fmt.Println(hex.EncodeToString(c.Bytes()))
				fmt.Println(hex.EncodeToString(r.CID))
				return fmt.Errorf("cid not equal, %s", hex.EncodeToString(k))
			}
			p := nb.Get(r.CID)
			if p == nil {
				return errors.Wrap(ds.ErrNotFound, "can't read proto node from node bucket")
			}
			return nil
		})
	})

	if err != nil {
		logger.Fatal("", logger.Err(err))
	}

}

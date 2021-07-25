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

package persist

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/vars"
)

func CopyBucket(src, dst *bbolt.DB, name []byte) error {
	return dst.Batch(func(dstTx *bbolt.Tx) error {
		dstBucket, err := dstTx.CreateBucketIfNotExists(name)
		if err != nil {
			return errors.Wrap(err, "failed to create bucket in dst DB")
		}

		return src.View(func(srcTx *bbolt.Tx) error {
			srcBucket := srcTx.Bucket(name)

			return srcBucket.ForEach(func(k, v []byte) error {
				if bytes.Equal(name, vars.NodeBucketName()) {
					_, err := cid.Parse(k)
					if err != nil {
						logger.Error(err)
						fmt.Println(hex.Dump(k))
					}
				}

				return dstBucket.Put(k, v)
			})
		})
	})
}

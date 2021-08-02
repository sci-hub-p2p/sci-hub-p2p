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

package utils

import (
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
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
				return dstBucket.Put(k, v)
			})
		})
	})
}

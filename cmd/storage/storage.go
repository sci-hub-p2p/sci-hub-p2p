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

package storage

import (
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/pkg/consts"
	"sci_hub_p2p/pkg/indexes"
	"sci_hub_p2p/pkg/vars"
)

func GetIndexRecord(doi []byte) (*indexes.Record, error) {
	iDB, err := bbolt.Open(vars.IndexesBoltPath(), consts.DefaultFilePerm, bbolt.DefaultOptions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open indexes database")
	}
	defer iDB.Close()

	var r *indexes.Record
	err = iDB.View(func(tx *bbolt.Tx) error {
		if v := tx.Bucket(consts.IndexBucketName()).Get(doi); v != nil {
			r = indexes.LoadRecordV0(v)
		}

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to read from Database")
	}
	if r == nil {
		return nil, errors.Wrap(err, "failed to read doi in DB")
	}

	return r, nil
}

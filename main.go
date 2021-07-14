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

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"go.etcd.io/bbolt"

	"sci_hub_p2p/cmd"
	"sci_hub_p2p/pkg/indexes"
)

func main() {
	cmd.Execute()
}
func init1() {
	var defaultFileMode os.FileMode = 0644
	out := filepath.Join(`C:\Users\Trim21\proj\sci_hub_p2p\out\8438d7c356229789a1da513e724405f588b25b61.indexes`)
	db, err := bbolt.Open(out, defaultFileMode, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("can't open %s to write indexes: %s", out, err)
	}
	defer db.Close()

	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("paper-v0"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("DOI=%s,\tvalue=%+v", k, indexes.LoadRecord(v))
		}

		return nil
	})
}

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
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"reflect"

	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	ufsio "github.com/ipfs/go-unixfs/io"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/cmd"
	"sci_hub_p2p/pkg/dagserv"
	"sci_hub_p2p/pkg/hash"
)

func main() {
	var err = try()
	if err != nil {
		panic(err)
	}

	// expectedKey := datastore.RawKey("/UDSAEIE5SHNQHV5U6G4HOPOWINVOOCVWBBRYVTZYKB3DZGT6QZGOWTSFE4")

	cmd.Execute()
}
func prepare() error {
	db, err := bbolt.Open("./test.bolt", 0600, &bbolt.Options{})
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = dagserv.AddFile(db, "./testdata/big_file.bin")
	if err != nil {
		return err
	}
	return nil
}
func try() error {
	err := prepare()
	if err != nil {
		return err
	}

	db, err := bbolt.Open("./test.bolt", 0600, &bbolt.Options{ReadOnly: true})
	if err != nil {
		return err
	}
	defer db.Close()

	dag := dagserv.New(db, 0)
	err = p2(dag)

	if err != nil {
		return err
	}

	c, err := cid.Parse("bafykbzaceavd6aaauynuqgkkrg6lapmno5crbsyinmp3um5sn3daztzsghvl2")

	if err != nil {
		return err
	}
	n, err := dag.Get(context.TODO(), c)
	if err != nil {
		return err
	}

	fmt.Println(reflect.TypeOf(n))
	fmt.Println("start to read from dag server")
	reader, err := ufsio.NewDagReader(context.TODO(), n, dag)
	if err != nil {
		return errors.Wrap(err, "can't create DAG reader")
	}

	raw, err := io.ReadAll(reader)

	if err != nil {
		return err
	}
	fmt.Println(hex.EncodeToString(hash.Sha1SumBytes(raw)))

	return nil
}

func p2(dag ipld.DAGService) error {
	c, err := cid.Parse("bafk2bzacebcktc35bpqrurwqk5eadvgrdoswyzlpjjwol57sqgzqmj3himwr6")
	if err != nil {
		return err
	}

	n, err := dag.Get(context.TODO(), c)
	if err != nil {
		return err
	}
	if v, ok := n.(*merkledag.RawNode); ok {
		fmt.Println(v.RawData())
	}
	fmt.Println(n.Cid())
	return nil
}

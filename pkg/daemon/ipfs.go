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

package daemon

// This example launches an IPFS-Lite peer and fetches a hello-world
// hash from the IPFS network.

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/ipfs/go-datastore"
	config "github.com/ipfs/go-ipfs-config"
	log2 "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-core/routing"
	discovery "github.com/libp2p/go-libp2p-discovery"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/internal/ipfslite"
	"sci_hub_p2p/pkg/constants"
	"sci_hub_p2p/pkg/dagserv"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/store"
	"sci_hub_p2p/pkg/vars"
)

func Start() error {
	setupLogger()
	ctx := context.Background()
	db, err := bbolt.Open("./test.bolt", constants.DefaultFilePerm, bbolt.DefaultOptions)
	if err != nil {
		return errors.Wrap(err, "failed to open database")
	}
	defer db.Close()

	if err = dagserv.InitDB(db); err != nil {
		return errors.Wrap(err, "can't init database")
	}
	rawDS := store.NewMapDatastore(db)
	ds := store.NewLogDatastore(rawDS, "debug")
	startHTTPServer(ds)

	privKey, err := genKey()
	if err != nil {
		return err
	}
	logger.Info("finish load key")
	listen, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/4005")
	pnetKey, err := pnetKey()
	if err != nil {
		return err
	}

	h, dht, err := ipfslite.SetupLibp2p(
		ctx,
		privKey,
		pnetKey,
		[]multiaddr.Multiaddr{listen},
		ds,
		ipfslite.Libp2pOptionsExtra...,
	)
	if err != nil {
		return err
	}

	lite, err := ipfslite.New(ctx, ds, h, dht, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create new peer")
	}

	logger.Info("ipfs peer started")
	fmt.Printf("/ip4/127.0.0.1/tcp/4005/p2p/%s\n", h.ID())

	lite.Bootstrap(ipfslite.DefaultBootstrapPeers())
	p, err := config.ParseBootstrapPeers([]string{
		"/ip4/127.0.0.1/tcp/4001/p2p/12D3KooWCsZQqmqi42PXHKmAXAHvevp8HXnfViWg4Txp5ayJoqSq",
	})
	if err != nil {
		panic(err)
	}

	lite.Bootstrap(p)
	// lite.Bootstrap(ipfslite.DefaultBootstrapPeers())
	// db.View(func(tx *bbolt.Tx) error {
	// 	return tx.Bucket(vars.BlockBucketName()).ForEach(func(k, v []byte) error {
	// 		var r = &dagserv.Block{}
	// 		err := proto.Unmarshal(v, r)
	// 		if err != nil {
	// 			panic(err)
	// 		}
	//
	// 		c, err := cid.Parse(r.CID)
	// 		if err != nil {
	// 			panic(err)
	// 		}
	//
	// 		if rand.Intn(10000) < 2 {
	// 			if r.Type == dagserv.BlockType_proto {
	// 				fmt.Println(c)
	// 				fmt.Println(store.MultiHashToKey(k))
	// 				fmt.Println(dshelp.MultihashToDsKey(c.Hash()))
	// 			}
	// 		}
	// 		return nil
	// 	})
	// })
	for {
		bootIPFSDaemon(context.Background(), dht)
		time.Sleep(time.Second * 20)
	}
}

func setupLogger() {
	err := log2.SetLogLevel("*", "error")
	if err != nil {
		panic(err)
	}
}

func startHTTPServer(d datastore.Datastore) {
	var m = mux.NewRouter()
	m.HandleFunc("/GET/blocks/{key}", func(w http.ResponseWriter, r *http.Request) {
		var v = mux.Vars(r)
		node, err := d.Get(datastore.NewKey("/blocks/" + v["key"]))
		if errors.Is(err, datastore.ErrNotFound) {
			fmt.Fprintf(w, "missing block %s", err)
			return
		}
		fmt.Fprintf(w, "length %d\n", len(node))
		fmt.Fprintf(w, "hex:\n")
		hex.Dumper(w).Write(node)

	})

	m.HandleFunc("/GETSIZE/blocks/{key}", func(w http.ResponseWriter, r *http.Request) {
		var v = mux.Vars(r)
		l, err := d.GetSize(datastore.NewKey("/blocks/" + v["key"]))
		if errors.Is(err, datastore.ErrNotFound) {
			fmt.Fprintf(w, "missing block %s", err)
			return
		}
		fmt.Fprintf(w, "length %d", l)
	})

	go http.ListenAndServe(":2333", m)
}

func bootIPFSDaemon(ctx context.Context, dht routing.ContentRouting) {
	nodeName := fmt.Sprintf("sci-hub-p2p %s", vars.Ref)
	logger.Debug("Announcing ourselves...")
	routingDiscovery := discovery.NewRoutingDiscovery(dht)
	discovery.Advertise(ctx, routingDiscovery, nodeName)
	logger.Debug("Successfully announced!")
}

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
	"strconv"
	"time"

	"github.com/gorilla/mux"
	ds "github.com/ipfs/go-datastore"
	log2 "github.com/ipfs/go-log"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"

	"sci_hub_p2p/cmd/flag"
	"sci_hub_p2p/internal/ipfslite"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/store"
)

const interval = time.Second * 20

func Start(db *bbolt.DB, port int) error {
	setupIPFSLogger()
	ctx := context.Background()
	var datastore ds.Batching = store.NewArchiveFallbackDatastore(db)
	if flag.Debug {
		startHTTPServer(datastore)
		datastore = store.NewLogDatastore(datastore, "LogDatastore")
	}
	privKey, err := genKey()
	if err != nil {
		return err
	}
	logger.Info("finish load key")
	listen, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/" + strconv.Itoa(port))
	pnetKey, err := pnetKey()
	if err != nil {
		return err
	}

	h, dht, err := ipfslite.SetupLibp2p(
		ctx,
		privKey,
		pnetKey,
		[]multiaddr.Multiaddr{listen},
		datastore,
		ipfslite.Libp2pOptionsExtra...,
	)

	if err != nil {
		return errors.Wrap(err, "failed to start libp2p")
	}

	lite, err := ipfslite.New(ctx, datastore, h, dht, &ipfslite.Config{ReprovideInterval: time.Hour})
	if err != nil {
		return errors.Wrap(err, "failed to create new peer")
	}

	logger.WithLogger("ipfs").Info("peer started")
	fmt.Printf("your peer address is /ip4/127.0.0.1/tcp/%d/p2p/%s\n", port, h.ID())

	lite.Bootstrap(ipfslite.DefaultBootstrapPeers())

	for {
		time.Sleep(interval)
	}
}

func setupIPFSLogger() {
	err := log2.SetLogLevel("*", "error")
	if err != nil {
		panic(err)
	}
}

func startHTTPServer(d ds.Datastore) {
	var m = mux.NewRouter()
	m.HandleFunc("/GET/blocks/{key}", func(w http.ResponseWriter, r *http.Request) {
		var v = mux.Vars(r)
		node, err := d.Get(ds.NewKey("/blocks/" + v["key"]))
		if errors.Is(err, ds.ErrNotFound) {
			fmt.Fprintf(w, "missing block %s", err)

			return
		}
		fmt.Fprintf(w, "length %d\n", len(node))
		fmt.Fprintf(w, "hex:\n")
		d := hex.Dumper(w)
		_, _ = d.Write(node)
		d.Close()
	})

	m.HandleFunc("/GETSIZE/blocks/{key}", func(w http.ResponseWriter, r *http.Request) {
		var v = mux.Vars(r)
		l, err := d.GetSize(ds.NewKey("/blocks/" + v["key"]))
		if errors.Is(err, ds.ErrNotFound) {
			fmt.Fprintf(w, "missing block %s", err)

			return
		}
		fmt.Fprintf(w, "length %d", l)
	})

	go func() {
		err := http.ListenAndServe(":2333", m)
		if err != nil {
			logger.Error("failed to create debug server", zap.Error(err))
		} else {
			logger.Info("start debug http server in http://127.0.0.1:2333")
		}
	}()
}

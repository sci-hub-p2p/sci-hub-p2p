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
package daemon

// This example launches an IPFS-Lite peer and fetches a hello-world
// hash from the IPFS network.

import (
	"context"
	"fmt"
	"strconv"
	"time"

	ds "github.com/ipfs/go-datastore"
	log2 "github.com/ipfs/go-log"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

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
		ipfslite.DefaultLibp2pOptions()...,
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

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
	"time"

	ds "github.com/ipfs/go-datastore"
	log2 "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p"
	libp2pquic "github.com/libp2p/go-libp2p-quic-transport"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/cmd/flag"
	"sci_hub_p2p/internal/ipfslite"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/store"
)

const interval = time.Second * 20

func New(db *bbolt.DB, port int, cacheSize int64) (*ipfslite.Peer, error) {
	var ctx = context.Background()
	var datastore ds.Batching = store.NewArchiveFallbackDatastore(db, cacheSize)

	setupIPFSLogger()

	if flag.Debug {
		datastore = store.NewLogDatastore(datastore, "LogDatastore")
	}

	privKey, err := genKey()
	if err != nil {
		return nil, err
	}

	logger.Info("finish load key")

	listen, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))
	if err != nil {
		panic(err)
	}

	listenAddrs := []multiaddr.Multiaddr{listen}

	pnetKey, err := pnetKey()
	if err != nil {
		return nil, err
	}

	var options = ipfslite.DefaultLibp2pOptions()

	if pnetKey != nil {
		logger.Warn("you are using pnet key, disable quic support")
	} else {
		options = append(options, libp2p.Transport(libp2pquic.NewTransport))
		listen, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", port))
		if err != nil {
			panic(err)
		}
		listenAddrs = append(listenAddrs, listen)
	}

	h, dht, err := ipfslite.SetupLibp2p(ctx, privKey, pnetKey, listenAddrs, datastore, options...)

	// var p namesys.Resolver = dht

	if err != nil {
		return nil, errors.Wrap(err, "failed to start libp2p")
	}

	lite, err := ipfslite.New(ctx, datastore, h, dht, &ipfslite.Config{ReprovideInterval: time.Hour})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new peer")
	}

	logger.WithLogger("ipfs").Info("peer started")
	fmt.Printf("your peer address is /ip4/127.0.0.1/tcp/%d/p2p/%s\n", port, h.ID())
	lite.Bootstrap(ipfslite.DefaultBootstrapPeers())

	return lite, nil
}

func Start(db *bbolt.DB, port int, cacheSize int64) error {
	_, err := New(db, port, cacheSize)
	if err != nil {
		return errors.Wrap(err, "failed to create new peer")
	}

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

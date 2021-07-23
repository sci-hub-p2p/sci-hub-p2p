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
	"fmt"
	"time"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-cid"
	dssync "github.com/ipfs/go-datastore/sync"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
	"go.etcd.io/bbolt"

	// "github.com/ipfs/go-datastore"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/multiformats/go-multiaddr"

	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/store"
)

func Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := bbolt.Open("./test.bolt", 0600, bbolt.DefaultOptions)
	if err != nil {
		return err
	}
	defer db.Close()

	ds := dssync.MutexWrap(store.NewLogDatastore(store.NewMapDatastore(db), "debug"))
	privateKey, _, err := crypto.GenerateKeyPair(crypto.RSA, 4096)
	if err != nil {
		return err
	}

	listen, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/4005")

	h, dht, err := ipfslite.SetupLibp2p(
		ctx,
		privateKey,
		nil,
		[]multiaddr.Multiaddr{listen},
		ds,
		ipfslite.Libp2pOptionsExtra...,
	)

	if err != nil {
		return err
	}

	lite, err := ipfslite.New(ctx, ds, h, dht, nil)
	if err != nil {
		return err
	}

	lite.Bootstrap(ipfslite.DefaultBootstrapPeers())

	c, err := cid.Parse("bafykbzacecozdwyd262pdodxhxleg2xhbk3aqy4kz44fa5r4tj7imthljzcso")
	if err != nil {
		return err
	}

	fmt.Println(dshelp.MultihashToDsKey(c.Hash()))

	n, err := lite.GetFile(context.TODO(), c)
	if err != nil {
		return err
	}
	fmt.Println(n)

	logger.Info("listening")
	for {
		time.Sleep(10 * time.Second)
	}
	return nil
}

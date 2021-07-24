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
	"os"
	"time"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/pkg/constants"
	"sci_hub_p2p/pkg/dagserv"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/store"
	"sci_hub_p2p/pkg/variable"
)

const dhtConcurrency = 10

func Start() error {
	db, err := bbolt.Open("./test.bolt", constants.DefaultFilePerm, bbolt.DefaultOptions)
	if err != nil {
		return errors.Wrap(err, "failed to open database")
	}
	defer db.Close()

	if err = dagserv.InitDB(db); err != nil {
		return errors.Wrap(err, "can't init database")
	}
	rawDS := store.NewMapDatastore(db)
	mutexDS := dssync.MutexWrap(rawDS)
	ds := store.NewLogDatastore(mutexDS, "debug")
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

	h, dht, err := SetupLibp2p(
		context.TODO(),
		privKey,
		pnetKey,
		[]multiaddr.Multiaddr{listen},
		ds,
		ipfslite.Libp2pOptionsExtra...,
	)
	if err != nil {
		return err
	}

	lite, err := ipfslite.New(context.TODO(), ds, h, dht, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create new peer")
	}

	// lite.Bootstrap(ipfslite.DefaultBootstrapPeers())
	logger.Info("listening")
	fmt.Printf("/ip4/127.0.0.1/tcp/4005/p2p/%s\n", h.ID())
	ipfslite.NewInMemoryDatastore()
	count := 0
	_ = db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(variable.NodeBucketName()).ForEach(func(k, v []byte) error {
			c, err := cid.Parse(k)
			if err != nil {
				count++
				logger.Error(err)
				fmt.Println(hex.Dump(k))
				if count > 4 {
					os.Exit(1)
				}
			}
			_ = datastore.NewKey("/blocks/").Child(dshelp.MultihashToDsKey(c.Hash()))
			return nil
		})
		//
		// 	if c.ByteLen() != 0 {
		// 		b := tx.Bucket(variable.BlockBucketName())
		// 		v := b.Get(c.Hash())
		// 		if v == nil {
		// 			panic("should not be nil")
		// 		}
		// 	}
		//
		// 	return nil
	})

	for {
		bootIPFSDaemon(lite, h)
		time.Sleep(time.Second)
	}
}

// for local testing.
func bootIPFSDaemon(lite *ipfslite.Peer, h host.Host) {
	hostIpfs, err := multiaddr.NewMultiaddr(
		"/ip4/127.0.0.1/tcp/4001/p2p/12D3KooWF8AC8XXVGcQZXjoQUpSgKHZMv71Nj8iCo9GSGrq3UFPt")
	if err != nil {
		panic(err)
	}
	p, err := peer.AddrInfosFromP2pAddrs(hostIpfs)
	if err != nil {
		panic(err)
	}
	lite.Bootstrap(p)
	for _, info := range p {
		h.Connect(context.TODO(), info)
	}
}

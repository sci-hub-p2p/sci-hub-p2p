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
	"crypto/rsa"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	ipfslite "github.com/hsanjuan/ipfs-lite"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	dshelp "github.com/ipfs/go-ipfs-ds-help"
	"github.com/ipfs/go-ipns"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/pnet"
	"github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	dualdht "github.com/libp2p/go-libp2p-kad-dht/dual"
	record "github.com/libp2p/go-libp2p-record"
	"github.com/mr-tron/base58"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/pkg/constants"
	"sci_hub_p2p/pkg/dagserv"
	"sci_hub_p2p/pkg/hash"
	"sci_hub_p2p/pkg/key"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/store"
)

const dhtConcurrency = 10

func Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := bbolt.Open("./test.bolt", constants.DefaultFilePerm, bbolt.DefaultOptions)
	if err != nil {
		return errors.Wrap(err, "failed to open database")
	}
	defer db.Close()

	err = dagserv.InitDB(db)
	if err != nil {
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

	// lite.Bootstrap(ipfslite.DefaultBootstrapPeers())
	bootIPFSDaemon(lite)
	logger.Info("listening")
	fmt.Printf("/ip4/127.0.0.1/tcp/4005/p2p/%s\n", h.ID())

	for {
		mutexDS.RLock()
		store.List(rawDS)
		mutexDS.RUnlock()
		c, _ := cid.Parse("bafykbzaceavd6aaauynuqgkkrg6lapmno5crbsyinmp3um5sn3daztzsghvl2")
		k := fmt.Sprintf("/blocks%s", dshelp.MultihashToDsKey(c.Hash()))
		fmt.Println(k)
		mh, err := dshelp.DsKeyToMultihash(datastore.NewKey(strings.TrimPrefix(k, "/blocks")))
		if err != nil {
			logger.Error(err)
		} else {
			fmt.Println(base58.Encode(mh))
		}
		r, err := lite.GetFile(context.TODO(), c)
		if err != nil {
			logger.Error(err)
		}
		fmt.Println(hash.Sha256SumReader(r))

		time.Sleep(time.Second)
	}
}

func SetupLibp2p(
	ctx context.Context,
	hostKey crypto.PrivKey,
	secret pnet.PSK,
	listenAddrs []multiaddr.Multiaddr,
	ds datastore.Batching,
	opts ...libp2p.Option,
) (host.Host, *dualdht.DHT, error) {
	var ddht *dualdht.DHT
	var err error
	finalOpts := []libp2p.Option{
		libp2p.Identity(hostKey),
		libp2p.ListenAddrs(listenAddrs...),
		libp2p.PrivateNetwork(secret),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			ddht, err = newDHT(ctx, h, ds)

			return ddht, err
		}),
	}
	finalOpts = append(finalOpts, opts...)

	h, err := libp2p.New(
		ctx,
		finalOpts...,
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed init libp2p")
	}

	return h, ddht, nil
}

func newDHT(ctx context.Context, h host.Host, ds datastore.Batching) (*dualdht.DHT, error) {
	dhtOpts := []dualdht.Option{
		dualdht.DHTOption(dht.NamespacedValidator("pk", record.PublicKeyValidator{})),
		dualdht.DHTOption(dht.NamespacedValidator("ipns", ipns.Validator{KeyBook: h.Peerstore()})),
		dualdht.DHTOption(dht.Concurrency(dhtConcurrency)),
		dualdht.DHTOption(dht.Mode(dht.ModeAuto)),
	}
	if ds != nil {
		dhtOpts = append(dhtOpts, dualdht.DHTOption(dht.Datastore(ds)))
	}

	d, err := dualdht.New(ctx, h, dhtOpts...)

	return d, errors.Wrap(err, "failed to init dualdht")
}

func pnetKey() (pnet.PSK, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Wrap(err, "failed to detect homedir")
	}
	var keyPath = filepath.Join(home, ".ipfs/swarm.key")
	r, err := os.Open(keyPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to read pnet key %s", keyPath)
	}
	defer r.Close()
	logger.Info("using pnet key")
	k, err := pnet.DecodeV1PSK(r)

	return k, errors.Wrap(err, "failed to decode pnet KEY")
}

func genKey() (crypto.PrivKey, error) {
	const keyPath = "./out/private.key"
	var raw, err = os.ReadFile(keyPath)
	if errors.Is(err, os.ErrNotExist) {
		logger.Info("Generating New Rsa Key")
		priv, _, err := crypto.GenerateKeyPair(crypto.RSA, constants.PrivateKeyLength)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate RSA key")
		}
		stdKey, err := crypto.PrivKeyToStdKey(priv)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert libp2p key to std key")
		}
		v, ok := stdKey.(*rsa.PrivateKey)
		if !ok {
			panic("can't cast private key to *rsa.PrivateKey")
		}
		raw = key.ExportRsaPrivateKeyAsPem(v)
		err = os.WriteFile(keyPath, raw, constants.SecurityPerm)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to save key to file %s", keyPath)
		}

		return priv, nil
	}

	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}
	k, err := crypto.UnmarshalRsaPrivateKey(block.Bytes)

	return k, errors.Wrap(err, "filed to parse encode keyfile content")
}

func bootIPFSDaemon(lite *ipfslite.Peer) {
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
}

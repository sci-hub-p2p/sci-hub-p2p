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

package bootstrap

import (
	"context"
	"fmt"

	"github.com/asdine/storm/v3"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-path"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"

	"sci_hub_p2p/internal/ipfslite"
	"sci_hub_p2p/internal/ipns"
	"sci_hub_p2p/pkg/consts"
	"sci_hub_p2p/pkg/daemon"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/vars"
)

var Cmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstrap all needed data from P2P network",
	RunE: func(cmd *cobra.Command, args []string) error {
		// 2 thinks need to do:
		// download all indexes
		// download all torrents
		fmt.Println("indexes command")
		logger.Info("open database", zap.String("db", vars.IpfsDBPath()))
		db, err := bbolt.Open(vars.IpfsDBPath(), consts.DefaultFilePerm, bbolt.DefaultOptions)
		if err != nil {
			return errors.Wrap(err, "failed to open database")
		}
		defer db.Close()
		err = db.View(func(tx *bbolt.Tx) error {
			if tx.Bucket(consts.BlockBucketName()) == nil {
				return errors.New("database is empty")
			}
			if tx.Bucket(consts.NodeBucketName()) == nil {
				return errors.New("database is empty")
			}

			return nil
		})
		if err != nil {
			return err
		}

		peer, err := daemon.New(db, 4000, 256) // nolint
		if err != nil {
			return errors.Wrap(err, "failed to start ipfs node")
		}
		// this is a test address.
		err = bootstrap(peer, "/ipns/k51qzi5uqu5dgwgi53wn9vhjgq75ic57g0c8xqjtva0fj0ubgc286oacldv29s")

		return err
	},
}

type Added struct {
	InfoHash []byte
}

func bootstrap(p *ipfslite.Peer, name path.Path) error {
	r, err := ipns.ResolveIPNS(context.TODO(), p.NameResolver(), name)
	if err != nil {
		return errors.Wrap(err, "failed to resolve ipns")
	}

	c, err := cid.Parse(r.String())
	if err != nil {
		return errors.Wrapf(err, "failed to parse CID %s", r)
	}

	db, err := storm.Open("my.db")
	if err != nil {
		return errors.Wrap(err, "failed to open database")
	}
	defer db.Close()

	n, err := p.Get(context.TODO(), c)
	if err != nil {
		return errors.Wrap(err, "failed to read ipns content")
	}

	v, ok := n.(*merkledag.ProtoNode)
	if !ok {
		return fmt.Errorf("address doesn't point to sci-hub-p2p's indexes")
	}

	fmt.Println(v.Links())

	return nil
}

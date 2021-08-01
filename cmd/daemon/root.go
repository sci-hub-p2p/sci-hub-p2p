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

import (
	"fmt"
	"math"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"

	"sci_hub_p2p/internal/utils"
	"sci_hub_p2p/pkg/consts"
	"sci_hub_p2p/pkg/daemon"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/vars"
	"sci_hub_p2p/pkg/web"
)

var Cmd = &cobra.Command{
	Use: "daemon",
}

var httpAPICmd = &cobra.Command{
	Use:     "http",
	Short:   "start http server for http api and Web-UI",
	PreRunE: utils.EnsureDir(vars.GetAppBaseDir()),
	RunE: func(cmd *cobra.Command, args []string) error {
		return web.Start(port)
	},
}

var ipfsCmd = &cobra.Command{
	Use:     "ipfs",
	Short:   "start ipfs node",
	PreRunE: utils.EnsureDir(vars.GetAppBaseDir()),
	RunE: func(cmd *cobra.Command, args []string) error {
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

		fmt.Println(cacheSize * MB)
		fmt.Println(math.Log2(float64(cacheSize * MB)))

		return daemon.Start(db, port, cacheSize*MB)
	},
}
var port int
var cacheSize int64

const MB int64 = 1 << 20
const defaultDaemonPort = 4005
const defaultWebPort = 2333
const defaultCacheSize = 1 << 9

func init() {
	Cmd.AddCommand(ipfsCmd, httpAPICmd)
	ipfsCmd.Flags().IntVarP(&port, "port", "p", defaultDaemonPort, "IPFS peer default port")
	ipfsCmd.Flags().Int64Var(&cacheSize, "cache", defaultCacheSize, "memory cache size for disk in MB")

	httpAPICmd.Flags().IntVarP(&port, "port", "p", defaultWebPort, "IPFS peer default port")
}

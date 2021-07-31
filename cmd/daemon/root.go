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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"
	"go.uber.org/zap"

	"sci_hub_p2p/internal/utils"
	"sci_hub_p2p/pkg/consts"
	"sci_hub_p2p/pkg/daemon"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/vars"
)

var Cmd = &cobra.Command{
	Use: "daemon",
}

var startCmd = &cobra.Command{
	Use:     "start",
	Short:   "start daemon",
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

		return daemon.Start(db, port)
	},
}
var port int

const defaultDaemonPort = 4005

func init() {
	Cmd.AddCommand(startCmd)
	startCmd.Flags().IntVarP(&port, "port", "p", defaultDaemonPort, "IPFS peer default port")
}

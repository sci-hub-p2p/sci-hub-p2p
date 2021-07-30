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

package cmd

import (
	"os"

	"github.com/spf13/cobra"

	indexes2 "sci_hub_p2p/cmd/client/indexes"
	ipfs2 "sci_hub_p2p/cmd/client/ipfs"
	paper2 "sci_hub_p2p/cmd/client/paper"
	torrent2 "sci_hub_p2p/cmd/client/torrent"
	"sci_hub_p2p/cmd/shared"
	"sci_hub_p2p/pkg/vars"
)

var rootCmd = &cobra.Command{
	Use:   "sci-hub",
	Short: "sci-hub-p2p is cli tool to fetch paper from p2p network.",
	Long: "Complete documentation is available at " +
		"https://github.com/Trim21/sci-hub-p2p/wiki",
	Version:            vars.Ref,
	SilenceUsage:       true,
	SilenceErrors:      false,
	PersistentPreRunE:  shared.PersistentPreRunE,
	PersistentPostRunE: shared.PersistentPostRunE,
}

func init() {
	shared.SetupGlobalFlag(rootCmd)
	rootCmd.AddCommand(indexes2.Cmd, torrent2.Cmd, paper2.Cmd, shared.DebugCmd, ipfs2.Cmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

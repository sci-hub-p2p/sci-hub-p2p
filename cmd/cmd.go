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
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"sci_hub_p2p/cmd/indexes"
	"sci_hub_p2p/pkg/logger"
)

var rootCmd = &cobra.Command{
	Use:   "sci-hub",
	Short: "sci-hub-p2p is cli tool to fetch paper from p2p network.",
	Long: "Complete documentation is available at" +
		"https://github.com/Trim21/sci-hub-p2p/wiki",
	SilenceUsage:  true,
	SilenceErrors: false,
}

var debug bool

const (
	exitCode2 = 2
	exitCode1 = 1
)

func Execute() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug")
	err := logger.Setup(debug)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Can't setup logger", err)
		os.Exit(exitCode2)
	}

	rootCmd.AddCommand(indexes.IndexCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(exitCode1)
	}
}

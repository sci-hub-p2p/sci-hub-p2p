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
)

var rootCmd = &cobra.Command{
	Use:   "sci_hub",
	Short: "sci-hub-p2p is cli tool to fetch paper from p2p network.",
	Long: "Complete documentation is available at" +
		"https://github.com/Trim21/sci-hub-p2p/wiki",
	SilenceUsage:  true,
	SilenceErrors: false,
	// RunE: func(cmd *cobra.Command, args []string) error {
	//	return nil
	// },
}

var debug bool

func Execute() {
	rootCmd.AddCommand(
		indexCmd,
		// serverCmd,
		// spiderCmd,
		// cronCmd,
	)
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug")
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

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
	"bytes"
	"fmt"
	"os"
	"runtime"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"sci_hub_p2p/cmd/flag"
	"sci_hub_p2p/cmd/indexes"
	"sci_hub_p2p/cmd/paper"
	"sci_hub_p2p/cmd/torrent"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/variable"
)

var rootCmd = &cobra.Command{
	Use:   "sci-hub",
	Short: "sci-hub-p2p is cli tool to fetch paper from p2p network.",
	Long: "Complete documentation is available at" +
		"https://github.com/Trim21/sci-hub-p2p/wiki",
	Version:       variable.Ref,
	SilenceUsage:  true,
	SilenceErrors: false,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := logger.Setup()
		if err != nil {
			return errors.Wrap(err, "Can't setup logger")
		}

		return nil
	},
}

var debugCmd = &cobra.Command{
	Use:          "debug",
	Short:        "show debug message",
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		var buf bytes.Buffer
		buf.WriteString("=====build into=====\n")
		buf.WriteString(fmt.Sprintln("version:      ", variable.Ref))
		buf.WriteString(fmt.Sprintln("commit:       ", variable.Commit))
		buf.WriteString(fmt.Sprintln("compiler:     ", variable.Builder))
		buf.WriteString(fmt.Sprintln("compile time: ", variable.BuildTime))
		buf.WriteString("=====runtime into=====\n")
		buf.WriteString(fmt.Sprintln("OS:      ", runtime.GOOS))
		buf.WriteString(fmt.Sprintln("Arch:    ", runtime.GOARCH))
		buf.WriteString(fmt.Sprintln("BaseDir: ", variable.GetAppBaseDir()))
		fmt.Println(buf.String())
	},
}

func Execute() {
	rootCmd.AddCommand(indexes.Cmd, torrent.Cmd, paper.Cmd, debugCmd)

	rootCmd.PersistentFlags().BoolVar(&flag.Debug, "debug", false, "enable Debug")

	var defaultParallel = 3

	rootCmd.PersistentFlags().IntVarP(&flag.Parallel, "parallel", "n",
		defaultParallel, "how many CPU will be used")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

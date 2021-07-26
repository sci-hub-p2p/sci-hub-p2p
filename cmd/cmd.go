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
	"runtime/pprof"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"sci_hub_p2p/cmd/daemon"
	"sci_hub_p2p/cmd/flag"
	"sci_hub_p2p/cmd/indexes"
	"sci_hub_p2p/cmd/ipfs"
	"sci_hub_p2p/cmd/paper"
	"sci_hub_p2p/cmd/torrent"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/vars"
)

var rootCmd = &cobra.Command{
	Use:   "sci-hub",
	Short: "sci-hub-p2p is cli tool to fetch paper from p2p network.",
	Long: "Complete documentation is available at " +
		"https://github.com/Trim21/sci-hub-p2p/wiki",
	Version:       vars.Ref,
	SilenceUsage:  true,
	SilenceErrors: false,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := logger.Setup()
		if err != nil {
			return errors.Wrap(err, "Can't setup logger")
		}
		if flag.CPUProfile {
			logger.Info("start profile, save data to ./cpu_profile")
			f, err := os.Create("cpu_profile")
			if err != nil {
				logger.Error("failed to open ./cpu_profile to write profile data", zap.Error(err))
			} else {
				err = pprof.StartCPUProfile(f)
				if err != nil {
					logger.Error("failed to start profile", zap.Error(err))
				}
			}
		}

		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if flag.CPUProfile {
			pprof.StopCPUProfile()
		}

		return logger.Sync()
	},
}

var debugCmd = &cobra.Command{
	Use:          "debug",
	Short:        "show debug message",
	SilenceUsage: true,
	Hidden:       true,
	Run: func(cmd *cobra.Command, args []string) {
		var buf bytes.Buffer
		buf.WriteString("=====build into=====\n")
		buf.WriteString(fmt.Sprintln("version:      ", vars.Ref))
		buf.WriteString(fmt.Sprintln("commit:       ", vars.Commit))
		buf.WriteString(fmt.Sprintln("compiler:     ", vars.Builder))
		buf.WriteString(fmt.Sprintln("compile time: ", vars.BuildTime))
		buf.WriteString("=====runtime into=====\n")
		buf.WriteString(fmt.Sprintln("OS:      ", runtime.GOOS))
		buf.WriteString(fmt.Sprintln("Arch:    ", runtime.GOARCH))
		buf.WriteString(fmt.Sprintln("BaseDir: ", vars.GetAppBaseDir()))
		fmt.Println(buf.String())
	},
}

func Execute() {
	rootCmd.AddCommand(daemon.Cmd, indexes.Cmd, torrent.Cmd, paper.Cmd, debugCmd, ipfs.Cmd)

	rootCmd.PersistentFlags().StringVar(&flag.LogFile, "log-file", "", "extra logger file, eg: ./out/log.jsonlines")
	rootCmd.PersistentFlags().BoolVar(&flag.Debug, "debug", false, "enable Debug")
	var defaultParallel = 3

	rootCmd.PersistentFlags().IntVarP(&flag.Parallel, "parallel", "n",
		defaultParallel, "how many CPU will be used")

	rootCmd.PersistentFlags().BoolVar(&flag.CPUProfile, "cpu-profile", false, "generate a cpu profile")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

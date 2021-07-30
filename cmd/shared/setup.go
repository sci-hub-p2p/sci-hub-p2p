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

package shared

import (
	"os"
	"runtime/pprof"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"sci_hub_p2p/cmd/flag"
	"sci_hub_p2p/pkg/logger"
)

const defaultParallel = 3

func SetupGlobalFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&flag.LogFile, "log-file", "", "extra logger file, eg: ./out/log.jsonlines")
	cmd.PersistentFlags().BoolVar(&flag.Debug, "debug", false, "enable Debug")
	cmd.PersistentFlags().IntVar(&flag.Parallel, "parallel", defaultParallel, "how many CPU will be used")
	cmd.PersistentFlags().BoolVar(&flag.CPUProfile, "cpu-profile", false, "generate a cpu profile")
}

func PersistentPreRunE(cmd *cobra.Command, args []string) error {
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
}

func PersistentPostRunE(cmd *cobra.Command, args []string) error {
	if flag.CPUProfile {
		pprof.StopCPUProfile()
	}

	return logger.Sync()
}

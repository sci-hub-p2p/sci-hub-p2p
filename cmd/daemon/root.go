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

import (
	"github.com/spf13/cobra"

	"sci_hub_p2p/pkg/daemon"
)

var Cmd = &cobra.Command{
	Use: "daemon",
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "start daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		return daemon.Start()
	},
}

func init() {
	Cmd.AddCommand(startCmd)
}

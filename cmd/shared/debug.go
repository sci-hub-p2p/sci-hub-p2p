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
	"bytes"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"sci_hub_p2p/pkg/vars"
)

var DebugCmd = &cobra.Command{
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

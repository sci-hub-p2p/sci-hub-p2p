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

package util

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func MarkFlagsRequired(c *cobra.Command, flags ...string) error {
	for _, flag := range flags {
		err := c.MarkFlagRequired(flag)
		if err != nil {
			return errors.Wrap(err, "failed to mark flag as required")
		}
	}

	return nil
}

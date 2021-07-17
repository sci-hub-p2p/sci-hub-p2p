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

package utils

import (
	"os"

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

func EnsureDir(name string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		s, err := os.Stat(name)
		if err != nil {
			if os.IsNotExist(err) {
				err := os.MkdirAll(name, os.ModeDir)
				if err != nil {
					return errors.Wrapf(err, "can't create app base dir %s", name)
				}

				return nil
			}

			return errors.Wrap(err, "unexpected error")
		}
		if !s.IsDir() {
			return errors.Wrapf(err, "app base dir %s is not a dir", name)
		}

		return nil
	}
}

// Unique make sure all element are unique and omit slice order
func Unique(s []string) []string {
	var m = make(map[string]bool)
	for _, v := range s {
		m[v] = true
	}
	s = make([]string, 0, len(m))
	for key := range m {
		s = append(s, key)
	}

	return s
}

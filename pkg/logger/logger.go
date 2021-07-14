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

package logger

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

func Setup(debug bool) error {
	if debug {
		log.SetLevel(log.DebugLevel)
	}
	return nil
}
func Info(args ...interface{}) {
	log.Infoln(args...)
}

func Debugf(format string, args ...interface{}) {
	log.Debug(fmt.Sprintf(format, args))
}

func Debug(args ...interface{}) {
	log.Debug(args)
}

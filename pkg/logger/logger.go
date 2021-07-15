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
	log "github.com/sirupsen/logrus"

	"sci_hub_p2p/cmd/flag"
)

func Setup() error {
	if flag.Debug {
		log.SetLevel(log.DebugLevel)
	}
	log.SetFormatter(&log.TextFormatter{
		ForceQuote:       true,
		TimestampFormat:  "2006-01-02 15:04:05.000",
		DisableSorting:   true,
		PadLevelText:     true,
		QuoteEmptyFields: true,
	})

	return nil
}

func WithField(key string, value interface{}) *log.Entry {
	return log.WithField(key, value)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Info(msg ...interface{}) {
	log.Infoln(msg...)
}

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Debug(args ...interface{}) {
	log.Debugln(args...)
}

func Fatal(args ...interface{}) {
	log.Fatalln(args...)
}

func Error(args ...interface{}) {
	log.Error(args...)
}

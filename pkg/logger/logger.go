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
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"

	"sci_hub_p2p/cmd/flag"
)

func Setup() error {
	if flag.Debug {
		log.SetLevel(log.DebugLevel)
	}
	if strings.Contains(strings.ToLower(os.Getenv("LOG_LEVEL")), "trace") {
		log.SetLevel(log.TraceLevel)
	}
	log.SetFormatter(&log.TextFormatter{
		ForceQuote:             true,
		TimestampFormat:        "2006-01-02 15:04:05.000",
		DisableSorting:         false,
		PadLevelText:           true,
		QuoteEmptyFields:       true,
		DisableLevelTruncation: false,
		ForceColors:            true,
	})
	if flag.LogFile != "" {
		rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
			Filename: flag.LogFile,
			Level:    log.InfoLevel,
			Formatter: &log.JSONFormatter{
				TimestampFormat: time.RFC822,
			},
		})
		if err != nil {
			return errors.Wrap(err, "can't save log to file")
		}
		log.AddHook(rotateFileHook)
	}

	return nil
}

func Func(value string) *log.Entry {
	f := getFrame()

	return log.WithField("func", value).WithField("file", f.File).WithField("line", f.Line)
}

func WithLogger(value string) *log.Entry {
	f := getFrame()

	return log.WithField("logger", value).WithField("file", f.File).WithField("line", f.Line)
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
	f := getFrame()
	log.WithField("file", f.File).WithField("line", f.Line).Debugf(format, args...)
}

func Trace(args ...interface{}) {
	f := getFrame()
	log.WithField("file", f.File).WithField("line", f.Line).Traceln(args...)
}
func Tracef(format string, args ...interface{}) {
	f := getFrame()
	log.WithField("file", f.File).WithField("line", f.Line).Tracef(format, args...)
}
func Debug(args ...interface{}) {
	f := getFrame()
	log.WithField("file", f.File).WithField("line", f.Line).Debugln(args...)
}

func Warn(args ...interface{}) {
	f := getFrame()
	log.WithField("file", f.File).WithField("line", f.Line).Warnln(args...)
}

func Warnf(format string, args ...interface{}) {
	f := getFrame()
	log.WithField("file", f.File).WithField("line", f.Line).Warnf(format, args...)
}

func Fatal(args ...interface{}) {
	log.Fatalln(args...)
}

func Error(args ...interface{}) {
	log.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

const skip = 2

func getFrame() runtime.Frame {
	// We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
	targetFrameIndex := 1 + skip

	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+skip)
	n := runtime.Callers(0, programCounters)

	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}

	frame.File = strings.TrimPrefix(frame.File, wd)

	return frame
}

var wd string // nolint

func init() { // nolint
	var err error
	wd, err = os.Getwd()
	if err != nil {
		panic("can't get CWD")
	}
	wd = strings.ReplaceAll(wd, string(os.PathSeparator), "/")
}

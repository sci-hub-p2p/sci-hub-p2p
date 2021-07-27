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

	ds "github.com/ipfs/go-datastore"
	"github.com/natefinch/lumberjack"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"sci_hub_p2p/cmd/flag"
)

const defaultLogFileMaxSize = 100 // MB

var log *zap.Logger = zap.NewNop()

func Setup() error {
	consoleEncoding := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "msg",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("01-02 15:04:05"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	jsonEncoding := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var infoLevel = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.InfoLevel
	})
	var onlyInfo = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl <= zapcore.InfoLevel
	})
	var errorLevel = zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	if !flag.Debug {
		onlyInfo = func(lvl zapcore.Level) bool {
			return lvl == zapcore.InfoLevel
		}
	}
	cores := []zapcore.Core{
		zapcore.NewCore(zapcore.NewConsoleEncoder(consoleEncoding), zapcore.NewMultiWriteSyncer(os.Stdout), onlyInfo),
		zapcore.NewCore(zapcore.NewConsoleEncoder(consoleEncoding), zapcore.NewMultiWriteSyncer(os.Stderr), errorLevel),
	}

	if flag.LogFile != "" {
		lumberJackLogger := &lumberjack.Logger{
			Filename: flag.LogFile,
			MaxSize:  defaultLogFileMaxSize,
			Compress: false,
		}
		cores = append(cores,
			zapcore.NewCore(zapcore.NewJSONEncoder(jsonEncoding), zapcore.AddSync(lumberJackLogger), infoLevel))
	}
	log = zap.New(zapcore.NewTee(cores...))

	return nil
}

func WithLogger(name string) *zap.Logger {
	return log.Named(name)
}

func Debug(msg string, fields ...zapcore.Field) {
	log.Debug(msg, fields...)
}

func Info(msg string, fields ...zapcore.Field) {
	log.Info(msg, fields...)
}

func Warn(msg string, fields ...zapcore.Field) {
	log.Warn(msg, fields...)
}

func Error(msg string, fields ...zapcore.Field) {
	log.Error(msg, fields...)
}

func Fatal(msg string, fields ...zapcore.Field) {
	log.Fatal(msg, fields...)
}

func Sync() error {
	return errors.Wrap(log.Sync(), "failed to flush log to disk")
}

func Key(key ds.Key) zapcore.Field {
	return zap.String("key", key.String())
}

// Copyright 2021 Trim21 <trim21.me@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.

package indexes

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/itchio/lzma"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"
)

func LoadIndexReader(b *bbolt.Bucket, r io.Reader) (success int, err error) {
	reader := lzma.NewReader(r)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		var s []string

		err = json.Unmarshal(scanner.Bytes(), &s)
		if err != nil || len(s) != 2 {
			return 0, errors.Wrap(err, "can't parse json "+scanner.Text())
		}

		value, err := base64.StdEncoding.DecodeString(s[1])
		if err != nil {
			return 0, errors.Wrap(err, "can't decode base64")
		}

		key, err := url.QueryUnescape(strings.TrimSuffix(s[0], ".pdf"))
		if err != nil {
			return 0, errors.Wrap(err, "failed to URl unescape the filename")
		}

		err = b.Put([]byte(key), value)
		if err != nil {
			return 0, errors.Wrap(err, "can't save record to database")
		}

		success++
	}

	err = scanner.Err()
	if err != nil {
		return 0, errors.Wrap(err, "can't scan file")
	}

	return success, errors.Wrap(scanner.Err(), "can't scan file")
}

func LoadIndexFile(b *bbolt.Bucket, name string) (success int, err error) {
	f, err := os.Open(name)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return LoadIndexReader(b, f)
}

func LoadIndexContent(b *bbolt.Bucket, raw []byte) (success int, err error) {
	return LoadIndexReader(b, bytes.NewReader(raw))
}

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
package torrent

import (
	"bytes"
	"io"
	"os"

	bencodeMap "github.com/IncSW/go-bencode"
	"github.com/jackpal/bencode-go"
	"github.com/pkg/errors"

	"sci_hub_p2p/pkg/hash"
)

func ParseFile(name string) (*Torrent, error) {
	b, err := os.ReadFile(name)
	if err != nil {
		return nil, errors.Wrap(err, "can't open file")
	}

	return ParseRaw(b)
}

func ParseReader(r io.Reader) (*Torrent, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "can't read from buffer")
	}

	return ParseRaw(content)
}

func ParseRaw(raw []byte) (*Torrent, error) {
	t := &torrentFile{}

	err := bencode.Unmarshal(bytes.NewReader(raw), t)
	if err != nil {
		return nil, errors.Wrap(err, "content is not valid bencoding bytes")
	}

	tt, err := t.toTorrent()
	if err != nil {
		return nil, err
	}

	tt.raw = raw

	infoHash, err := getInfoHash(raw)
	if err != nil {
		return nil, err
	}

	tt.setInfoHash(infoHash)

	return tt, nil
}

func getInfoHash(content []byte) ([]byte, error) {
	// it's annoying but we have to decode it twice to calculate info-hash
	data, err := bencodeMap.Unmarshal(content)

	m, ok := data.(map[string]interface{})
	if !ok {
		return nil, errors.Wrap(err, "torrent data is not valid")
	}

	info, ok := m["info"]
	if !ok {
		return nil, errors.Wrap(err, "torrent missing `info` field")
	}

	s, err := bencodeMap.Marshal(info)
	if err != nil {
		return nil, errors.Wrap(err, "can't marshal torrent info")
	}

	return hash.Sha1SumBytes(s), nil
}

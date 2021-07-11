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

package torrent

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"

	bencodeMap "github.com/IncSW/go-bencode"
	"github.com/jackpal/bencode-go"
	"github.com/pkg/errors"
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
		return nil, errors.Wrap(err, "can't parse torrent")
	}

	tt, err := t.toTorrent()
	if err != nil {
		return nil, err
	}

	infoHash, err := getInfoHash(raw)
	if err != nil {
		return nil, err
	}

	tt.InfoHash = infoHash

	return tt, err
}

func getInfoHash(content []byte) (string, error) {
	// it's annoying but we have to decode it twice to calculate info-hash
	data, err := bencodeMap.Unmarshal(content)

	m, ok := data.(map[string]interface{})
	if !ok {
		return "", errors.Wrap(err, "torrent data is not valid")
	}

	info, ok := m["info"]
	if !ok {
		return "", errors.Wrap(err, "torrent missing `info` field")
	}

	s, err := bencodeMap.Marshal(info)
	if err != nil {
		return "", errors.Wrap(err, "can't marshal torrent info")
	}

	return sha1Sum(s), nil
}

func sha1Sum(b []byte) string {
	h := sha1.New()
	_, _ = h.Write(b)
	sum := h.Sum(nil)

	return hex.EncodeToString(sum)
}

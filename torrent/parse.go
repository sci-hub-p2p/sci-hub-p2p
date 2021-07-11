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

	"github.com/jackpal/bencode-go"
	"github.com/pkg/errors"
)

func ParseReader(reader io.Reader) (*Torrent, error) {
	t := &torrentFile{}

	err := bencode.Unmarshal(reader, t)
	if err != nil {
		return nil, errors.Wrap(err, "can't parse torrent")
	}

	tt, err := t.toTorrent()
	if err != nil {
		return nil, err
	}

	infoHash, err := getInfoHash(t.Info)
	if err != nil {
		return nil, err
	}

	tt.InfoHash = infoHash

	return tt, err
}

func ParseRaw(raw []byte) (*Torrent, error) {
	return ParseReader(bytes.NewReader(raw))
}

func getInfoHash(info info) (string, error) {
	var buf bytes.Buffer

	if err := bencode.Marshal(&buf, info); err != nil {
		return "", errors.Wrap(err, "can't marshal info to bytes")
	}

	content, err := io.ReadAll(&buf)
	if err != nil {
		return "", errors.Wrap(err, "can't read from buffer")
	}

	return sha1Sum(content), nil
}

func sha1Sum(b []byte) string {
	h := sha1.New()
	_, _ = h.Write(b)
	sum := h.Sum(nil)

	return hex.EncodeToString(sum)
}

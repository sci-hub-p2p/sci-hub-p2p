package torrent

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"io"

	"github.com/jackpal/bencode-go"
)

func ParseReader(reader io.Reader) (*Torrent, error) {
	var t = &torrentFile{}

	err := bencode.Unmarshal(reader, t)
	if err != nil {
		return nil, err
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
		return "", err
	}

	content, err := io.ReadAll(&buf)
	if err != nil {
		return "", err
	}

	return sha1Sum(content), nil

}

func sha1Sum(b []byte) string {
	var h = sha1.New()
	h.Write(b)
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}

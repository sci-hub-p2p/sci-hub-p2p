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

package client

import (
	"bytes"
	"fmt"
	"io"

	"github.com/anacrolix/log"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"github.com/pkg/errors"

	"sci_hub_p2p/pkg/hash"
	"sci_hub_p2p/pkg/indexes"
	"sci_hub_p2p/pkg/vars"
)

func Fetch(c *torrent.Client, p *indexes.PerFile, rawTorrent []byte) ([]byte, error) {
	mi, err := metainfo.Load(bytes.NewReader(rawTorrent))
	if err != nil {
		return nil, errors.Wrap(err, "can't parse torrent file")
	}

	t, err := c.AddTorrent(mi)
	if err != nil {
		return nil, errors.Wrap(err, "can't add torrent to BT client")
	}

	b, err := extract(t, p)

	return b, err
}

type nilLogger struct {
}

func (l nilLogger) Log(_ log.Msg) {

}

func GetClient() (*torrent.Client, error) {
	cfg := torrent.NewDefaultClientConfig()
	cfg.DefaultStorage = storage.NewBoltDB(vars.GetAppTmpDir())
	cfg.Bep20 = "-GT0003-"
	cfg.Logger = log.Logger{LoggerImpl: nilLogger{}}
	cfg.DisableUTP = true
	c, err := torrent.NewClient(cfg)

	return c, errors.Wrap(err, "can't initialize BitTorrent client")
}

func extract(t *torrent.Torrent, p *indexes.PerFile) ([]byte, error) {
	fmt.Println("start downloading")
	t.DownloadPieces(p.PieceStart, p.PieceEnd+1)

	var (
		tmpBinary = make([]byte, p.CompressedSize)
		file      = t.Files()[p.FileIndex]
		reader    = file.NewReader()
	)

	defer reader.Close()

	if _, err := reader.Seek(p.OffsetFromZip, io.SeekStart); err != nil {
		return nil, errors.Wrap(err, "can't download data from BitTorrent network")
	}

	if _, err := reader.Read(tmpBinary); err != nil {
		return nil, errors.Wrap(err, "can't download data from BitTorrent network")
	}

	fmt.Println("expected CID:", p.CID)

	hex, err := hash.Cid(bytes.NewReader(tmpBinary))
	if err != nil {
		return tmpBinary, errors.Wrap(err, "can't calculate CID file data")
	}

	if hex != p.CID {
		return nil, fmt.Errorf("received CID: %s %w", hex, ErrHashMisMatch)
	}

	fmt.Println("received CID:", hex)

	return tmpBinary, nil
}

var ErrHashMisMatch = errors.New("hash mismatch")

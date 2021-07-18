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

package client

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/anacrolix/log"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/pkg/constants"
	"sci_hub_p2p/pkg/hash"
	"sci_hub_p2p/pkg/indexes"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/persist"
	"sci_hub_p2p/pkg/variable"
)

func Fetch(doi string) ([]byte, error) {
	db, err := bbolt.Open(variable.GetPaperBoltPath(), constants.DefaultFileMode, bbolt.DefaultOptions)
	if err != nil {
		return nil, errors.Wrap(err, "can't open indexes database file")
	}
	var p *indexes.PerFile
	var raw []byte
	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(constants.PaperBucket())
		var err error
		p, raw, err = persist.GetPerFileAndRawTorrent(b, doi)
		if err != nil {
			return errors.Wrap(err, "can't get file indexes")
		}

		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Wrap(err,
				"find doi in index, but we can't find torrent file contains this paper, try load torrents again")
		}

		return nil, errors.Wrap(err, "can't find indexes of this paper")
	}

	c, err := getClient()
	if err != nil {
		logger.Fatal(err)
	}
	defer c.Close()
	mi, err := metainfo.Load(bytes.NewReader(raw))
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

func getClient() (*torrent.Client, error) {
	cfg := torrent.NewDefaultClientConfig()
	cfg.DefaultStorage = storage.NewBoltDB(variable.GetAppTmpDir())
	cfg.Bep20 = "-GT0003-"
	cfg.Logger = log.Logger{LoggerImpl: nilLogger{}}
	cfg.DisableUTP = true
	c, err := torrent.NewClient(cfg)

	return c, errors.Wrap(err, "can't initialize BitTorrent client")
}

func extract(t *torrent.Torrent, p *indexes.PerFile) ([]byte, error) {
	fmt.Println("start downloading")
	t.DownloadPieces(p.PieceStart, p.PieceEnd+1)
	var tmpBinary = make([]byte, p.CompressedSize)
	file := t.Files()[p.FileIndex]
	reader := file.NewReader()
	defer reader.Close()

	if _, err := reader.Seek(p.OffsetFromZip, io.SeekStart); err != nil {
		return nil, errors.Wrap(err, "can't download data from BitTorrent network")
	}
	if _, err := reader.Read(tmpBinary); err != nil {
		return nil, errors.Wrap(err, "can't download data from BitTorrent network")
	}

	fmt.Println("expected sha256:", p.Sha256)
	hex := hash.Sha256SumHex(tmpBinary)
	if hex != p.Sha256 {
		return nil, fmt.Errorf("received sha256: %s %w", hex, ErrHashMisMatch)
	}
	fmt.Println("received sha256:", hex)

	return tmpBinary, nil
}

var ErrHashMisMatch = errors.New("hash mismatch")

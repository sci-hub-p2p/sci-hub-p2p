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
	"fmt"
	"io"
	"os"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"
	"go.etcd.io/bbolt"

	torrent2 "sci_hub_p2p/internal/torrent"
	"sci_hub_p2p/pkg/constants"
	"sci_hub_p2p/pkg/hash"
	"sci_hub_p2p/pkg/indexes"
	"sci_hub_p2p/pkg/logger"
	"sci_hub_p2p/pkg/persist"
	"sci_hub_p2p/pkg/variable"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func Fetch(doi string) {
	c, err := getClient()
	if err != nil {
		logger.Fatal(err)
	}
	defer c.Close()
	torrentPath := "./out/d57b1013eee9138a8906bcd274d727b5d7e8a307.torrent"
	myT, err := torrent2.ParseFile(torrentPath)
	checkErr(err)
	t, err := c.AddTorrentFromFile(torrentPath)
	checkErr(err)

	extract("10.1002/%28sici%291096-9098%28199710%2966%3A2%3C110%3A%3Aaid-jso7%3E3.0.co%3B2-g.pdf", t, myT)

}

func getRecord(doi string, t *torrent2.Torrent) (indexes.PerFile, error) {
	db, err := bbolt.Open("./out/"+t.InfoHash+".indexes", constants.DefaultFileMode, bbolt.DefaultOptions)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	var raw []byte
	err = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(constants.PaperBucket())
		raw = b.Get([]byte(doi))
		if raw == nil {
			return persist.ErrNotFound
		}
		return nil
	})
	if err != nil {
		return indexes.PerFile{}, err
	}
	return indexes.LoadRecordV0(raw).Build(doi, t), nil
}

func getClient() (*torrent.Client, error) {
	cfg := torrent.NewDefaultClientConfig()
	cfg.DefaultStorage = storage.NewBoltDB(variable.GetAppBaseDir())
	cfg.Bep20 = "-GT0003-"
	c, err := torrent.NewClient(cfg)
	return c, err
}

func extract(doi string, t *torrent.Torrent, internalT *torrent2.Torrent) (indexes.PerFile, error) {
	p, err := getRecord(doi, internalT)
	checkErr(err)

	fmt.Println("starts waiting to download")
	t.DownloadPieces(p.PieceStart, p.PieceEnd+1)

	var tmpBinary = make([]byte, p.CompressedSize)
	file := t.Files()[p.FileIndex]
	fmt.Println(file.DisplayPath())
	reader := file.NewReader()
	_, err = reader.Seek(p.OffsetFromZip, io.SeekStart)
	checkErr(err)
	_, err = reader.Read(tmpBinary)
	checkErr(err)

	hex := hash.Sha256SumHex(tmpBinary)
	logger.Info(hex)
	if hex != p.Sha256 {
		logger.Fatal("sha256 mismatch, expected", p.Sha256, "actual", hex)
	}
	os.WriteFile("./out/papers/map-reduce.pdf", tmpBinary, constants.DefaultFileMode)
	return p, nil
}

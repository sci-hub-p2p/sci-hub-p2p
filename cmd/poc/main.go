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

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/storage"

	"go.etcd.io/bbolt"

	torrent2 "sci_hub_p2p/internal/torrent"
	"sci_hub_p2p/pkg/constants"
	"sci_hub_p2p/pkg/hash"
	"sci_hub_p2p/pkg/indexes"
	"sci_hub_p2p/pkg/persist"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

type P string

func (receiver P) String() string {
	return string(receiver)
}

func main() {
	cfg := torrent.NewDefaultClientConfig()
	fmt.Println(cfg.PeerID)
	impl := storage.NewBoltDB("./")
	cfg.DefaultStorage = impl
	// cfg.PeerID = "-BOWxxx-7ah267cakuha"
	cfg.Bep20 = "-GT0003-"
	c, err := torrent.NewClient(cfg)
	checkErr(err)
	myT, err := torrent2.ParseFile("./out/d57b1013eee9138a8906bcd274d727b5d7e8a307.torrent")
	checkErr(err)
	t, err := c.AddTorrentFromFile("./out/d57b1013eee9138a8906bcd274d727b5d7e8a307.torrent")
	checkErr(err)
	p, err := getRecord("10.1145/1327452.1327492.pdf", myT)
	checkErr(err)
	var myPeer = torrent.PeerInfo{
		Addr:    P("36.35.28.92:58093"),
		Trusted: true,
	}
	t.AddPeers([]torrent.PeerInfo{myPeer})
	fmt.Println("starts waiting to download")
	fmt.Println(p.PieceStart, p.PieceEnd, p.String())
	t.DownloadPieces(p.PieceStart, p.PieceEnd+1)
	for range time.Tick(time.Second * 10) {
		s := t.Stats()
		downloaded := true
		for _, pIndex := range p.Pieces {
			fmt.Println("piece", pIndex)
			ss := t.PieceState(pIndex)

			if !ss.Complete {
				downloaded = false
			}

			if ss.Complete {
				t.Piece(pIndex).VerifyData()
			}

			fmt.Println(MustMarshal(ss))
		}
		if downloaded {
			break
		}
		fmt.Println(MustMarshal(s))
	}
	var tmpBinary = make([]byte, p.CompressedSize)

	for _, pIndex := range p.Pieces {
		currentPiece := t.Piece(pIndex)
		if currentPiece == nil {
			return
		}
		torrentImpl, err := impl.OpenTorrent(t.Info(), t.InfoHash())
		checkErr(err)

		readLen, err := torrentImpl.Piece(currentPiece.Info()).ReadAt(tmpBinary, p.OffsetFromPiece)
		checkErr(err)
		fmt.Println(readLen)
	}
	hex := hash.Sha256SumHex(tmpBinary)
	os.WriteFile("./map-reduce.pdf", tmpBinary, constants.DefaultFileMode)
	fmt.Println(hex, p.Sha256)
}

func getTorrent(db *bbolt.DB, hash []byte) (*torrent2.Torrent, error) {
	var t *torrent2.Torrent
	err := db.View(func(tx *bbolt.Tx) error {
		var err error
		b := tx.Bucket(constants.TorrentBucket())
		t, err = persist.GetTorrent(b, hash)
		if err != nil {
			return err
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return t, nil

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

func MustMarshalIndent(data interface{}) string {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}

func MustMarshal(data interface{}) string {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return string(b)
}

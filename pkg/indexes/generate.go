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

package indexes

import (
	"archive/zip"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/nozzle/throttler"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/cmd/flag"
	"sci_hub_p2p/internal/torrent"
	"sci_hub_p2p/pkg/hash"
	"sci_hub_p2p/pkg/logger"
)

type PDFFileOffSet struct {
	IndexInDB
	DOI string
}

func (f PDFFileOffSet) Key() []byte {
	return []byte(f.DOI)
}

// IndexZipFile is intended to be used in goroutine for parallel.
func IndexZipFile(c chan *PDFFileOffSet, dataDir string, index int, t *torrent.Torrent) {
	pieceLength := t.PieceLength
	var currentZipOffset int64 = 0
	for i, file := range t.Files {
		if i < index {
			currentZipOffset += file.Length
		} else {
			break
		}
	}
	torrentFile := t.Files[index]
	fs := filepath.Join(torrentFile.Path...)
	abs := filepath.Join(dataDir, fs)
	r, err := zip.OpenReader(abs)
	if err != nil {
		log.Fatalf("can't open file %s: %s", abs, err)
	}

	defer r.Close()

	for _, file := range r.File {
		i := &PDFFileOffSet{
			DOI: file.Name, // file name is just doi
			IndexInDB: IndexInDB{
				InfoHash:         [20]byte{},
				PieceStart:       0,
				DataOffset:       0,
				CompressedMethod: 0,
				CompressedSize:   0,
				Sha256:           [32]byte{},
			},
		}
		infoHash, err := hex.DecodeString(t.InfoHash)
		if err != nil {
			log.Fatal(err)
		}
		copy(i.InfoHash[:], infoHash)

		offset, err := file.DataOffset()
		if err != nil {
			log.Fatalf("can't get file offset in zip %s: %s", abs, err)
		}
		// FIXME: this need to be convert to offset from first piece, not file start
		i.DataOffset = uint32(offset)

		i.PieceStart = uint32((int64(i.DataOffset) + currentZipOffset) / int64(pieceLength))
		i.CompressedMethod = file.Method
		i.CompressedSize = file.CompressedSize64
		f, err := file.Open()
		if err != nil {
			log.Fatalf("can't decompress file %s in zip %s: %s", file.Name, abs, err)
		}
		sha256, err := hash.Sha256SumReader(f)
		if err != nil {
			log.Fatalf("can't decompress file %s in zip %s: %s", file.Name, abs, err)
		}
		copy(i.Sha256[:], sha256)
		c <- i
	}

	return
}

func Generate(dirName, outDir string, t *torrent.Torrent) error {
	fmt.Println("start generate indexes")
	c := make(chan *PDFFileOffSet, flag.Parallel)
	th := throttler.New(flag.Parallel, len(t.Files))

	go collectResult(c, outDir, t)

	for i, file := range t.Files {
		logger.Debug("skip hash check here because files are too big, " +
			"hopefully ew didn't generate indexes from wrong data")

		go func(index int, file torrent.File) {
			fs := filepath.Join(file.Path...)
			if !strings.HasSuffix(fs, ".zip") {
				th.Done(nil)

				return
			}
			abs := filepath.Join(dirName, fs)
			s, err := os.Stat(abs)
			if err != nil {
				th.Done(err)
				log.Fatalf("can't generate indexes, file %s is broken", fs)
			}
			if s.Size() != file.Length {
				log.Fatalf(
					"can't generate indexes, file %s has a wrong size, expected %d",
					fs, file.Length)
			}
			IndexZipFile(c, dirName, index, t)
			th.Done(nil)
		}(i, file)
		th.Throttle()
	}

	return nil
}

func collectResult(c chan *PDFFileOffSet, outDir string, t *torrent.Torrent) {
	var defaultFileMode os.FileMode = 0644
	out := filepath.Join(outDir, t.InfoHash+".indexes")
	db, err := bbolt.Open(out, defaultFileMode, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("can't open %s to write indexes: %s", out, err)
	}
	defer db.Close()

	bar := pb.StartNew(filesPerTorrent)
	defer bar.Finish()

	db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucket([]byte("paper"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		for i := range c {
			bar.Increment()
			err = b.Put(i.Key(), i.Dump())
			if err != nil {
				logger.Error(err)
			}
		}

		return nil
	})
}

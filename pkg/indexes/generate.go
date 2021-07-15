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
	"compress/gzip"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/cmd/flag"
	"sci_hub_p2p/internal/torrent"
	"sci_hub_p2p/pkg/hash"
	"sci_hub_p2p/pkg/logger"
)

const filesPerTorrent = 100_000

type PDFFileOffSet struct {
	Record
	DOI string
}

func (f PDFFileOffSet) Key() []byte {
	return []byte(f.DOI)
}

// IndexZipFile is intended to be used in goroutine for parallel.
func IndexZipFile(c chan *PDFFileOffSet, dataDir string, index int, t *torrent.Torrent) {
	pieceLength := t.PieceLength
	var currentZipOffset int64
	for i, file := range t.Files {
		if i < index {
			currentZipOffset += file.Length
		} else {
			break
		}
	}
	torrentFile := t.Files[index]
	fs := filepath.Join(torrentFile.Path...)
	abs := filepath.Join(dataDir, t.Name, fs)
	r, err := zip.OpenReader(abs)
	if err != nil {
		log.Fatalf("can't open file %s: %s", abs, err)
	}

	defer r.Close()

	for _, file := range r.File {
		if file.CompressedSize64 == 0 {
			continue
		}

		i := &PDFFileOffSet{
			DOI: file.Name, // file name is just doi
			Record: Record{
				InfoHash:         [20]byte{},
				CompressedMethod: file.Method,
				CompressedSize:   file.CompressedSize64,
				Sha256:           [32]byte{},
			},
		}
		infoHash, _ := hex.DecodeString(t.InfoHash)
		copy(i.InfoHash[:], infoHash)

		offset, err := file.DataOffset()
		if err != nil {
			r.Close()
			log.Fatalf("can't offset in zip %s: %s, maybe file is broken", abs, err)
		}

		i.PieceStart = uint32((offset + currentZipOffset) / int64(pieceLength))
		i.OffsetInPiece = uint32((offset + currentZipOffset) % int64(pieceLength))

		f, err := file.Open()
		if err != nil {
			log.Fatalf("can't decompress file %s in zip %s: %s", file.Name, abs, err)
		}
		defer f.Close()

		sha256, err := hash.Sha256SumReaderBytes(f)
		if err != nil {
			log.Fatalf("can't decompress file %s in zip %s: %s", file.Name, abs, err)
		}

		copy(i.Sha256[:], sha256)

		c <- i
	}
}

func Generate(dirName, outDir string, t *torrent.Torrent) error {
	fmt.Println("start generate indexes")
	c := make(chan *PDFFileOffSet, flag.Parallel)

	done := make(chan int)
	in := make(chan int)

	var wg sync.WaitGroup
	wg.Add(flag.Parallel)

	go collectResult(c, outDir, t, done)

	for i := 0; i < flag.Parallel; i++ {
		go func() {
			for i := range in {
				IndexZipFile(c, dirName, i, t)
			}
			wg.Done()
		}()
	}

	logger.Debug("skip hash check here because files are too big,",
		"hopefully we didn't generate indexes from wrong data")

	for i, file := range t.Files {
		fs := filepath.Join(file.Path...)
		abs := filepath.Join(dirName, t.Name, fs)
		logger.Debug(abs, fs)
		if !strings.HasSuffix(fs, ".zip") {
			continue
		}
		s, err := os.Stat(abs)
		if err != nil {
			log.Fatalf("can't generate indexes, file %s is broken", fs)
		}
		if s.Size() != file.Length {
			log.Fatalf(
				"can't generate indexes, file %s has a wrong size, expected %d",
				fs, file.Length)
		}
		in <- i
	}

	for len(in) > 0 {
		time.Sleep(time.Second)
	}
	close(in)
	wg.Wait()
	close(c)
	<-done

	return nil
}

func collectResult(c chan *PDFFileOffSet, outDir string, t *torrent.Torrent, done chan int) {
	var defaultFileMode os.FileMode = 0644
	out := filepath.Join(outDir, t.InfoHash+".indexes")
	db, err := bbolt.Open(out, defaultFileMode, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("can't open %s to write indexes: %s", out, err)
	}
	defer func(db *bbolt.DB) {
		err := db.Close()
		if err != nil {
			logger.Error(err)
		}
	}(db)

	bar := pb.StartNew(filesPerTorrent)
	defer bar.Finish()

	err = db.Batch(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("paper-v0"))
		if err != nil {
			return errors.Wrap(err, "can't create bucket, maybe indexes file is not writeable")
		}
		for i := range c {
			bar.Increment()
			d := i.DumpV0()
			err = b.Put(i.Key(), d)
			if err != nil {
				log.Fatalln("can't save record:", err)
			}
		}

		return nil
	})

	if err != nil {
		logger.Fatal("can't write indexes:", err)
	}

	err = db.View(func(tx *bbolt.Tx) error {
		fs := fmt.Sprintf("./out/%s.indexes.gz", t.InfoHash)
		f, err := os.Create(fs)
		if err != nil {
			return errors.Wrapf(err, "can't create file %s", fs)
		}
		r, err := gzip.NewWriterLevel(f, gzip.BestCompression)
		if err != nil {
			return errors.Wrap(err, "can't compress indexes")
		}
		defer f.Close()
		_, err = tx.WriteTo(r)

		return errors.Wrap(err, "can't dump indexes to file")
	})

	if err != nil {
		logger.Error(err)
	}

	logger.Debug("sync database")
	if err := db.Sync(); err != nil {
		logger.Error(err)
	}
	done <- 1
}

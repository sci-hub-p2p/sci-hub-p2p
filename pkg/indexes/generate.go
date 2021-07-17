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
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/itchio/lzma"
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/cmd/flag"
	"sci_hub_p2p/internal/torrent"
	"sci_hub_p2p/internal/utils"
	"sci_hub_p2p/pkg/constants"
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
func IndexZipFile(c chan *PDFFileOffSet, dataDir string, index int, t *torrent.Torrent) error {
	var currentZipOffset int64
	file := t.Files[index]
	fs := filepath.Join(file.Path...)
	abs := filepath.Join(dataDir, t.Name, fs)

	if !strings.HasSuffix(fs, ".zip") {
		return nil
	}
	s, err := os.Stat(abs)
	if err != nil {
		return errors.Wrapf(err, "can't generate indexes, file %s is broken", fs)
	}
	if s.Size() != file.Length {
		return errors.Wrapf(err, "can't generate indexes, file %s has a wrong size, expected %d", fs, file.Length)
	}

	for i, file := range t.Files {
		if i < index {
			currentZipOffset += file.Length
		} else {
			break
		}
	}

	r, err := zip.OpenReader(abs)
	if err != nil {
		return errors.Wrap(err, "can't open zip f "+abs)
	}

	defer r.Close()

	var infoHash [20]byte
	copy(infoHash[:], t.RawInfoHash())

	for _, f := range r.File {
		if f.CompressedSize64 == 0 {
			continue
		}
		i, err := zipFileToRecord(f, currentZipOffset, t.PieceLength)
		if err != nil {
			return err
		}
		i.InfoHash = infoHash
		c <- i
	}

	return nil
}

func zipFileToRecord(file *zip.File, currentZipOffset int64, pieceLength int64) (*PDFFileOffSet, error) {
	i := &PDFFileOffSet{
		DOI: file.Name, // file name is just doi
		Record: Record{
			InfoHash:         [20]byte{},
			CompressedMethod: file.Method,
			CompressedSize:   file.CompressedSize64,
			Sha256:           [32]byte{},
		},
	}

	offset, err := file.DataOffset()
	if err != nil {
		return nil, errors.Wrapf(err, "can't offset in zip %s: maybe zip file is broken", file.Name)
	}

	i.PieceStart = uint32((offset + currentZipOffset) / pieceLength)
	i.OffsetInPiece = (offset + currentZipOffset) % pieceLength

	f, err := file.Open()
	if err != nil {
		return nil, errors.Wrapf(err, "can't decompress file %s", file.Name)
	}
	defer f.Close()

	sha256, err := hash.Sha256SumReaderBytes(f)
	if err != nil {
		return nil, errors.Wrapf(err, "can't decompress file %s", file.Name)
	}
	copy(i.Sha256[:], sha256)

	return i, nil
}

func Generate(dataDir, outDir string, t *torrent.Torrent) error {
	exist, err := utils.DirExist(filepath.Join(dataDir, t.Name))
	if err != nil {
		return errors.Wrap(err, "can't find torrent data")
	}
	if !exist {
		return errors.New("can't find torrent data")
	}

	var c = make(chan *PDFFileOffSet, flag.Parallel)
	var done = make(chan int)
	var in = make(chan int)
	var wg sync.WaitGroup
	var out = filepath.Join(outDir, t.InfoHash+".indexes")

	defer close(done)
	wg.Add(flag.Parallel)

	db, err := bbolt.Open(out, constants.DefaultFileMode, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return errors.Wrapf(err, "can't open %s to write indexes", out)
	}
	defer db.Close()
	go collectResult(c, outDir, t, done, db)

	for i := 0; i < flag.Parallel; i++ {
		go func(index int, t *torrent.Torrent) {
			for i := range in {
				err := IndexZipFile(c, dataDir, i, t)
				if err != nil {
					logger.Error(err)

					break
				}
			}
			logger.Debugf("exit worker %d", index+1)
			wg.Done()
		}(i, t)
	}

	logger.Debug("skip hash check here because files are too big,",
		"hopefully we didn't generate indexes from wrong data")

	for i := range t.Files {
		in <- i
	}

	logger.Debug("wait for all zip file to be indexed")
	for len(in) > 0 {
		time.Sleep(time.Second)
	}
	close(in)

	logger.Debug("wait all worker exit")
	wg.Wait()

	close(c)
	logger.Debug("wait closing write db worker")
	<-done

	return nil
}

func collectResult(c chan *PDFFileOffSet, outDir string, t *torrent.Torrent, done chan int, db *bbolt.DB) {
	bar := pb.StartNew(filesPerTorrent)
	err := db.Batch(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(constants.PaperBucket())
		if err != nil {
			return errors.Wrap(err, "can't create bucket, maybe indexes file is not writeable")
		}
		for i := range c {
			bar.Increment()
			d := i.DumpV0()
			err = b.Put(i.Key(), d)
			if err != nil {
				return errors.Wrap(err, "can't save record")
			}
		}

		return nil
	})
	bar.Finish()

	if err != nil {
		logger.Error("can't save indexes:", err)

		return
	}

	fmt.Println("start dumping data to file")
	err = db.View(func(tx *bbolt.Tx) error {
		return dumpToFile(tx, filepath.Join(outDir, t.InfoHash))
	})

	if err != nil {
		logger.Error("can't dump database", err)
	}

	logger.Debug("sync database")
	if err := db.Sync(); err != nil {
		logger.Error(err)
	}
	done <- 1
}

func dumpToFile(tx *bbolt.Tx, name string) error {
	t, err := os.Create(name + ".jsonlines.lzma")
	if err != nil {
		return errors.Wrap(err, "can't create lzma file to save indexes")
	}
	defer func() {
		if e := t.Close(); e != nil {
			e = errors.Wrap(e, "can't save lzma file to disk")
			if err == nil {
				err = e
			} else {
				logger.Error(e)
			}
		}
	}()

	// Assume bucket exists and has keys
	b := tx.Bucket(constants.PaperBucket())

	c := b.Cursor()

	w := lzma.NewWriterLevel(t, lzma.BestCompression)
	defer w.Close()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		b64 := base64.StdEncoding.EncodeToString(v)
		_, err := fmt.Fprintf(w, "[\"%s\", \"%s\"]\n", k, b64)
		if err != nil {
			return errors.Wrap(err, "can't write to compressed file")
		}
	}

	return nil
}

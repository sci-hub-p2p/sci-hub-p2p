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
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cheggaaa/pb/v3"
	"github.com/pkg/errors"

	"sci_hub_p2p/internal/torrent"
	"sci_hub_p2p/pkg/hash"
	"sci_hub_p2p/pkg/logger"
)

func indexZipFile(r *zip.ReadCloser, f *File) (err error) {
	var sha1Buffer bytes.Buffer
	var sha256Buffer bytes.Buffer

	bar := pb.StartNew(len(r.File))
	defer bar.Finish()
	for _, file := range r.File {
		bar.Increment()
		if file == nil {
			continue
		}

		if file.CompressedSize64 == 0 {
			continue
		}
		offset, err := file.DataOffset()
		if err != nil {
			return errors.Wrap(err, "zip file broken")
		}

		r, err := file.Open()
		if err != nil {
			return errors.Wrap(err, "can't decompress zip file")
		}
		defer r.Close()

		sha1, sha256, err := hash.Sha1Sha256SumReader(r)
		if err != nil {
			return errors.Wrapf(err, "can't hash file %s", file.Name)
		}

		f.FileNames = append(f.FileNames, file.Name)
		f.Methods = append(f.Methods, file.Method)
		f.Offset = append(f.Offset, offset)
		f.CompressedSizes = append(f.CompressedSizes, file.CompressedSize64)
		f.Crc32 = append(f.Crc32, file.CRC32)

		sha1Buffer.Write(sha1)
		sha256Buffer.Write(sha256)
	}

	sha1, err := io.ReadAll(&sha1Buffer)
	if err != nil {
		return errors.Wrap(err, "can't hash files")
	}
	f.Sha1 = append(f.Sha1, sha1...)

	sha256, err := io.ReadAll(&sha256Buffer)
	if err != nil {
		return errors.Wrap(err, "can't hash files")
	}
	f.Sha256 = append(f.Sha256, sha256...)

	return nil
}

func FromDataDir(dirName string, t *torrent.Torrent) (*File, error) {
	f := NewWithPre(filesPerTorrent)
	f.InfoHash = t.InfoHash

	fmt.Println("start generate indexes")
	totalZipFiles := len(t.Files)
	for i, file := range t.Files {
		fmt.Printf("\nIndexing file %d/%d\n", i+1, totalZipFiles)
		fs := filepath.Join(file.Path...)
		abs := filepath.Join(dirName, fs)
		s, err := os.Stat(abs)
		if err != nil {
			return nil, fmt.Errorf("can't generate indexes, file %s is broken %w",
				fs, ErrTorrentDataBroken)
		}
		if s.Size() != file.Length {
			return nil, errors.Wrapf(ErrTorrentDataBroken,
				"can't generate indexes, file %s has a wrong size, expected %d",
				fs, file.Length)
		}
		logger.Debug("skip hash check here because files are too big, " +
			"hopefully ew didn't generate indexes from wrong data")

		r, err := zip.OpenReader(abs)
		if err != nil {
			_ = r.Close()

			return nil, errors.Wrap(err, "can't open zip file")
		}
		err = indexZipFile(r, &f)
		// should close file right after index, don't use defer
		if err != nil {
			_ = r.Close()

			return nil, err
		}
		// we don't write data, just omit error
		_ = r.Close()
	}

	return &f, nil
}

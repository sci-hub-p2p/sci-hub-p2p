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

// Package indexes zip file indexes
package indexes

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/pkg/errors"

	"sci_hub_p2p/internal/torrent"
)

type Record struct {
	OffsetInPiece    int64
	CompressedSize   uint64
	PieceStart       uint32
	CompressedMethod uint16
	CID              [38]byte // v1 with DagProtobuf blake2b-256 size-262144 raw-leaves
	InfoHash         [20]byte
}

func (r Record) String() string {
	return fmt.Sprintf("Record{infohash=%s, compressedSize=%d, CID=%s}",
		hex.EncodeToString(r.InfoHash[:]), r.CompressedSize, hex.EncodeToString(r.CID[:]))
}

func (r Record) HexInfoHash() string {
	return hex.EncodeToString(r.InfoHash[:])
}

func (r Record) DumpV0() []byte {
	var buf bytes.Buffer
	// buffer.Write won't return a err
	_ = binary.Write(&buf, binary.LittleEndian, r.InfoHash)
	_ = binary.Write(&buf, binary.LittleEndian, r.PieceStart)
	_ = binary.Write(&buf, binary.LittleEndian, r.OffsetInPiece)
	_ = binary.Write(&buf, binary.LittleEndian, r.CompressedMethod)
	_ = binary.Write(&buf, binary.LittleEndian, r.CompressedSize)
	_ = binary.Write(&buf, binary.LittleEndian, r.CID)

	return buf.Bytes()
}

func LoadRecordV0(p []byte) *Record {
	var i = &Record{}
	var buf = bytes.NewBuffer(p)
	_ = binary.Read(buf, binary.LittleEndian, i.InfoHash[:])
	_ = binary.Read(buf, binary.LittleEndian, &i.PieceStart)
	_ = binary.Read(buf, binary.LittleEndian, &i.OffsetInPiece)
	_ = binary.Read(buf, binary.LittleEndian, &i.CompressedMethod)
	_ = binary.Read(buf, binary.LittleEndian, &i.CompressedSize)
	_ = binary.Read(buf, binary.LittleEndian, i.CID[:])

	return i
}

func (r Record) Build(doi string, t *torrent.Torrent) (*PerFile, error) {
	var pieceOffset = t.PieceLength*int64(r.PieceStart) + r.OffsetInPiece
	var currentZipOffset int64
	var fileStart int64 = -1
	var f torrent.File
	var fileIndex int

	for i, file := range t.Files {
		if currentZipOffset+file.Length > pieceOffset {
			fileStart = currentZipOffset
			// i = fileIndex
			f = file
			fileIndex = i

			break
		}

		currentZipOffset += file.Length
	}

	_, c, err := cid.CidFromBytes(r.CID[:])
	if err != nil {
		return nil, errors.Wrap(err, "can't parse CID")
	}

	return &PerFile{
		Doi:             doi,
		CompressMethod:  r.CompressedMethod,
		CompressedSize:  int64(r.CompressedSize),
		FileName:        f.Name(),
		CID:             c,
		Pieces:          makeRange(int(r.PieceStart), int(r.PieceStart)+int(int64(r.CompressedSize)/t.PieceLength)),
		PieceStart:      int(r.PieceStart),
		PieceEnd:        int(r.PieceStart) + int(int64(r.CompressedSize)/t.PieceLength),
		PieceLength:     t.PieceLength,
		OffsetFromZip:   r.OffsetInPiece + int64(r.PieceStart)*t.PieceLength - fileStart,
		OffsetFromPiece: r.OffsetInPiece,
		FileIndex:       fileIndex,
		File:            f.Copy(),
		Torrent:         t.Copy(),
	}, nil
}

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}

	return a
}

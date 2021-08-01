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

package indexes_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"

	"sci_hub_p2p/pkg/indexes"
)

func TestDumpLoad(t *testing.T) {
	t.Parallel()

	o := &indexes.Record{
		InfoHash: [20]byte{132, 56, 215, 195, 86, 34, 151, 137, 161,
			218, 81, 62, 114, 68, 5, 245, 136, 178, 91, 97},
		PieceStart:       11111,
		OffsetInPiece:    888137412,
		CompressedMethod: 8,
		CompressedSize:   13241729341923,
		CID: [38]byte{18, 1, 101, 51, 98, 48, 99, 52, 52, 50,
			57, 56, 102, 99, 49, 99, 49, 52,
			57, 97, 102, 98, 102, 52, 99, 56,
			57, 57, 54, 102, 98, 57, 50, 52},
	}

	b := o.DumpV0()
	n := indexes.LoadRecordV0(b)

	assert.Equal(t, hex.EncodeToString(o.InfoHash[:]), hex.EncodeToString(n.InfoHash[:]))
	assert.Equal(t, o.PieceStart, n.PieceStart)
	assert.Equal(t, o.OffsetInPiece, n.OffsetInPiece)
	assert.Equal(t, o.CompressedMethod, n.CompressedMethod)
	assert.Equal(t, o.CompressedSize, n.CompressedSize)
	assert.Equal(t, hex.EncodeToString(o.CID[:]), hex.EncodeToString(n.CID[:]))
}

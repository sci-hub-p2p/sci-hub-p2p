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

package hash_test

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"

	_ "sci_hub_p2p/internal/testing"
	"sci_hub_p2p/pkg/hash"
	"sci_hub_p2p/pkg/indexes"
)

func TestCID(t *testing.T) {
	t.Parallel()
	e, err := cid.Parse("bafykbzaceavd6aaauynuqgkkrg6lapmno5crbsyinmp3um5sn3daztzsghvl2")
	assert.Nil(t, err)

	raw, err := os.ReadFile("./testdata/big_file.bin")
	assert.Nil(t, err)

	a, err := hash.Cid(bytes.NewBuffer(raw))
	assert.Nil(t, err)
	assert.EqualValues(t, e.Hash(), a.Hash(), fmt.Sprintln(e.Prefix(), a.Prefix()))
}

func TestCIDSaved(t *testing.T) {
	t.Parallel()
	var r = indexes.Record{
		InfoHash:         [20]byte{},
		PieceStart:       0,
		OffsetInPiece:    0,
		CompressedMethod: 0,
		CompressedSize:   0,
		MultiHash:        [36]byte{},
	}

	raw, err := os.ReadFile("./testdata/big_file.bin")
	assert.Nil(t, err)

	a, err := hash.Black2dBalancedSized256K(bytes.NewBuffer(raw))
	assert.Nil(t, err)
	copy(r.MultiHash[:], a)

	n := indexes.LoadRecordV0(r.DumpV0())

	assert.EqualValues(t, a, n.MultiHash[:], "cid hash should be the save after dump and load")
}

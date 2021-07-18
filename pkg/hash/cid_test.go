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
	"os"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"

	_ "sci_hub_p2p/internal/testing"
	"sci_hub_p2p/pkg/hash"
)

func TestSha256CidBalanced(t *testing.T) {
	t.Parallel()

	c, err := cid.Parse("QmVBAYRwHA5zCbteHvY7psWdgVVAMcPkuYS5hAWFPvVXiS")
	assert.Nil(t, err)

	b, err := os.ReadFile("./testdata/sm_00900000-00999999.torrent")
	assert.Nil(t, err)

	h, err := hash.Sha256CidBalanced(bytes.NewBuffer(b))
	assert.Nil(t, err)

	assert.EqualValues(t, c.Hash(), h)
}

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

package dagserv_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"sci_hub_p2p/pkg/dagserv"
)

func TestZipArchive(t *testing.T) {
	raw, err := os.ReadFile("./../../testdata/big_file.bin")
	assert.Nil(t, err)
	t.Parallel()

	n, err := dagserv.Build(raw, 0)
	assert.Nil(t, err)
	fmt.Println(n.Cid())
}

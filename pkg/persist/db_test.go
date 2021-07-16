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

package persist_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"

	_ "sci_hub_p2p/internal/testing"
	"sci_hub_p2p/internal/torrent"
	"sci_hub_p2p/pkg/constants"
	"sci_hub_p2p/pkg/persist"
)

func TestSaveTorrent(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	db, err := bbolt.Open(filepath.Join(tmp, "d.bolt"), 0600, bbolt.DefaultOptions)
	assert.Nil(t, err)
	defer func() { assert.Nil(t, db.Close()) }()
	assert.Nil(t, db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucket(constants.TorrentBucket())
		assert.Nil(t, err)
		tor, err := torrent.ParseFile("./testdata/sm_00900000-00999999.torrent")
		assert.Nil(t, err)
		assert.Nil(t, persist.PutTorrent(b, tor))

		return nil
	}))
}

func TestGetTorrent(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	db, err := bbolt.Open(filepath.Join(tmp, "d.bolt"), 0600, bbolt.DefaultOptions)
	assert.Nil(t, err)
	defer func() { assert.Nil(t, db.Close()) }()
	assert.Nil(t, db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucket(constants.TorrentBucket())
		assert.Nil(t, err)
		tor, err := torrent.ParseFile("./testdata/sm_00900000-00999999.torrent")
		assert.Nil(t, err)
		assert.Nil(t, persist.PutTorrent(b, tor))
		newT, err := persist.GetTorrent(b, tor.RawInfoHash())
		assert.Nil(t, err)
		assert.Equal(t, tor, newT)

		return nil
	}))
}

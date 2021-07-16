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

package db

import (
	"github.com/pkg/errors"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/internal/torrent"
	"sci_hub_p2p/pkg/constants"
)

var ErrNotFound = errors.New("not found in database")

func GetTorrent(tx *bbolt.Tx, hash []byte) (*torrent.Torrent, error) {
	t := tx.Bucket(constants.TorrentBucket()).Get(hash)
	if t != nil {
		return nil, ErrNotFound
	}
	return torrent.ParseRaw(t)
}

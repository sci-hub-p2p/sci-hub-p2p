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

// Package torrent parse raw torrent file and generate info hash.
package torrent

import "fmt"

type Node struct {
	host string
	port int
}

type Torrent struct {
	info
	Announce     string
	AnnounceList [][]string
	CreationDate int
	Nodes        []Node
	InfoHash     string
	Pieces       []byte
}

func (t Torrent) String() string {
	return fmt.Sprintf("Torrent{Name=%s, info_hash=%s}", t.Name, t.InfoHash)
}

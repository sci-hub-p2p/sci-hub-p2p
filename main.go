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

package main

import (
	"fmt"
	"os"

	"sci_hub_p2p/torrent"
)

func main() {
	const torrentPath = "tests/fixtures/sm_83500000-83599999.torrent"
	// content, err := os.ReadFile(torrentPath)
	// if err != nil {
	// 	return
	// }
	//
	// data, err := bencode1.Unmarshal(content)
	// if err != nil {
	// 	return
	// }

	file, err := os.Open(torrentPath)
	if err != nil {
		return
	}

	t, err := torrent.ParseReader(file)
	if err != nil {
		fmt.Println("error:", err)

		return
	}

	fmt.Println(t)
}

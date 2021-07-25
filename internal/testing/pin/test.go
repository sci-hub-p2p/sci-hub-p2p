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

// nolint
package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/itchio/lzma"

	"sci_hub_p2p/pkg/indexes"
	"sci_hub_p2p/pkg/logger"
)

func main() {
	f, err := os.Open("./out/d57b1013eee9138a8906bcd274d727b5d7e8a307.jsonlines.lzma")
	if err != nil {
		logger.Fatal(err)
	}
	defer f.Close()
	r := lzma.NewReader(f)
	defer r.Close()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		var s []string
		err = json.Unmarshal(scanner.Bytes(), &s)
		if err != nil || len(s) != 2 {
			logger.Fatal(err)
		}
		value, err := base64.StdEncoding.DecodeString(s[1])
		if err != nil {
			logger.Fatal(err)
		}

		key, err := url.QueryUnescape(strings.TrimSuffix(s[0], ".pdf"))
		if err != nil {
			logger.Fatal(err)
		}
		c, err := cid.Parse(indexes.LoadRecordV0(value).CID[:])
		if err != nil {
			logger.Fatal(err)
		}
		fmt.Println(key, c)
	}

	err = scanner.Err()
	if err != nil {
		logger.Fatal(err)
	}

}

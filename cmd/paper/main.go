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

package paper

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"

	"sci_hub_p2p/cmd/storage"
	"sci_hub_p2p/internal/client"
	"sci_hub_p2p/internal/torrent"
	"sci_hub_p2p/internal/utils"
	"sci_hub_p2p/pkg/consts"
	"sci_hub_p2p/pkg/vars"
)

var Cmd = &cobra.Command{
	Use:           "paper",
	SilenceErrors: false,
}

var fetchCmd = &cobra.Command{
	Use:           "fetch",
	Short:         "fetch a paper from p2p network",
	Example:       "paper fetch --doi '10.1145/1327452.1327492' -o map-reduce.pdf",
	SilenceErrors: false,
	PreRunE:       utils.EnsureDir(vars.GetAppTmpDir()),
	RunE: func(cmd *cobra.Command, args []string) error {
		if doi == "" {
			return errors.New("doi can't be empty string")
		}

		doi = strings.TrimSuffix(doi, ".pdf")
		r, err := storage.GetIndexRecord([]byte(doi))
		if err != nil {
			return err
		}

		tDB, err := bbolt.Open(vars.TorrentDBPath(), consts.DefaultFilePerm, bbolt.DefaultOptions)
		if err != nil {
			return errors.Wrap(err, "failed to open torrent database")
		}
		defer tDB.Close()

		var raw []byte
		err = tDB.View(func(tx *bbolt.Tx) error {
			raw = tx.Bucket(consts.TorrentBucket()).Get(r.InfoHash[:])

			return nil
		})
		if err != nil {
			return errors.Wrap(err, "failed to read from torrent DB")
		}
		if raw == nil {
			return errors.Wrap(err, "failed to find torrent in DB")
		}

		t, err := torrent.ParseRaw(raw)
		if err != nil {
			return errors.Wrapf(err, "failed to parse torrent from DB, please reload torrent %s it into database",
				r.HexInfoHash())
		}

		p, err := r.Build(doi, t)
		if err != nil {
			return err
		}

		b, err := client.Fetch(p, raw)
		if err != nil {
			return err
		}
		err = os.WriteFile(out, b, consts.DefaultFilePerm)

		return err
	},
}

var doi string
var out string

func init() {
	Cmd.AddCommand(fetchCmd)

	fetchCmd.Flags().StringVar(&doi, "doi", "", "")
	fetchCmd.Flags().StringVarP(&out, "output", "o", "", "output file path")

	if err := utils.MarkFlagsRequired(fetchCmd, "doi", "output"); err != nil {
		panic(err)
	}
}

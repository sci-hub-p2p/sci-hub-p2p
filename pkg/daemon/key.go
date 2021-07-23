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

package daemon

import (
	"crypto/rsa"
	"encoding/pem"
	"os"
	"path/filepath"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/pnet"
	"github.com/pkg/errors"

	"sci_hub_p2p/pkg/constants"
	"sci_hub_p2p/pkg/key"
	"sci_hub_p2p/pkg/logger"
)

func pnetKey() (pnet.PSK, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Wrap(err, "failed to detect homedir")
	}
	var keyPath = filepath.Join(home, ".ipfs/swarm.key")
	r, err := os.Open(keyPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to read pnet key %s", keyPath)
	}
	defer r.Close()
	logger.Info("using pnet key")
	k, err := pnet.DecodeV1PSK(r)

	return k, errors.Wrap(err, "failed to decode pnet KEY")
}

func genKey() (crypto.PrivKey, error) {
	const keyPath = "./out/private.key"
	var raw, err = os.ReadFile(keyPath)
	if errors.Is(err, os.ErrNotExist) {
		logger.Info("Generating New Rsa Key")
		priv, _, err := crypto.GenerateKeyPair(crypto.RSA, constants.PrivateKeyLength)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate RSA key")
		}
		stdKey, err := crypto.PrivKeyToStdKey(priv)
		if err != nil {
			return nil, errors.Wrap(err, "failed to convert libp2p key to std key")
		}
		v, ok := stdKey.(*rsa.PrivateKey)
		if !ok {
			panic("can't cast private key to *rsa.PrivateKey")
		}
		raw = key.ExportRsaPrivateKeyAsPem(v)
		err = os.WriteFile(keyPath, raw, constants.SecurityPerm)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to save key to file %s", keyPath)
		}

		return priv, nil
	}

	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}
	k, err := crypto.UnmarshalRsaPrivateKey(block.Bytes)

	return k, errors.Wrap(err, "filed to parse encode keyfile content")
}

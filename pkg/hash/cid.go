// Copyright 2021 Trim21 <trim21.me@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
package hash

import (
	"io"

	"github.com/ipfs/go-cid"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/pkg/errors"

	"sci_hub_p2p/internal/memorydag"
	"sci_hub_p2p/pkg/storage"
)

func Black2dBalancedSized256K(r io.Reader) ([]byte, error) {
	var c, err = Cid(r)
	if err != nil {
		return nil, errors.Wrap(err, "can't generate cid")
	}

	return c.Bytes(), nil
}

func Cid(r io.Reader) (cid.Cid, error) {
	var n, err = addFile(r)
	if err != nil {
		return cid.Cid{}, errors.Wrap(err, "can't generate cid")
	}

	return n.Cid(), nil
}

func addFile(r io.Reader) (ipld.Node, error) {
	n, err := storage.Add(memorydag.New(), r)

	return n, errors.Wrap(err, "failed to generate DAG from reader")
}

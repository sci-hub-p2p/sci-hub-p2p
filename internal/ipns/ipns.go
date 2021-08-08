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

package ipns

import (
	"context"
	"strings"

	"github.com/ipfs/go-namesys"
	"github.com/ipfs/go-path"
	"github.com/pkg/errors"
)

// ErrNoNS is an explicit error for when an IPFS node doesn't
// (yet) have a name system.
var ErrNoNS = errors.New("core/resolve: no Namesys on IpfsNode - can't resolve ipns entry")
var ErrIpns = errors.New("invalid ipns")

// ResolveIPNS resolves /ipns paths.
func ResolveIPNS(ctx context.Context, ns namesys.Resolver, p path.Path) (path.Path, error) {
	if !strings.HasPrefix(p.String(), "/ipns/") {
		return p, nil
	}

	// /ipns/<hash>

	if ns == nil {
		return "", ErrNoNS
	}

	seg := p.Segments()

	if len(seg) < 2 || seg[1] == "" { // just "/<protocol/>" without further segments
		return "", errors.Wrapf(ErrIpns, "missing IPNS ID %q", p)
	}

	extensions := seg[2:]

	resolvable, err := path.FromSegments("/", seg[0], seg[1])
	if err != nil {
		return "", errors.Wrap(err, "failed to create new path")
	}

	respath, err := ns.Resolve(ctx, resolvable.String())
	if err != nil {
		return "", errors.Wrap(err, "failed to resolve")
	}

	segments := append(respath.Segments(), extensions...)

	p, err = path.FromSegments("/", segments...)
	if err != nil {
		return "", errors.Wrap(err, "failed to create new path")
	}

	return p, nil
}

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

package dagserv

// customized layout to store piece offset in metadata db.

import (
	"errors"

	ipld "github.com/ipfs/go-ipld-format"
	ft "github.com/ipfs/go-unixfs"
	"github.com/ipfs/go-unixfs/importer/helpers"
)

func BalanceLayout(db *helpers.DagBuilderHelper) (ipld.Node, error) {
	if db.Done() {
		// No data, return just an empty node.
		root, err := db.NewLeafNode(nil, ft.TFile)
		if err != nil {
			return nil, err
		}

		return root, db.Add(root)
	}

	// The first `root` will be a single leaf node with data
	// (corner case), after that subsequent `root` nodes will
	// always be internal nodes (with a depth > 0) that can
	// be handled by the loop.
	root, fileSize, err := db.NewLeafDataNode(ft.TFile)
	if err != nil {
		return nil, err
	}

	// Each time a DAG of a certain `depth` is filled (because it
	// has reached its maximum capacity of `db.Maxlinks()` per node)
	// extend it by making it a sub-DAG of a bigger DAG with `depth+1`.
	for depth := 1; !db.Done(); depth++ {
		// Add the old `root` as a child of the `newRoot`.
		newRoot := db.NewFSNodeOverDag(ft.TFile)
		newRoot.AddChild(root, fileSize, db)

		// Fill the `newRoot` (that has the old `root` already as child)
		// and make it the current `root` for the next iteration (when
		// it will become "old").
		root, fileSize, err = fillNodeRec(db, newRoot, depth)
		if err != nil {
			return nil, err
		}
	}

	return root, db.Add(root)
}

func fillNodeRec(db *helpers.DagBuilderHelper, node *helpers.FSNodeOverDag, depth int) (filledNode ipld.Node, nodeFileSize uint64, err error) {
	if depth < 1 {
		return nil, 0, errors.New("attempt to fillNode at depth < 1")
	}

	if node == nil {
		node = db.NewFSNodeOverDag(ft.TFile)
	}

	// Child node created on every iteration to add to parent `node`.
	// It can be a leaf node or another internal node.
	var childNode ipld.Node
	// File size from the child node needed to update the `FSNode`
	// in `node` when adding the child.
	var childFileSize uint64

	// While we have room and there is data available to be added.
	for node.NumChildren() < db.Maxlinks() && !db.Done() {
		if depth == 1 {
			// Base case: add leaf node with data.
			childNode, childFileSize, err = db.NewLeafDataNode(ft.TFile)
			if err != nil {
				return nil, 0, err
			}
		} else {
			// Recursion case: create an internal node to in turn keep
			// descending in the DAG and adding child nodes to it.
			childNode, childFileSize, err = fillNodeRec(db, nil, depth-1)
			if err != nil {
				return nil, 0, err
			}
		}

		err = node.AddChild(childNode, childFileSize, db)
		if err != nil {
			return nil, 0, err
		}
	}

	nodeFileSize = node.FileSize()

	// Get the final `dag.ProtoNode` with the `FSNode` data encoded inside.
	filledNode, err = node.Commit()
	if err != nil {
		return nil, 0, err
	}

	return filledNode, nodeFileSize, nil
}

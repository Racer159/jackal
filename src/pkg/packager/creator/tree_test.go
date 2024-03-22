// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package creator contains functions for creating Zarf packages.
package creator

import (
	"fmt"
	"testing"

	"github.com/defenseunicorns/zarf/src/types"
)

// TestNewNode tests the creation of a new node
func TestNewNode(t *testing.T) {

	// Start with one component, the parent
	// two components below that, the parent will have edges to those
	// Both of these components will have an edge to a sub-component

	root := types.ZarfComponent{Name: "k3s",
		Import: types.ZarfComponentImport{Path: "packages/distros/k3s"}}

	amdK3s := types.ZarfComponent{Name: "k3s",
		Import: types.ZarfComponentImport{Path: "common", ComponentName: "k3s"}}

	armK3s := types.ZarfComponent{Name: "k3s",
		Import: types.ZarfComponentImport{Path: "common", ComponentName: "k3s"}}

	commonK3s := types.ZarfComponent{Name: "k3s",
		Import: types.ZarfComponentImport{}}

	node := NewNode(root)
	amdNode := NewNode(amdK3s)
	armNode := NewNode(armK3s)
	commonNode := NewNode(commonK3s)

	node.addEdge(armNode)
	node.addEdge(amdNode)
	amdNode.addEdge(commonNode)
	armNode.addEdge(commonNode)

	fmt.Printf("node %+v", node)

}

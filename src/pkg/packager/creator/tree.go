// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package creator contains functions for creating Zarf packages.
package creator

import "github.com/defenseunicorns/zarf/src/types"

// Define a DAGNode struct with a value and a slice of pointers to other Nodes (edges)
type DAGNode struct {
	types.ZarfComponent
	edges []*DAGNode
}

// Create a new Node
func NewNode(value types.ZarfComponent) *DAGNode {
	return &DAGNode{ZarfComponent: value}
}

// Add an edge from one node to another
func (n *DAGNode) addEdge(target *DAGNode) {
	n.edges = append(n.edges, target)
}

// Define a DAG struct that holds a slice of all nodes for convenience
type DAG struct {
	nodes []*DAGNode
}

// Create a new DAG
func NewDAG() *DAG {
	return &DAG{}
}

// Add a new node to the DAG
func (d *DAG) AddNode(node *DAGNode) {
	d.nodes = append(d.nodes, node)
}

// Add an edge between nodes in the DAG
func (d *DAG) AddEdge(from, to *DAGNode) {
	from.addEdge(to)
}

func NewImportTree() {
	// Start with all with components in the package
	// Each component is going to branch out

}

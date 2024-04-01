// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package composer contains functions for composing components within Jackal packages.
package composer

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/defenseunicorns/pkg/helpers"
	"github.com/racer159/jackal/src/internal/packager/validate"
	"github.com/racer159/jackal/src/pkg/layout"
	"github.com/racer159/jackal/src/pkg/packager/deprecated"
	"github.com/racer159/jackal/src/pkg/utils"
	"github.com/racer159/jackal/src/pkg/zoci"
	"github.com/racer159/jackal/src/types"
)

// Node is a node in the import chain
type Node struct {
	types.JackalComponent

	index int

	vars   []types.JackalPackageVariable
	consts []types.JackalPackageConstant

	relativeToHead      string
	originalPackageName string

	prev *Node
	next *Node
}

// Index returns the .components index location for this node's source `jackal.yaml`
func (n *Node) Index() int {
	return n.index
}

// OriginalPackageName returns the .metadata.name for this node's source `jackal.yaml`
func (n *Node) OriginalPackageName() string {
	return n.originalPackageName
}

// ImportLocation gets the path from the base `jackal.yaml` to the imported `jackal.yaml`
func (n *Node) ImportLocation() string {
	if n.prev != nil {
		if n.prev.JackalComponent.Import.URL != "" {
			return n.prev.JackalComponent.Import.URL
		}
	}
	return n.relativeToHead
}

// Next returns next node in the chain
func (n *Node) Next() *Node {
	return n.next
}

// Prev returns previous node in the chain
func (n *Node) Prev() *Node {
	return n.prev
}

// ImportName returns the name of the component to import
// If the component import has a ComponentName defined, that will be used
// otherwise the name of the component will be used
func (n *Node) ImportName() string {
	name := n.JackalComponent.Name
	if n.Import.ComponentName != "" {
		name = n.Import.ComponentName
	}
	return name
}

// ImportChain is a doubly linked list of component import definitions
type ImportChain struct {
	head *Node
	tail *Node

	remote *zoci.Remote
}

// Head returns the first node in the import chain
func (ic *ImportChain) Head() *Node {
	return ic.head
}

// Tail returns the last node in the import chain
func (ic *ImportChain) Tail() *Node {
	return ic.tail
}

func (ic *ImportChain) append(c types.JackalComponent, index int, originalPackageName string,
	relativeToHead string, vars []types.JackalPackageVariable, consts []types.JackalPackageConstant) {
	node := &Node{
		JackalComponent:     c,
		index:               index,
		originalPackageName: originalPackageName,
		relativeToHead:      relativeToHead,
		vars:                vars,
		consts:              consts,
		prev:                nil,
		next:                nil,
	}
	if ic.head == nil {
		ic.head = node
		ic.tail = node
	} else {
		p := ic.tail
		node.prev = p
		p.next = node
		ic.tail = node
	}
}

// NewImportChain creates a new import chain from a component
// Returning the chain on error so we can have additional information to use during lint
func NewImportChain(head types.JackalComponent, index int, originalPackageName, arch, flavor string) (*ImportChain, error) {
	ic := &ImportChain{}
	if arch == "" {
		return ic, fmt.Errorf("cannot build import chain: architecture must be provided")
	}

	ic.append(head, index, originalPackageName, ".", nil, nil)

	history := []string{}

	node := ic.head
	for node != nil {
		isLocal := node.Import.Path != ""
		isRemote := node.Import.URL != ""

		if !isLocal && !isRemote {
			// This is the end of the import chain,
			// as the current node/component is not importing anything
			return ic, nil
		}

		// TODO: stuff like this should also happen in linting
		if err := validate.ImportDefinition(&node.JackalComponent); err != nil {
			return ic, err
		}

		// ensure that remote components are not importing other remote components
		if node.prev != nil && node.prev.Import.URL != "" && isRemote {
			return ic, fmt.Errorf("detected malformed import chain, cannot import remote components from remote components")
		}
		// ensure that remote components are not importing local components
		if node.prev != nil && node.prev.Import.URL != "" && isLocal {
			return ic, fmt.Errorf("detected malformed import chain, cannot import local components from remote components")
		}

		var pkg types.JackalPackage

		var relativeToHead string
		var importURL string
		if isLocal {
			history = append(history, node.Import.Path)
			relativeToHead = filepath.Join(history...)

			// prevent circular imports (including self-imports)
			// this is O(n^2) but the import chain should be small
			prev := node
			for prev != nil {
				if prev.relativeToHead == relativeToHead {
					return ic, fmt.Errorf("detected circular import chain: %s", strings.Join(history, " -> "))
				}
				prev = prev.prev
			}

			// this assumes the composed package is following the jackal layout
			if err := utils.ReadYaml(filepath.Join(relativeToHead, layout.JackalYAML), &pkg); err != nil {
				return ic, err
			}
		} else if isRemote {
			importURL = node.Import.URL
			remote, err := ic.getRemote(node.Import.URL)
			if err != nil {
				return ic, err
			}
			pkg, err = remote.FetchJackalYAML(context.TODO())
			if err != nil {
				return ic, err
			}
		}

		name := node.ImportName()

		// 'found' and 'index' are parallel slices. Each element in found[x] corresponds to pkg[index[x]]
		// found[0] and pkg[index[0]] would be the same component for example
		found := []types.JackalComponent{}
		index := []int{}
		for i, component := range pkg.Components {
			if component.Name == name && CompatibleComponent(component, arch, flavor) {
				found = append(found, component)
				index = append(index, i)
			}
		}

		if len(found) == 0 {
			componentNotFound := "component %q not found in %q"
			if isLocal {
				return ic, fmt.Errorf(componentNotFound, name, relativeToHead)
			} else if isRemote {
				return ic, fmt.Errorf(componentNotFound, name, importURL)
			}
		} else if len(found) > 1 {
			multipleComponentsFound := "multiple components named %q found in %q satisfying %q"
			if isLocal {
				return ic, fmt.Errorf(multipleComponentsFound, name, relativeToHead, arch)
			} else if isRemote {
				return ic, fmt.Errorf(multipleComponentsFound, name, importURL, arch)
			}
		}

		ic.append(found[0], index[0], pkg.Metadata.Name, relativeToHead, pkg.Variables, pkg.Constants)
		node = node.next
	}
	return ic, nil
}

// String returns a string representation of the import chain
func (ic *ImportChain) String() string {
	if ic.head.next == nil {
		return fmt.Sprintf("component %q imports nothing", ic.head.Name)
	}

	s := strings.Builder{}

	name := ic.head.ImportName()

	if ic.head.Import.Path != "" {
		s.WriteString(fmt.Sprintf("component %q imports %q in %s", ic.head.Name, name, ic.head.Import.Path))
	} else {
		s.WriteString(fmt.Sprintf("component %q imports %q in %s", ic.head.Name, name, ic.head.Import.URL))
	}

	node := ic.head.next
	for node != ic.tail {
		name := node.ImportName()
		s.WriteString(", which imports ")
		if node.Import.Path != "" {
			s.WriteString(fmt.Sprintf("%q in %s", name, node.Import.Path))
		} else {
			s.WriteString(fmt.Sprintf("%q in %s", name, node.Import.URL))
		}

		node = node.next
	}

	return s.String()
}

// Migrate performs migrations on the import chain
func (ic *ImportChain) Migrate(build types.JackalBuildData) (warnings []string) {
	node := ic.head
	for node != nil {
		migrated, w := deprecated.MigrateComponent(build, node.JackalComponent)
		node.JackalComponent = migrated
		warnings = append(warnings, w...)
		node = node.next
	}
	if len(warnings) > 0 {
		final := fmt.Sprintf("Migrations were performed on the import chain of: %q", ic.head.Name)
		warnings = append(warnings, final)
	}
	return warnings
}

// Compose merges the import chain into a single component
// fixing paths, overriding metadata, etc
func (ic *ImportChain) Compose() (composed *types.JackalComponent, err error) {
	composed = &ic.tail.JackalComponent

	if ic.tail.prev == nil {
		// only had one component in the import chain
		return composed, nil
	}

	if err := ic.fetchOCISkeleton(); err != nil {
		return nil, err
	}

	// start with an empty component to compose into
	composed = &types.JackalComponent{}

	// start overriding with the tail node
	node := ic.tail
	for node != nil {
		fixPaths(&node.JackalComponent, node.relativeToHead)

		// perform overrides here
		err := overrideMetadata(composed, node.JackalComponent)
		if err != nil {
			return nil, err
		}

		overrideDeprecated(composed, node.JackalComponent)
		overrideResources(composed, node.JackalComponent)
		overrideActions(composed, node.JackalComponent)

		composeExtensions(composed, node.JackalComponent, node.relativeToHead)

		node = node.prev
	}

	return composed, nil
}

// MergeVariables merges variables from the import chain
func (ic *ImportChain) MergeVariables(existing []types.JackalPackageVariable) (merged []types.JackalPackageVariable) {
	exists := func(v1 types.JackalPackageVariable, v2 types.JackalPackageVariable) bool {
		return v1.Name == v2.Name
	}

	node := ic.tail
	for node != nil {
		// merge the vars
		merged = helpers.MergeSlices(node.vars, merged, exists)
		node = node.prev
	}
	merged = helpers.MergeSlices(existing, merged, exists)

	return merged
}

// MergeConstants merges constants from the import chain
func (ic *ImportChain) MergeConstants(existing []types.JackalPackageConstant) (merged []types.JackalPackageConstant) {
	exists := func(c1 types.JackalPackageConstant, c2 types.JackalPackageConstant) bool {
		return c1.Name == c2.Name
	}

	node := ic.tail
	for node != nil {
		// merge the consts
		merged = helpers.MergeSlices(node.consts, merged, exists)
		node = node.prev
	}
	merged = helpers.MergeSlices(existing, merged, exists)

	return merged
}

// CompatibleComponent determines if this component is compatible with the given create options
func CompatibleComponent(c types.JackalComponent, arch, flavor string) bool {
	satisfiesArch := c.Only.Cluster.Architecture == "" || c.Only.Cluster.Architecture == arch
	satisfiesFlavor := c.Only.Flavor == "" || c.Only.Flavor == flavor
	return satisfiesArch && satisfiesFlavor
}

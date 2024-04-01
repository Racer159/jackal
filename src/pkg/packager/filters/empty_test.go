// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package filters contains core implementations of the ComponentFilterStrategy interface.
package filters

import (
	"testing"

	"github.com/defenseunicorns/jackal/src/types"
	"github.com/stretchr/testify/require"
)

func TestEmptyFilter_Apply(t *testing.T) {
	components := []types.JackalComponent{
		{
			Name: "component1",
		},
		{
			Name: "component2",
		},
	}
	pkg := types.JackalPackage{
		Components: components,
	}
	filter := Empty()

	result, err := filter.Apply(pkg)

	require.NoError(t, err)
	require.Equal(t, components, result)
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package filters contains core implementations of the ComponentFilterStrategy interface.
package filters

import (
	"testing"

	"github.com/racer159/jackal/src/types"
	"github.com/stretchr/testify/require"
)

func TestCombine(t *testing.T) {
	f1 := BySelectState("*a*")
	f2 := BySelectState("*bar, foo")
	f3 := Empty()

	combo := Combine(f1, f2, f3)

	pkg := types.JackalPackage{
		Components: []types.JackalComponent{
			{
				Name: "foo",
			},
			{
				Name: "bar",
			},
			{
				Name: "baz",
			},
			{
				Name: "foobar",
			},
		},
	}

	expected := []types.JackalComponent{
		{
			Name: "bar",
		},
		{
			Name: "foobar",
		},
	}

	result, err := combo.Apply(pkg)
	require.NoError(t, err)
	require.Equal(t, expected, result)

	// Test error propagation
	combo = Combine(f1, f2, ForDeploy("group with no default", false))
	pkg.Components = append(pkg.Components, types.JackalComponent{
		Name:            "group with no default",
		DeprecatedGroup: "g1",
	})
	_, err = combo.Apply(pkg)
	require.Error(t, err)
}

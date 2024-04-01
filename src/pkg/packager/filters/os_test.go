// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package filters contains core implementations of the ComponentFilterStrategy interface.
package filters

import (
	"testing"

	"github.com/racer159/jackal/src/internal/packager/validate"
	"github.com/racer159/jackal/src/types"
	"github.com/stretchr/testify/require"
)

func TestLocalOSFilter(t *testing.T) {

	pkg := types.JackalPackage{}
	for _, os := range validate.SupportedOS() {
		pkg.Components = append(pkg.Components, types.JackalComponent{
			Only: types.JackalComponentOnlyTarget{
				LocalOS: os,
			},
		})
	}

	for _, os := range validate.SupportedOS() {
		filter := ByLocalOS(os)
		result, err := filter.Apply(pkg)
		if os == "" {
			require.ErrorIs(t, err, ErrLocalOSRequired)
		} else {
			require.NoError(t, err)
		}
		for _, component := range result {
			if component.Only.LocalOS != "" {
				require.Equal(t, os, component.Only.LocalOS)
			}
		}
	}
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package composer contains functions for composing components within Jackal packages.
package composer

import (
	"github.com/Racer159/jackal/src/extensions/bigbang"
	"github.com/Racer159/jackal/src/types"
)

func composeExtensions(c *types.JackalComponent, override types.JackalComponent, relativeTo string) {
	bigbang.Compose(c, override, relativeTo)
}

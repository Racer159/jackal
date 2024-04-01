// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package main is the entrypoint for the Jackal binary.
package main

import (
	"embed"

	"github.com/Racer159/jackal/src/cmd"
	"github.com/Racer159/jackal/src/config"
	"github.com/Racer159/jackal/src/pkg/packager/lint"
)

//go:embed cosign.pub
var cosignPublicKey string

//go:embed jackal.schema.json
var jackalSchema embed.FS

func main() {
	config.CosignPublicKey = cosignPublicKey
	lint.JackalSchema = jackalSchema
	cmd.Execute()
}

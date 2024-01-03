// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package main is the entrypoint for the Zarf binary.
package main

import (
	"embed"

	"github.com/defenseunicorns/zarf/src/cmd"
	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/packager/lint"
)

//go:embed cosign.pub
var cosignPublicKey string

//go:embed zarf.schema.json
var zarfSchema embed.FS

func main() {
	config.CosignPublicKey = cosignPublicKey
	lint.ZarfSchema = zarfSchema
	cmd.Execute()
}

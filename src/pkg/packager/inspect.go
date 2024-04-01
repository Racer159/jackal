// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package packager contains functions for interacting with, managing and deploying Jackal packages.
package packager

import (
	"github.com/racer159/jackal/src/internal/packager/sbom"
	"github.com/racer159/jackal/src/pkg/utils"
)

// Inspect list the contents of a package.
func (p *Packager) Inspect() (err error) {
	wantSBOM := p.cfg.InspectOpts.ViewSBOM || p.cfg.InspectOpts.SBOMOutputDir != ""

	p.cfg.Pkg, p.warnings, err = p.source.LoadPackageMetadata(p.layout, wantSBOM, true)
	if err != nil {
		return err
	}

	utils.ColorPrintYAML(p.cfg.Pkg, nil, false)

	sbomDir := p.layout.SBOMs.Path

	if p.cfg.InspectOpts.SBOMOutputDir != "" {
		out, err := p.layout.SBOMs.OutputSBOMFiles(p.cfg.InspectOpts.SBOMOutputDir, p.cfg.Pkg.Metadata.Name)
		if err != nil {
			return err
		}
		sbomDir = out
	}

	if p.cfg.InspectOpts.ViewSBOM {
		sbom.ViewSBOMFiles(sbomDir)
	}

	return nil
}

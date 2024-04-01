// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package packager contains functions for interacting with, managing and deploying Jackal packages.
package packager

import (
	"fmt"
	"os"

	"github.com/defenseunicorns/jackal/src/config"
	"github.com/defenseunicorns/jackal/src/internal/packager/validate"
	"github.com/defenseunicorns/jackal/src/pkg/layout"
	"github.com/defenseunicorns/jackal/src/pkg/message"
	"github.com/defenseunicorns/jackal/src/pkg/packager/creator"
	"github.com/defenseunicorns/pkg/helpers"
)

// Create generates a Jackal package tarball for a given PackageConfig and optional base directory.
func (p *Packager) Create() (err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := os.Chdir(p.cfg.CreateOpts.BaseDir); err != nil {
		return fmt.Errorf("unable to access directory %q: %w", p.cfg.CreateOpts.BaseDir, err)
	}

	message.Note(fmt.Sprintf("Using build directory %s", p.cfg.CreateOpts.BaseDir))

	pc := creator.NewPackageCreator(p.cfg.CreateOpts, p.cfg, cwd)

	if err := helpers.CreatePathAndCopy(layout.JackalYAML, p.layout.JackalYAML); err != nil {
		return err
	}

	p.cfg.Pkg, p.warnings, err = pc.LoadPackageDefinition(p.layout)
	if err != nil {
		return err
	}

	// Perform early package validation.
	if err := validate.Run(p.cfg.Pkg); err != nil {
		return fmt.Errorf("unable to validate package: %w", err)
	}

	if !p.confirmAction(config.JackalCreateStage) {
		return fmt.Errorf("package creation canceled")
	}

	if err := pc.Assemble(p.layout, p.cfg.Pkg.Components, p.cfg.Pkg.Metadata.Architecture); err != nil {
		return err
	}

	// cd back for output
	if err := os.Chdir(cwd); err != nil {
		return err
	}

	return pc.Output(p.layout, &p.cfg.Pkg)
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package packager contains functions for interacting with, managing and deploying Jackal packages.
package packager

import (
	"fmt"
	"os"
	"runtime"

	"github.com/Racer159/jackal/src/config"
	"github.com/Racer159/jackal/src/internal/packager/validate"
	"github.com/Racer159/jackal/src/pkg/layout"
	"github.com/Racer159/jackal/src/pkg/message"
	"github.com/Racer159/jackal/src/pkg/packager/creator"
	"github.com/Racer159/jackal/src/pkg/packager/filters"
	"github.com/Racer159/jackal/src/pkg/packager/variables"
	"github.com/Racer159/jackal/src/types"
	"github.com/defenseunicorns/pkg/helpers"
)

// DevDeploy creates + deploys a package in one shot
func (p *Packager) DevDeploy() error {
	config.CommonOptions.Confirm = true
	p.cfg.CreateOpts.SkipSBOM = !p.cfg.CreateOpts.NoYOLO

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	if err := os.Chdir(p.cfg.CreateOpts.BaseDir); err != nil {
		return fmt.Errorf("unable to access directory %q: %w", p.cfg.CreateOpts.BaseDir, err)
	}

	pc := creator.NewPackageCreator(p.cfg.CreateOpts, p.cfg, cwd)

	if err := helpers.CreatePathAndCopy(layout.JackalYAML, p.layout.JackalYAML); err != nil {
		return err
	}

	p.cfg.Pkg, p.warnings, err = pc.LoadPackageDefinition(p.layout)
	if err != nil {
		return err
	}

	filter := filters.Combine(
		filters.ByLocalOS(runtime.GOOS),
		filters.ForDeploy(p.cfg.PkgOpts.OptionalComponents, false),
	)
	p.cfg.Pkg.Components, err = filter.Apply(p.cfg.Pkg)
	if err != nil {
		return err
	}

	if err := validate.Run(p.cfg.Pkg); err != nil {
		return fmt.Errorf("unable to validate package: %w", err)
	}

	if err := variables.SetVariableMapInConfig(p.cfg); err != nil {
		return err
	}

	// If building in yolo mode, strip out all images and repos
	if !p.cfg.CreateOpts.NoYOLO {
		for idx := range p.cfg.Pkg.Components {
			p.cfg.Pkg.Components[idx].Images = []string{}
			p.cfg.Pkg.Components[idx].Repos = []string{}
		}
	}

	if err := pc.Assemble(p.layout, p.cfg.Pkg.Components, p.cfg.Pkg.Metadata.Architecture); err != nil {
		return err
	}

	message.HeaderInfof("📦 PACKAGE DEPLOY %s", p.cfg.Pkg.Metadata.Name)

	p.connectStrings = make(types.ConnectStrings)

	if !p.cfg.CreateOpts.NoYOLO {
		p.cfg.Pkg.Metadata.YOLO = true
	} else {
		p.hpaModified = false
		// Reset registry HPA scale down whether an error occurs or not
		defer p.resetRegistryHPA()
	}

	// Get a list of all the components we are deploying and actually deploy them
	deployedComponents, err := p.deployComponents()
	if err != nil {
		return err
	}
	if len(deployedComponents) == 0 {
		message.Warn("No components were selected for deployment.  Inspect the package to view the available components and select components interactively or by name with \"--components\"")
	}

	// Notify all the things about the successful deployment
	message.Successf("Jackal dev deployment complete")

	message.HorizontalRule()
	message.Title("Next steps:", "")

	message.JackalCommand("package inspect %s", p.cfg.Pkg.Metadata.Name)

	// cd back
	return os.Chdir(cwd)
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package packager contains functions for interacting with, managing and deploying Jackal packages.
package packager

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/Racer159/jackal/src/config"
	"github.com/Racer159/jackal/src/pkg/message"
	"github.com/Racer159/jackal/src/pkg/packager/filters"
	"github.com/Racer159/jackal/src/types"
)

// Mirror pulls resources from a package (images, git repositories, etc) and pushes them to remotes in the air gap without deploying them
func (p *Packager) Mirror() (err error) {
	filter := filters.Combine(
		filters.ByLocalOS(runtime.GOOS),
		filters.BySelectState(p.cfg.PkgOpts.OptionalComponents),
	)

	p.cfg.Pkg, p.warnings, err = p.source.LoadPackage(p.layout, filter, true)
	if err != nil {
		return fmt.Errorf("unable to load the package: %w", err)
	}

	var sbomWarnings []string
	p.sbomViewFiles, sbomWarnings, err = p.layout.SBOMs.StageSBOMViewFiles()
	if err != nil {
		return err
	}

	p.warnings = append(p.warnings, sbomWarnings...)

	// Confirm the overall package mirror
	if !p.confirmAction(config.JackalMirrorStage) {
		return fmt.Errorf("mirror cancelled")
	}

	p.cfg.State = &types.JackalState{
		RegistryInfo: p.cfg.InitOpts.RegistryInfo,
		GitServer:    p.cfg.InitOpts.GitServer,
	}

	for _, component := range p.cfg.Pkg.Components {
		if err := p.mirrorComponent(component); err != nil {
			return err
		}
	}
	return nil
}

// mirrorComponent mirrors a Jackal Component.
func (p *Packager) mirrorComponent(component types.JackalComponent) error {
	componentPaths := p.layout.Components.Dirs[component.Name]

	// All components now require a name
	message.HeaderInfof("📦 %s COMPONENT", strings.ToUpper(component.Name))

	hasImages := len(component.Images) > 0
	hasRepos := len(component.Repos) > 0

	if hasImages {
		if err := p.pushImagesToRegistry(component.Images, p.cfg.MirrorOpts.NoImgChecksum); err != nil {
			return fmt.Errorf("unable to push images to the registry: %w", err)
		}
	}

	if hasRepos {
		if err := p.pushReposToRepository(componentPaths.Repos, component.Repos); err != nil {
			return fmt.Errorf("unable to push the repos to the repository: %w", err)
		}
	}

	return nil
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package packager contains functions for interacting with, managing and deploying Jackal packages.
package packager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/defenseunicorns/pkg/helpers"
	goyaml "github.com/goccy/go-yaml"
	"github.com/racer159/jackal/src/config"
	"github.com/racer159/jackal/src/internal/packager/validate"
	"github.com/racer159/jackal/src/pkg/layout"
	"github.com/racer159/jackal/src/pkg/message"
	"github.com/racer159/jackal/src/types"
)

// Generate generates a Jackal package definition.
func (p *Packager) Generate() (err error) {
	generatedJackalYAMLPath := filepath.Join(p.cfg.GenerateOpts.Output, layout.JackalYAML)
	spinner := message.NewProgressSpinner("Generating package for %q at %s", p.cfg.GenerateOpts.Name, generatedJackalYAMLPath)

	if !helpers.InvalidPath(generatedJackalYAMLPath) {
		prefixed := filepath.Join(p.cfg.GenerateOpts.Output, fmt.Sprintf("%s-%s", p.cfg.GenerateOpts.Name, layout.JackalYAML))

		message.Warnf("%s already exists, writing to %s", generatedJackalYAMLPath, prefixed)

		generatedJackalYAMLPath = prefixed

		if !helpers.InvalidPath(generatedJackalYAMLPath) {
			return fmt.Errorf("unable to generate package, %s already exists", generatedJackalYAMLPath)
		}
	}

	generatedComponent := types.JackalComponent{
		Name:     p.cfg.GenerateOpts.Name,
		Required: helpers.BoolPtr(true),
		Charts: []types.JackalChart{
			{
				Name:      p.cfg.GenerateOpts.Name,
				Version:   p.cfg.GenerateOpts.Version,
				Namespace: p.cfg.GenerateOpts.Name,
				URL:       p.cfg.GenerateOpts.URL,
				GitPath:   p.cfg.GenerateOpts.GitPath,
			},
		},
	}

	p.cfg.Pkg = types.JackalPackage{
		Kind: types.JackalPackageConfig,
		Metadata: types.JackalMetadata{
			Name:        p.cfg.GenerateOpts.Name,
			Version:     p.cfg.GenerateOpts.Version,
			Description: "auto-generated using `jackal dev generate`",
		},
		Components: []types.JackalComponent{
			generatedComponent,
		},
	}

	images, err := p.findImages()
	if err != nil {
		// purposefully not returning error here, as we can still generate the package without images
		message.Warnf("Unable to find images: %s", err.Error())
	}

	for i := range p.cfg.Pkg.Components {
		name := p.cfg.Pkg.Components[i].Name
		p.cfg.Pkg.Components[i].Images = images[name]
	}

	if err := validate.Run(p.cfg.Pkg); err != nil {
		return err
	}

	if err := helpers.CreateDirectory(p.cfg.GenerateOpts.Output, helpers.ReadExecuteAllWriteUser); err != nil {
		return err
	}

	b, err := goyaml.MarshalWithOptions(p.cfg.Pkg, goyaml.IndentSequence(true), goyaml.UseSingleQuote(false))
	if err != nil {
		return err
	}

	schemaComment := fmt.Sprintf("# yaml-language-server: $schema=https://raw.githubusercontent.com/%s/%s/jackal.schema.json", config.GithubProject, config.CLIVersion)
	content := schemaComment + "\n" + string(b)

	// lets space things out a bit
	content = strings.Replace(content, "kind:\n", "\nkind:\n", 1)
	content = strings.Replace(content, "metadata:\n", "\nmetadata:\n", 1)
	content = strings.Replace(content, "components:\n", "\ncomponents:\n", 1)

	spinner.Successf("Generated package for %q at %s", p.cfg.GenerateOpts.Name, generatedJackalYAMLPath)

	return os.WriteFile(generatedJackalYAMLPath, []byte(content), helpers.ReadAllWriteUser)
}

// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package creator contains functions for creating Jackal packages.
package creator

import (
	"fmt"

	"github.com/racer159/jackal/src/config"
	"github.com/racer159/jackal/src/config/lang"
	"github.com/racer159/jackal/src/pkg/interactive"
	"github.com/racer159/jackal/src/pkg/utils"
	"github.com/racer159/jackal/src/types"
)

// FillActiveTemplate merges user-specified variables into the configuration templates of a jackal.yaml.
func FillActiveTemplate(pkg types.JackalPackage, setVariables map[string]string) (types.JackalPackage, []string, error) {
	templateMap := map[string]string{}
	warnings := []string{}

	promptAndSetTemplate := func(templatePrefix string, deprecated bool) error {
		yamlTemplates, err := utils.FindYamlTemplates(&pkg, templatePrefix, "###")
		if err != nil {
			return err
		}

		for key := range yamlTemplates {
			if deprecated {
				warnings = append(warnings, fmt.Sprintf(lang.PkgValidateTemplateDeprecation, key, key, key))
			}

			_, present := setVariables[key]
			if !present && !config.CommonOptions.Confirm {
				setVal, err := interactive.PromptVariable(types.JackalPackageVariable{
					Name: key,
				})
				if err != nil {
					return err
				}
				setVariables[key] = setVal
			} else if !present {
				return fmt.Errorf("template %q must be '--set' when using the '--confirm' flag", key)
			}
		}

		for key, value := range setVariables {
			templateMap[fmt.Sprintf("%s%s###", templatePrefix, key)] = value
		}

		return nil
	}

	// update the component templates on the package
	if err := ReloadComponentTemplatesInPackage(&pkg); err != nil {
		return types.JackalPackage{}, nil, err
	}

	if err := promptAndSetTemplate(types.JackalPackageTemplatePrefix, false); err != nil {
		return types.JackalPackage{}, nil, err
	}
	// [DEPRECATION] Set the Package Variable syntax as well for backward compatibility
	if err := promptAndSetTemplate(types.JackalPackageVariablePrefix, true); err != nil {
		return types.JackalPackage{}, nil, err
	}

	// Add special variable for the current package architecture
	templateMap[types.JackalPackageArch] = pkg.Metadata.Architecture

	if err := utils.ReloadYamlTemplate(&pkg, templateMap); err != nil {
		return types.JackalPackage{}, nil, err
	}

	return pkg, warnings, nil
}

// ReloadComponentTemplate appends ###JACKAL_COMPONENT_NAME### for the component, assigns value, and reloads
// Any instance of ###JACKAL_COMPONENT_NAME### within a component will be replaced with that components name
func ReloadComponentTemplate(component *types.JackalComponent) error {
	mappings := map[string]string{}
	mappings[types.JackalComponentName] = component.Name
	err := utils.ReloadYamlTemplate(component, mappings)
	if err != nil {
		return err
	}
	return nil
}

// ReloadComponentTemplatesInPackage appends ###JACKAL_COMPONENT_NAME###  for each component, assigns value, and reloads
func ReloadComponentTemplatesInPackage(jackalPackage *types.JackalPackage) error {
	// iterate through components to and find all ###JACKAL_COMPONENT_NAME, assign to component Name and value
	for i := range jackalPackage.Components {
		if err := ReloadComponentTemplate(&jackalPackage.Components[i]); err != nil {
			return err
		}
	}

	return nil
}

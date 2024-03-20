// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package creator contains functions for creating Zarf packages.
package creator

import (
	"fmt"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/config/lang"
	"github.com/defenseunicorns/zarf/src/pkg/interactive"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/types"
)

// FillActiveTemplate merges user-specified variables into the configuration templates of a zarf.yaml.
func FillActiveTemplate(pkg types.ZarfPackage, setVariables map[string]string, warnings *message.Warnings) (types.ZarfPackage, error) {
	templateMap := map[string]string{}

	promptAndSetTemplate := func(templatePrefix string, deprecated bool) error {
		yamlTemplates, err := utils.FindYamlTemplates(&pkg, templatePrefix, "###")
		if err != nil {
			return err
		}

		for key := range yamlTemplates {
			if deprecated {
				warnings.Add(fmt.Sprintf(lang.PkgValidateTemplateDeprecation, key, key, key))
			}

			_, present := setVariables[key]
			if !present && !config.CommonOptions.Confirm {
				setVal, err := interactive.PromptVariable(types.ZarfPackageVariable{
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
		return types.ZarfPackage{}, err
	}

	if err := promptAndSetTemplate(types.ZarfPackageTemplatePrefix, false); err != nil {
		return types.ZarfPackage{}, err
	}
	// [DEPRECATION] Set the Package Variable syntax as well for backward compatibility
	if err := promptAndSetTemplate(types.ZarfPackageVariablePrefix, true); err != nil {
		return types.ZarfPackage{}, err
	}

	// Add special variable for the current package architecture
	templateMap[types.ZarfPackageArch] = pkg.Metadata.Architecture

	if err := utils.ReloadYamlTemplate(&pkg, templateMap); err != nil {
		return types.ZarfPackage{}, err
	}

	return pkg, nil
}

// ReloadComponentTemplate appends ###ZARF_COMPONENT_NAME### for the component, assigns value, and reloads
// Any instance of ###ZARF_COMPONENT_NAME### within a component will be replaced with that components name
func ReloadComponentTemplate(component *types.ZarfComponent) error {
	mappings := map[string]string{}
	mappings[types.ZarfComponentName] = component.Name
	err := utils.ReloadYamlTemplate(component, mappings)
	if err != nil {
		return err
	}
	return nil
}

// ReloadComponentTemplatesInPackage appends ###ZARF_COMPONENT_NAME###  for each component, assigns value, and reloads
func ReloadComponentTemplatesInPackage(zarfPackage *types.ZarfPackage) error {
	// iterate through components to and find all ###ZARF_COMPONENT_NAME, assign to component Name and value
	for i := range zarfPackage.Components {
		if err := ReloadComponentTemplate(&zarfPackage.Components[i]); err != nil {
			return err
		}
	}

	return nil
}

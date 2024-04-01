// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package lint contains functions for verifying jackal yaml files are valid
package lint

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Racer159/jackal/src/config"
	"github.com/Racer159/jackal/src/config/lang"
	"github.com/Racer159/jackal/src/pkg/layout"
	"github.com/Racer159/jackal/src/pkg/packager/composer"
	"github.com/Racer159/jackal/src/pkg/packager/creator"
	"github.com/Racer159/jackal/src/pkg/transform"
	"github.com/Racer159/jackal/src/pkg/utils"
	"github.com/Racer159/jackal/src/types"
	"github.com/defenseunicorns/pkg/helpers"
	"github.com/xeipuuv/gojsonschema"
)

// JackalSchema is exported so main.go can embed the schema file
var JackalSchema embed.FS

func getSchemaFile() ([]byte, error) {
	return JackalSchema.ReadFile("jackal.schema.json")
}

// Validate validates a jackal file against the jackal schema, returns *validator with warnings or errors if they exist
// along with an error if the validation itself failed
func Validate(createOpts types.JackalCreateOptions) (*Validator, error) {
	validator := Validator{}
	var err error

	if err := utils.ReadYaml(filepath.Join(createOpts.BaseDir, layout.JackalYAML), &validator.typedJackalPackage); err != nil {
		return nil, err
	}

	if err := utils.ReadYaml(filepath.Join(createOpts.BaseDir, layout.JackalYAML), &validator.untypedJackalPackage); err != nil {
		return nil, err
	}

	if err := os.Chdir(createOpts.BaseDir); err != nil {
		return nil, fmt.Errorf("unable to access directory '%s': %w", createOpts.BaseDir, err)
	}

	validator.baseDir = createOpts.BaseDir

	lintComponents(&validator, &createOpts)

	if validator.jsonSchema, err = getSchemaFile(); err != nil {
		return nil, err
	}

	if err = validateSchema(&validator); err != nil {
		return nil, err
	}

	return &validator, nil
}

func lintComponents(validator *Validator, createOpts *types.JackalCreateOptions) {
	for i, component := range validator.typedJackalPackage.Components {
		arch := config.GetArch(validator.typedJackalPackage.Metadata.Architecture)

		if !composer.CompatibleComponent(component, arch, createOpts.Flavor) {
			continue
		}

		chain, err := composer.NewImportChain(component, i, validator.typedJackalPackage.Metadata.Name, arch, createOpts.Flavor)
		baseComponent := chain.Head()

		var badImportYqPath string
		if baseComponent != nil {
			if baseComponent.Import.URL != "" {
				badImportYqPath = fmt.Sprintf(".components.[%d].import.url", i)
			}
			if baseComponent.Import.Path != "" {
				badImportYqPath = fmt.Sprintf(".components.[%d].import.path", i)
			}
		}
		if err != nil {
			validator.addError(validatorMessage{
				description:    err.Error(),
				packageRelPath: ".",
				packageName:    validator.typedJackalPackage.Metadata.Name,
				yqPath:         badImportYqPath,
			})
		}

		node := baseComponent
		for node != nil {
			checkForVarInComponentImport(validator, node)
			fillComponentTemplate(validator, node, createOpts)
			lintComponent(validator, node)
			node = node.Next()
		}
	}
}

func fillComponentTemplate(validator *Validator, node *composer.Node, createOpts *types.JackalCreateOptions) {
	err := creator.ReloadComponentTemplate(&node.JackalComponent)
	if err != nil {
		validator.addWarning(validatorMessage{
			description:    err.Error(),
			packageRelPath: node.ImportLocation(),
			packageName:    node.OriginalPackageName(),
		})
	}
	templateMap := map[string]string{}

	setVarsAndWarn := func(templatePrefix string, deprecated bool) {
		yamlTemplates, err := utils.FindYamlTemplates(node, templatePrefix, "###")
		if err != nil {
			validator.addWarning(validatorMessage{
				description:    err.Error(),
				packageRelPath: node.ImportLocation(),
				packageName:    node.OriginalPackageName(),
			})
		}

		for key := range yamlTemplates {
			if deprecated {
				validator.addWarning(validatorMessage{
					description:    fmt.Sprintf(lang.PkgValidateTemplateDeprecation, key, key, key),
					packageRelPath: node.ImportLocation(),
					packageName:    node.OriginalPackageName(),
				})
			}
			_, present := createOpts.SetVariables[key]
			if !present {
				validator.addWarning(validatorMessage{
					description:    lang.UnsetVarLintWarning,
					packageRelPath: node.ImportLocation(),
					packageName:    node.OriginalPackageName(),
				})
			}
		}
		for key, value := range createOpts.SetVariables {
			templateMap[fmt.Sprintf("%s%s###", templatePrefix, key)] = value
		}
	}

	setVarsAndWarn(types.JackalPackageTemplatePrefix, false)

	// [DEPRECATION] Set the Package Variable syntax as well for backward compatibility
	setVarsAndWarn(types.JackalPackageVariablePrefix, true)

	utils.ReloadYamlTemplate(node, templateMap)
}

func isPinnedImage(image string) (bool, error) {
	transformedImage, err := transform.ParseImageRef(image)
	if err != nil {
		if strings.Contains(image, types.JackalPackageTemplatePrefix) ||
			strings.Contains(image, types.JackalPackageVariablePrefix) {
			return true, nil
		}
		return false, err
	}
	return (transformedImage.Digest != ""), err
}

func isPinnedRepo(repo string) bool {
	return (strings.Contains(repo, "@"))
}

func lintComponent(validator *Validator, node *composer.Node) {
	checkForUnpinnedRepos(validator, node)
	checkForUnpinnedImages(validator, node)
	checkForUnpinnedFiles(validator, node)
}

func checkForUnpinnedRepos(validator *Validator, node *composer.Node) {
	for j, repo := range node.Repos {
		repoYqPath := fmt.Sprintf(".components.[%d].repos.[%d]", node.Index(), j)
		if !isPinnedRepo(repo) {
			validator.addWarning(validatorMessage{
				yqPath:         repoYqPath,
				packageRelPath: node.ImportLocation(),
				packageName:    node.OriginalPackageName(),
				description:    "Unpinned repository",
				item:           repo,
			})
		}
	}
}

func checkForUnpinnedImages(validator *Validator, node *composer.Node) {
	for j, image := range node.Images {
		imageYqPath := fmt.Sprintf(".components.[%d].images.[%d]", node.Index(), j)
		pinnedImage, err := isPinnedImage(image)
		if err != nil {
			validator.addError(validatorMessage{
				yqPath:         imageYqPath,
				packageRelPath: node.ImportLocation(),
				packageName:    node.OriginalPackageName(),
				description:    "Invalid image reference",
				item:           image,
			})
			continue
		}
		if !pinnedImage {
			validator.addWarning(validatorMessage{
				yqPath:         imageYqPath,
				packageRelPath: node.ImportLocation(),
				packageName:    node.OriginalPackageName(),
				description:    "Image not pinned with digest",
				item:           image,
			})
		}
	}
}

func checkForUnpinnedFiles(validator *Validator, node *composer.Node) {
	for j, file := range node.Files {
		fileYqPath := fmt.Sprintf(".components.[%d].files.[%d]", node.Index(), j)
		if file.Shasum == "" && helpers.IsURL(file.Source) {
			validator.addWarning(validatorMessage{
				yqPath:         fileYqPath,
				packageRelPath: node.ImportLocation(),
				packageName:    node.OriginalPackageName(),
				description:    "No shasum for remote file",
				item:           file.Source,
			})
		}
	}
}

func checkForVarInComponentImport(validator *Validator, node *composer.Node) {
	if strings.Contains(node.Import.Path, types.JackalPackageTemplatePrefix) {
		validator.addWarning(validatorMessage{
			yqPath:         fmt.Sprintf(".components.[%d].import.path", node.Index()),
			packageRelPath: node.ImportLocation(),
			packageName:    node.OriginalPackageName(),
			description:    "Jackal does not evaluate variables at component.x.import.path",
			item:           node.Import.Path,
		})
	}
	if strings.Contains(node.Import.URL, types.JackalPackageTemplatePrefix) {
		validator.addWarning(validatorMessage{
			yqPath:         fmt.Sprintf(".components.[%d].import.url", node.Index()),
			packageRelPath: node.ImportLocation(),
			packageName:    node.OriginalPackageName(),
			description:    "Jackal does not evaluate variables at component.x.import.url",
			item:           node.Import.URL,
		})
	}
}

func makeFieldPathYqCompat(field string) string {
	if field == "(root)" {
		return field
	}
	// \b is a metacharacter that will stop at the next non-word character (including .)
	// https://regex101.com/r/pIRPk0/1
	re := regexp.MustCompile(`(\b\d+\b)`)

	wrappedField := re.ReplaceAllString(field, "[$1]")

	return fmt.Sprintf(".%s", wrappedField)
}

func validateSchema(validator *Validator) error {
	schemaLoader := gojsonschema.NewBytesLoader(validator.jsonSchema)
	documentLoader := gojsonschema.NewGoLoader(validator.untypedJackalPackage)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		for _, desc := range result.Errors() {
			validator.addError(validatorMessage{
				yqPath:         makeFieldPathYqCompat(desc.Field()),
				description:    desc.Description(),
				packageRelPath: ".",
				packageName:    validator.typedJackalPackage.Metadata.Name,
			})
		}
	}

	return err
}

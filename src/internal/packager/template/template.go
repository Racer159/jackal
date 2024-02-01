// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package template provides functions for templating yaml files.
package template

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/defenseunicorns/zarf/src/types"

	"github.com/defenseunicorns/zarf/src/config"
	"github.com/defenseunicorns/zarf/src/pkg/message"
	"github.com/defenseunicorns/zarf/src/pkg/utils"
	"github.com/defenseunicorns/zarf/src/pkg/utils/helpers"
	"github.com/defenseunicorns/zarf/src/pkg/variables"
)

const (
	depMarkerOld = "DATA_INJECTON_MARKER"
	depMarkerNew = "DATA_INJECTION_MARKER"
)

var DeprecatedKeys = map[string]string{
	fmt.Sprintf("###ZARF_%s###", depMarkerOld): fmt.Sprintf("###ZARF_%s###", depMarkerNew),
}

// GetTemplates returns the template keys and values to be used for templating.
func GetZarfTemplates(componentName string, state *types.ZarfState) (templateMap map[string]*variables.TextTemplate, err error) {
	templateMap = make(map[string]*variables.TextTemplate)

	if state != nil {
		regInfo := state.RegistryInfo
		gitInfo := state.GitServer

		builtinMap := map[string]string{
			"STORAGE_CLASS": state.StorageClass,

			// Registry info
			"REGISTRY":           regInfo.Address,
			"NODEPORT":           fmt.Sprintf("%d", regInfo.NodePort),
			"REGISTRY_AUTH_PUSH": regInfo.PushPassword,
			"REGISTRY_AUTH_PULL": regInfo.PullPassword,

			// Git server info
			"GIT_PUSH":      gitInfo.PushUsername,
			"GIT_AUTH_PUSH": gitInfo.PushPassword,
			"GIT_PULL":      gitInfo.PullUsername,
			"GIT_AUTH_PULL": gitInfo.PullPassword,
		}

		// Preserve existing misspelling for backwards compatibility
		builtinMap[depMarkerOld] = config.GetDataInjectionMarker()
		builtinMap[depMarkerNew] = config.GetDataInjectionMarker()

		// Don't template component-specific variables for every component
		switch componentName {
		case "zarf-agent":
			agentTLS := state.AgentTLS
			builtinMap["AGENT_CRT"] = base64.StdEncoding.EncodeToString(agentTLS.Cert)
			builtinMap["AGENT_KEY"] = base64.StdEncoding.EncodeToString(agentTLS.Key)
			builtinMap["AGENT_CA"] = base64.StdEncoding.EncodeToString(agentTLS.CA)

		case "zarf-seed-registry", "zarf-registry":
			builtinMap["SEED_REGISTRY"] = fmt.Sprintf("%s:%s", helpers.IPV4Localhost, config.ZarfSeedPort)
			htpasswd, err := generateHtpasswd(&regInfo)
			if err != nil {
				return templateMap, err
			}
			builtinMap["HTPASSWD"] = htpasswd
			builtinMap["REGISTRY_SECRET"] = regInfo.Secret

		case "logging":
			builtinMap["LOGGING_AUTH"] = state.LoggingSecret
		}

		// Iterate over any custom variables and add them to the mappings for templating
		for key, value := range builtinMap {
			// Builtin keys are always uppercase in the format ###ZARF_KEY###
			templateMap[strings.ToUpper(fmt.Sprintf("###ZARF_%s###", key))] = &variables.TextTemplate{
				Value: value,
			}

			if key == "LOGGING_AUTH" || key == "REGISTRY_SECRET" || key == "HTPASSWD" ||
				key == "AGENT_CA" || key == "AGENT_KEY" || key == "AGENT_CRT" || key == "GIT_AUTH_PULL" ||
				key == "GIT_AUTH_PUSH" || key == "REGISTRY_AUTH_PULL" || key == "REGISTRY_AUTH_PUSH" {
				// Sanitize any builtin templates that are sensitive
				templateMap[strings.ToUpper(fmt.Sprintf("###ZARF_%s###", key))].Sensitive = true
			}
		}
	}

	debugPrintTemplateMap(templateMap)

	return templateMap, nil
}

// generateHtpasswd returns an htpasswd string for the current state's RegistryInfo.
func generateHtpasswd(regInfo *types.RegistryInfo) (string, error) {
	// Only calculate this for internal registries to allow longer external passwords
	if regInfo.InternalRegistry {
		pushUser, err := utils.GetHtpasswdString(regInfo.PushUsername, regInfo.PushPassword)
		if err != nil {
			return "", fmt.Errorf("error generating htpasswd string: %w", err)
		}

		pullUser, err := utils.GetHtpasswdString(regInfo.PullUsername, regInfo.PullPassword)
		if err != nil {
			return "", fmt.Errorf("error generating htpasswd string: %w", err)
		}

		return fmt.Sprintf("%s\\n%s", pushUser, pullUser), nil
	}

	return "", nil
}

func debugPrintTemplateMap(templateMap map[string]*variables.TextTemplate) {
	debugText := "templateMap = { "

	for key, template := range templateMap {
		if template.Sensitive {
			debugText += fmt.Sprintf("\"%s\": \"**sanitized**\", ", key)
		} else {
			debugText += fmt.Sprintf("\"%s\": \"%s\", ", key, template.Value)
		}
	}

	debugText += " }"

	message.Debug(debugText)
}

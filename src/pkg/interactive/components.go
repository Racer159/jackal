// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package interactive contains functions for interacting with the user via STDIN.
package interactive

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/pterm/pterm"
	"github.com/racer159/jackal/src/pkg/message"
	"github.com/racer159/jackal/src/pkg/utils"
	"github.com/racer159/jackal/src/types"
)

// SelectOptionalComponent prompts to confirm optional components
func SelectOptionalComponent(component types.JackalComponent) (confirm bool, err error) {
	message.HorizontalRule()

	displayComponent := component
	displayComponent.Description = ""
	utils.ColorPrintYAML(displayComponent, nil, false)
	if component.Description != "" {
		message.Question(component.Description)
	}

	prompt := &survey.Confirm{
		Message: fmt.Sprintf("Deploy the %s component?", component.Name),
		Default: component.Default,
	}

	return confirm, survey.AskOne(prompt, &confirm)
}

// SelectChoiceGroup prompts to select component groups
func SelectChoiceGroup(componentGroup []types.JackalComponent) (types.JackalComponent, error) {
	message.HorizontalRule()

	var chosen int
	var options []string

	for _, component := range componentGroup {
		text := fmt.Sprintf("Name: %s\n  Description: %s\n", component.Name, component.Description)
		options = append(options, text)
	}

	prompt := &survey.Select{
		Message: "Select a component to deploy:",
		Options: options,
	}

	pterm.Println()

	return componentGroup[chosen], survey.AskOne(prompt, &chosen)
}

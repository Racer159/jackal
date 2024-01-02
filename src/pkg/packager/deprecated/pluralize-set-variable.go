// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package deprecated handles package deprecations and migrations
package deprecated

import (
	"fmt"

	"github.com/defenseunicorns/zarf/src/types"
)

type migrateSetVariableToSetVariables struct{}

func (m migrateSetVariableToSetVariables) name() string {
	return PluralizeSetVariable
}

// If the component has already been migrated, clear the deprecated setVariable.
func (m migrateSetVariableToSetVariables) clear(mc types.ZarfComponent) types.ZarfComponent {
	clear := func(actions []types.ZarfComponentAction) []types.ZarfComponentAction {
		for i := range actions {
			actions[i].DeprecatedSetVariable = ""
		}

		return actions
	}

	// Clear OnCreate SetVariables
	mc.Actions.OnCreate.After = clear(mc.Actions.OnCreate.After)
	mc.Actions.OnCreate.Before = clear(mc.Actions.OnCreate.Before)
	mc.Actions.OnCreate.OnSuccess = clear(mc.Actions.OnCreate.OnSuccess)
	mc.Actions.OnCreate.OnFailure = clear(mc.Actions.OnCreate.OnFailure)

	// Clear OnDeploy SetVariables
	mc.Actions.OnDeploy.After = clear(mc.Actions.OnDeploy.After)
	mc.Actions.OnDeploy.Before = clear(mc.Actions.OnDeploy.Before)
	mc.Actions.OnDeploy.OnSuccess = clear(mc.Actions.OnDeploy.OnSuccess)
	mc.Actions.OnDeploy.OnFailure = clear(mc.Actions.OnDeploy.OnFailure)

	// Clear OnRemove SetVariables
	mc.Actions.OnRemove.After = clear(mc.Actions.OnRemove.After)
	mc.Actions.OnRemove.Before = clear(mc.Actions.OnRemove.Before)
	mc.Actions.OnRemove.OnSuccess = clear(mc.Actions.OnRemove.OnSuccess)
	mc.Actions.OnRemove.OnFailure = clear(mc.Actions.OnRemove.OnFailure)

	return mc
}

func (m migrateSetVariableToSetVariables) run(c types.ZarfComponent) (types.ZarfComponent, string) {
	hasSetVariable := false

	migrate := func(actions []types.ZarfComponentAction) []types.ZarfComponentAction {
		for i := range actions {
			if actions[i].DeprecatedSetVariable != "" && len(actions[i].SetVariables) < 1 {
				hasSetVariable = true
				actions[i].SetVariables = []types.ZarfComponentActionSetVariable{
					{
						Name:      actions[i].DeprecatedSetVariable,
						Sensitive: false,
					},
				}
			}
		}

		return actions
	}

	// Migrate OnCreate SetVariables
	c.Actions.OnCreate.After = migrate(c.Actions.OnCreate.After)
	c.Actions.OnCreate.Before = migrate(c.Actions.OnCreate.Before)
	c.Actions.OnCreate.OnSuccess = migrate(c.Actions.OnCreate.OnSuccess)
	c.Actions.OnCreate.OnFailure = migrate(c.Actions.OnCreate.OnFailure)

	// Migrate OnDeploy SetVariables
	c.Actions.OnDeploy.After = migrate(c.Actions.OnDeploy.After)
	c.Actions.OnDeploy.Before = migrate(c.Actions.OnDeploy.Before)
	c.Actions.OnDeploy.OnSuccess = migrate(c.Actions.OnDeploy.OnSuccess)
	c.Actions.OnDeploy.OnFailure = migrate(c.Actions.OnDeploy.OnFailure)

	// Migrate OnRemove SetVariables
	c.Actions.OnRemove.After = migrate(c.Actions.OnRemove.After)
	c.Actions.OnRemove.Before = migrate(c.Actions.OnRemove.Before)
	c.Actions.OnRemove.OnSuccess = migrate(c.Actions.OnRemove.OnSuccess)
	c.Actions.OnRemove.OnFailure = migrate(c.Actions.OnRemove.OnFailure)

	// Leave deprecated setVariable in place, but warn users
	if hasSetVariable {
		return c, fmt.Sprintf("Component '%s' is using setVariable in actions which will be removed in Zarf v1.0.0. Please migrate to the list form of setVariables.", c.Name)
	}

	return c, ""
}

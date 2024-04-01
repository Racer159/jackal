// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package cluster contains Jackal-specific cluster management functions.
package cluster

import (
	"testing"

	"github.com/racer159/jackal/src/types"
	"github.com/stretchr/testify/require"
)

// TestPackageSecretNeedsWait verifies that Jackal waits for webhooks to complete correctly.
func TestPackageSecretNeedsWait(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name            string
		deployedPackage *types.DeployedPackage
		component       types.JackalComponent
		skipWebhooks    bool
		needsWait       bool
		waitSeconds     int
		hookName        string
	}

	var (
		componentName = "test-component"
		packageName   = "test-package"
		webhookName   = "test-webhook"
	)

	testCases := []testCase{
		{
			name:      "NoWebhooks",
			component: types.JackalComponent{Name: componentName},
			deployedPackage: &types.DeployedPackage{
				Name:              packageName,
				ComponentWebhooks: map[string]map[string]types.Webhook{},
			},
			needsWait:   false,
			waitSeconds: 0,
			hookName:    "",
		},
		{
			name:      "WebhookRunning",
			component: types.JackalComponent{Name: componentName},
			deployedPackage: &types.DeployedPackage{
				Name: packageName,
				ComponentWebhooks: map[string]map[string]types.Webhook{
					componentName: {
						webhookName: types.Webhook{
							Status:              types.WebhookStatusRunning,
							WaitDurationSeconds: 10,
						},
					},
				},
			},
			needsWait:   true,
			waitSeconds: 10,
			hookName:    webhookName,
		},
		// Ensure we only wait on running webhooks for the provided component
		{
			name:      "WebhookRunningOnDifferentComponent",
			component: types.JackalComponent{Name: componentName},
			deployedPackage: &types.DeployedPackage{
				Name: packageName,
				ComponentWebhooks: map[string]map[string]types.Webhook{
					"different-component": {
						webhookName: types.Webhook{
							Status:              types.WebhookStatusRunning,
							WaitDurationSeconds: 10,
						},
					},
				},
			},
			needsWait:   false,
			waitSeconds: 0,
			hookName:    "",
		},
		{
			name:      "WebhookSucceeded",
			component: types.JackalComponent{Name: componentName},
			deployedPackage: &types.DeployedPackage{
				Name: packageName,
				ComponentWebhooks: map[string]map[string]types.Webhook{
					componentName: {
						webhookName: types.Webhook{
							Status: types.WebhookStatusSucceeded,
						},
					},
				},
			},
			needsWait:   false,
			waitSeconds: 0,
			hookName:    "",
		},
		{
			name:      "WebhookFailed",
			component: types.JackalComponent{Name: componentName},
			deployedPackage: &types.DeployedPackage{
				Name: packageName,
				ComponentWebhooks: map[string]map[string]types.Webhook{
					componentName: {
						webhookName: types.Webhook{
							Status: types.WebhookStatusFailed,
						},
					},
				},
			},
			needsWait:   false,
			waitSeconds: 0,
			hookName:    "",
		},
		{
			name:      "WebhookRemoving",
			component: types.JackalComponent{Name: componentName},
			deployedPackage: &types.DeployedPackage{
				Name: packageName,
				ComponentWebhooks: map[string]map[string]types.Webhook{
					componentName: {
						webhookName: types.Webhook{
							Status: types.WebhookStatusRemoving,
						},
					},
				},
			},
			needsWait:   false,
			waitSeconds: 0,
			hookName:    "",
		},
		{
			name:      "SkipWaitForYOLO",
			component: types.JackalComponent{Name: componentName},
			deployedPackage: &types.DeployedPackage{
				Name: packageName,
				Data: types.JackalPackage{
					Metadata: types.JackalMetadata{
						YOLO: true,
					},
				},
				ComponentWebhooks: map[string]map[string]types.Webhook{
					componentName: {
						webhookName: types.Webhook{
							Status:              types.WebhookStatusRunning,
							WaitDurationSeconds: 10,
						},
					},
				},
			},
			needsWait:   false,
			waitSeconds: 0,
			hookName:    "",
		},
		{
			name:         "SkipWebhooksFlagUsed",
			component:    types.JackalComponent{Name: componentName},
			skipWebhooks: true,
			deployedPackage: &types.DeployedPackage{
				Name: packageName,
				ComponentWebhooks: map[string]map[string]types.Webhook{
					componentName: {
						webhookName: types.Webhook{
							Status:              types.WebhookStatusRunning,
							WaitDurationSeconds: 10,
						},
					},
				},
			},
			needsWait:   false,
			waitSeconds: 0,
			hookName:    "",
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			c := &Cluster{}

			needsWait, waitSeconds, hookName := c.PackageSecretNeedsWait(testCase.deployedPackage, testCase.component, testCase.skipWebhooks)

			require.Equal(t, testCase.needsWait, needsWait)
			require.Equal(t, testCase.waitSeconds, waitSeconds)
			require.Equal(t, testCase.hookName, hookName)
		})
	}
}

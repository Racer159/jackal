// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package helm contains operations for working with helm charts.
package helm

import (
	"regexp"

	"github.com/defenseunicorns/jackal/src/pkg/cluster"
	"github.com/defenseunicorns/jackal/src/pkg/message"
	"helm.sh/helm/v3/pkg/action"
)

// Destroy removes JackalInitPackage charts from the cluster and optionally all Jackal-installed charts.
func Destroy(purgeAllJackalInstallations bool) {
	spinner := message.NewProgressSpinner("Removing Jackal-installed charts")
	defer spinner.Stop()

	h := Helm{}

	// Initially load the actionConfig without a namespace
	err := h.createActionConfig("", spinner)
	if err != nil {
		// Don't fatal since this is a removal action
		spinner.Errorf(err, "Unable to initialize the K8s client")
		return
	}

	// Match a name that begins with "jackal-"
	// Explanation: https://regex101.com/r/3yzKZy/1
	jackalPrefix := regexp.MustCompile(`(?m)^jackal-`)

	// Get a list of all releases in all namespaces
	list := action.NewList(h.actionConfig)
	list.All = true
	list.AllNamespaces = true
	// Uninstall in reverse order
	list.ByDate = true
	list.SortReverse = true
	releases, err := list.Run()
	if err != nil {
		// Don't fatal since this is a removal action
		spinner.Errorf(err, "Unable to get the list of installed charts")
	}

	// Iterate over all releases
	for _, release := range releases {
		if !purgeAllJackalInstallations && release.Namespace != cluster.JackalNamespaceName {
			// Don't process releases outside the jackal namespace unless purge all is true
			continue
		}
		// Filter on jackal releases
		if jackalPrefix.MatchString(release.Name) {
			spinner.Updatef("Uninstalling helm chart %s/%s", release.Namespace, release.Name)
			if err = h.RemoveChart(release.Namespace, release.Name, spinner); err != nil {
				// Don't fatal since this is a removal action
				spinner.Errorf(err, "Unable to uninstall the chart")
			}
		}
	}

	spinner.Success()
}

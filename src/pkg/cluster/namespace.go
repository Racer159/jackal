// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package cluster contains Jackal-specific cluster management functions.
package cluster

import (
	"context"

	"github.com/Racer159/jackal/src/pkg/message"
)

// DeleteJackalNamespace deletes the Jackal namespace from the connected cluster.
func (c *Cluster) DeleteJackalNamespace() {
	spinner := message.NewProgressSpinner("Deleting the jackal namespace from this cluster")
	defer spinner.Stop()

	c.DeleteNamespace(context.TODO(), JackalNamespaceName)
}

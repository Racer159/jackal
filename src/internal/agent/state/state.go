// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package state provides helpers for interacting with the Jackal agent state.
package state

import (
	"encoding/json"
	"os"

	"github.com/racer159/jackal/src/types"
)

const jackalStatePath = "/etc/jackal-state/state"

// GetJackalStateFromAgentPod reads the state json file that was mounted into the agent pods.
func GetJackalStateFromAgentPod() (state *types.JackalState, err error) {
	// Read the state file
	stateFile, err := os.ReadFile(jackalStatePath)
	if err != nil {
		return nil, err
	}

	// Unmarshal the json file into a Go struct
	return state, json.Unmarshal(stateFile, &state)
}

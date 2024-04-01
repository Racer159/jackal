// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package message provides a rich set of functions for displaying messages to the user.
package message

import (
	"fmt"

	"github.com/defenseunicorns/jackal/src/types"
)

// PrintConnectStringTable prints a table of connect strings.
func PrintConnectStringTable(connectStrings types.ConnectStrings) {
	Debugf("message.PrintConnectStringTable(%#v)", connectStrings)

	if len(connectStrings) > 0 {
		connectData := [][]string{}
		// Loop over each connectStrings and convert to a string matrix
		for name, connect := range connectStrings {
			name = fmt.Sprintf("jackal connect %s", name)
			connectData = append(connectData, []string{name, connect.Description})
		}

		// Create the table output with the data
		header := []string{"Connect Command", "Description"}
		Table(header, connectData)
	}
}

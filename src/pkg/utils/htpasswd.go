// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package utils provides generic utility functions.
package utils

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// GetHtpasswdString converts a username and password to a properly formatted and hashed format for `htpasswd`.
func GetHtpasswdString(username string, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s", username, hash), nil
}

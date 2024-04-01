// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package common handles command configuration across all commands
package common

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/racer159/jackal/src/config/lang"
	"github.com/racer159/jackal/src/pkg/message"
)

// SuppressGlobalInterrupt suppresses the global error on an interrupt
var SuppressGlobalInterrupt = false

// SetBaseDirectory sets the base directory. This is a directory with a jackal.yaml.
func SetBaseDirectory(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return "."
}

// ExitOnInterrupt catches an interrupt and exits with fatal error
func ExitOnInterrupt() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if !SuppressGlobalInterrupt {
			message.Fatal(lang.ErrInterrupt, lang.ErrInterrupt.Error())
		}
	}()
}

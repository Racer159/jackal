// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Jackal Authors

// Package message provides a rich set of functions for displaying messages to the user.
package message

import (
	"context"
	"log/slog"
)

// JackalHandler is a simple handler that implements the slog.Handler interface
type JackalHandler struct{}

// Enabled is always set to true as jackal logging functions are already aware of if they are allowed to be called
func (z JackalHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

// WithAttrs is not suppported
func (z JackalHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return z
}

// WithGroup is not supported
func (z JackalHandler) WithGroup(_ string) slog.Handler {
	return z
}

// Handle prints the respective logging function in jackal
// This function ignores any key pairs passed through the record
func (z JackalHandler) Handle(_ context.Context, record slog.Record) error {
	level := record.Level
	message := record.Message

	switch level {
	case slog.LevelDebug:
		Debug(message)
	case slog.LevelInfo:
		Info(message)
	case slog.LevelWarn:
		Warn(message)
	case slog.LevelError:
		Warn(message)
	}
	return nil
}

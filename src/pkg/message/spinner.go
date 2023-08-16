// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package message provides a rich set of functions for displaying messages to the user.
package message

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/pterm/pterm"
)

var activeSpinner *Spinner

var sequence = []string{`  ⠋ `, `  ⠙ `, `  ⠹ `, `  ⠸ `, `  ⠼ `, `  ⠴ `, `  ⠦ `, `  ⠧ `, `  ⠇ `, `  ⠏ `}

// Spinner is a wrapper around pterm.SpinnerPrinter.
type Spinner struct {
	spinner        *pterm.SpinnerPrinter
	startText      string
	termWidth      int
	preserveWrites bool
}

// NewProgressSpinner creates a new progress spinner.
func NewProgressSpinner(format string, a ...any) *Spinner {
	if activeSpinner != nil {
		debugPrinter(2, "Active spinner already exists")
		return activeSpinner
	}

	var spinner *pterm.SpinnerPrinter
	text := pterm.Sprintf(format, a...)
	if NoProgress {
		Info(text)
	} else {
		spinner, _ = pterm.DefaultSpinner.
			WithRemoveWhenDone(false).
			// Src: https://github.com/gernest/wow/blob/master/spin/spinners.go#L335
			WithSequence(sequence...).
			Start(text)
	}

	activeSpinner = &Spinner{
		spinner:   spinner,
		startText: text,
		termWidth: pterm.GetTerminalWidth(),
	}

	return activeSpinner
}

// EnablePreserveWrites enables preserving writes to the terminal.
func (p *Spinner) EnablePreserveWrites() {
	p.preserveWrites = true
}

// DisablePreserveWrites disables preserving writes to the terminal.
func (p *Spinner) DisablePreserveWrites() {
	p.preserveWrites = false
}

// Write the given text to the spinner.
func (p *Spinner) Write(raw []byte) (int, error) {
	size := len(raw)
	if NoProgress {
		if p.preserveWrites {
			pterm.Printfln("     %s", string(raw))
		}

		return size, nil
	}

	// Split the text into lines and update the spinner for each line.
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		// Only be fancy if preserve writes is enabled.
		if p.preserveWrites {
			text := pterm.Sprintf("     %s", scanner.Text())
			pterm.Fprinto(p.spinner.Writer, strings.Repeat(" ", p.termWidth))
			pterm.Fprintln(p.spinner.Writer, text)
		} else {
			// Otherwise just update the spinner text.
			p.spinner.UpdateText(scanner.Text())
		}
	}

	return size, nil
}

// Updatef updates the spinner text.
func (p *Spinner) Updatef(format string, a ...any) {
	if NoProgress {
		debugPrinter(2, fmt.Sprintf(format, a...))
		return
	}

	text := pterm.Sprintf(format, a...)
	p.spinner.UpdateText(text)
}

// Stop the spinner.
func (p *Spinner) Stop() {
	if p.spinner != nil && p.spinner.IsActive {
		_ = p.spinner.Stop()
	}
	activeSpinner = nil
}

// Success prints a success message and stops the spinner.
func (p *Spinner) Success() {
	p.Successf(p.startText)
}

// Successf prints a success message with the spinner and stops it.
func (p *Spinner) Successf(format string, a ...any) {
	text := pterm.Sprintf(format, a...)
	if p.spinner != nil {
		p.spinner.Success(text)
	} else {
		Info(text)
	}
	p.Stop()
}

// Warnf prints a warning message with the spinner.
func (p *Spinner) Warnf(format string, a ...any) {
	text := pterm.Sprintf(format, a...)
	if p.spinner != nil {
		p.spinner.Warning(text)
	} else {
		Warn(text)
	}
}

// Errorf prints an error message with the spinner.
func (p *Spinner) Errorf(err error, format string, a ...any) {
	p.Warnf(format, a...)
	debugPrinter(2, err)
}

// Fatal calls message.Fatalf with the given error.
func (p *Spinner) Fatal(err error) {
	p.Fatalf(err, p.startText)
}

// Fatalf calls message.Fatalf with the given error and format.
func (p *Spinner) Fatalf(err error, format string, a ...any) {
	if p.spinner != nil {
		p.spinner.RemoveWhenDone = true
		_ = p.spinner.Stop()
		activeSpinner = nil
	}
	Fatalf(err, format, a...)
}

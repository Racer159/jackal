// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2021-Present The Zarf Authors

// Package variables contains functions for interacting with variables
package variables

// VariableType represents a type of a Zarf package variable
type VariableType string

const (
	// RawVariableType is the default type for a Zarf package variable
	RawVariableType VariableType = "raw"
	// FileVariableType is a type for a Zarf package variable that loads its contents from a file
	FileVariableType VariableType = "file"
)

// Variable is a variable that can be used to dynamically template K8s resources or run in actions.
type Variable struct {
	Name        string       `json:"name" jsonschema:"description=The name to be used for the variable,pattern=^[A-Z0-9_]+$"`
	Description string       `json:"description,omitempty" jsonschema:"description=A description of the variable to be used when prompting the user a value"`
	Default     string       `json:"default,omitempty" jsonschema:"description=The default value to use for the variable"`
	Prompt      bool         `json:"prompt,omitempty" jsonschema:"description=Whether to prompt the user for input for this variable"`
	Sensitive   bool         `json:"sensitive,omitempty" jsonschema:"description=Whether to mark this variable as sensitive to not print it in the Zarf log"`
	AutoIndent  bool         `json:"autoIndent,omitempty" jsonschema:"description=Whether to automatically indent the variable's value (if multiline) when templating. Based on the number of chars before the start of ###ZARF_VAR_."`
	Pattern     string       `json:"pattern,omitempty" jsonschema:"description=An optional regex pattern that a variable value must match before a package can be deployed."`
	Type        VariableType `json:"type,omitempty" jsonschema:"description=Changes the handling of a variable to load contents differently (i.e. from a file rather than as a raw variable - templated files should be kept below 1 MiB),enum=raw,enum=file"`
}

// Constant are constants that can be used to dynamically template K8s resources or run in actions.
type Constant struct {
	Name  string `json:"name" jsonschema:"description=The name to be used for the constant,pattern=^[A-Z0-9_]+$"`
	Value string `json:"value" jsonschema:"description=The value to set for the constant during deploy"`
	// Include a description that will only be displayed during package create/deploy confirm prompts
	Description string `json:"description,omitempty" jsonschema:"description=A description of the constant to explain its purpose on package create or deploy confirmation prompts"`
	AutoIndent  bool   `json:"autoIndent,omitempty" jsonschema:"description=Whether to automatically indent the variable's value (if multiline) when templating. Based on the number of chars before the start of ###ZARF_CONST_."`
	Pattern     string `json:"pattern,omitempty" jsonschema:"description=An optional regex pattern that a constant value must match before a package can be created."`
}

// SetVariable tracks internal variables that have been set during this run of Zarf
type SetVariable struct {
	Name       string       `json:"name" jsonschema:"description=The name to be used for the variable,pattern=^[A-Z0-9_]+$"`
	Sensitive  bool         `json:"sensitive,omitempty" jsonschema:"description=Whether to mark this variable as sensitive to not print it in the Zarf log"`
	AutoIndent bool         `json:"autoIndent,omitempty" jsonschema:"description=Whether to automatically indent the variable's value (if multiline) when templating. Based on the number of chars before the start of ###ZARF_VAR_."`
	Value      string       `json:"value" jsonschema:"description=The value the variable is currently set with"`
	Type       VariableType `json:"type,omitempty" jsonschema:"description=Changes the handling of a variable to load contents differently (i.e. from a file rather than as a raw variable - templated files should be kept below 1 MiB),enum=raw,enum=file"`
}

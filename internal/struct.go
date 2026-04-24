// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: MIT

package internal

// CustomDeleteListJSON is the JSON-based lambda deletion input schema.
type CustomDeleteListJSON struct {
	Lambdas []string `json:"lambdas"`
}

// CustomDeleteListYAML is the YAML-based lambda deletion input schema.
type CustomDeleteListYAML struct {
	Lambdas []string `yaml:"lambdas"`
}

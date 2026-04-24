// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: MIT

// Package internal contains input file parsing helpers.
package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

var (
	errReadInputFile     = errors.New("unable to read the input file")
	errDecodeYAMLFile    = errors.New("unable to decode the YAML file")
	errUnmarshalJSONFile = errors.New("unable to unmarshall the json file")
	errInvalidFileType   = errors.New("invalid file type provided. Must be of type json, yaml or yml")
)

// GenerateLambdaDeleteList is a function that takes a file path as input and returns a list of Lambdas to be deleted.
func GenerateLambdaDeleteList(filePath string) ([]string, error) {
	var (
		deleteListYAML CustomDeleteListYAML
		deleteListJSON CustomDeleteListJSON
		output         []string
	)

	fileType, err := determineFileType(filePath)
	if err != nil {
		return []string{}, err
	}

	if fileType == "json" {
		deleteListJSON, err = readConfigFileJSON(filePath)
		if err != nil {
			return deleteListJSON.Lambdas, err
		}

		output = deleteListJSON.Lambdas
	}

	if fileType == "yaml" {
		deleteListYAML, err = readConfigFileYAML(filePath)
		if err != nil {
			return deleteListYAML.Lambdas, err
		}

		output = deleteListYAML.Lambdas
	}

	return output, err
}

// readConfigFileYaml is a function that takes a file path as input and returns a list of Lambdas to be deleted. A YAML file is expected.
func readConfigFileYAML(file string) (CustomDeleteListYAML, error) {
	var (
		list CustomDeleteListYAML
	)

	// #nosec G304 -- this CLI intentionally accepts user-provided file paths.
	fileContent, err := os.ReadFile(file)
	if err != nil {
		return list, fmt.Errorf("%w: %w", errReadInputFile, err)
	}

	dc := yaml.NewDecoder(strings.NewReader(string(fileContent)))
	dc.KnownFields(true)

	err = dc.Decode(&list)
	if err != nil {
		return list, fmt.Errorf("%w. Ensure the file is in the correct format and that all fields are correct: %w", errDecodeYAMLFile, err)
	}

	return list, nil
}

// readConfigFileJson is a function that takes a file path as input and returns a list of Lambdas to be deleted. A JSON file is expected.
func readConfigFileJSON(file string) (CustomDeleteListJSON, error) {
	var (
		list CustomDeleteListJSON
	)

	// #nosec G304 -- this CLI intentionally accepts user-provided file paths.
	fileContent, err := os.ReadFile(file)
	if err != nil {
		return list, fmt.Errorf("%w: %w", errReadInputFile, err)
	}

	err = json.Unmarshal(fileContent, &list)
	if err != nil {
		return list, fmt.Errorf("%w: %w", errUnmarshalJSONFile, err)
	}

	return list, nil
}

// determineFileType validates the existence of an input file and ensures its prefix is json | yaml | yml.
func determineFileType(file string) (string, error) {
	f, err := os.Stat(file)
	if err != nil {
		return "none", fmt.Errorf("%w: %w", errReadInputFile, err)
	}

	var fileType string

	switch {
	case strings.HasSuffix(f.Name(), "yaml"):
		fileType = "yaml"

	case strings.HasSuffix(f.Name(), "json"):
		fileType = "json"

	case strings.HasSuffix(f.Name(), "yml"):
		fileType = "yaml"

	default:
		fileType = "none"
		err = errInvalidFileType
	}

	return fileType, err
}

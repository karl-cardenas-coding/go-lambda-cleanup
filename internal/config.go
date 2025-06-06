// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: MIT

package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// GenerateLambdaDeleteList is a function that takes a file path as input and returns a list of Lambdas to be deleted.
func GenerateLambdaDeleteList(filePath string) ([]string, error) {
	var (
		deleteListYaml CustomDeleteListYaml
		deleteListJson CustomDeleteListJson
		output         []string
	)

	fileType, err := determineFileType(filePath)
	if err != nil {
		return []string{}, err
	}

	if fileType == "json" {
		deleteListJson, err = readConfigFileJson(filePath)
		if err != nil {
			return deleteListJson.Lambdas, err
		}

		output = deleteListJson.Lambdas
	}

	if fileType == "yaml" {
		deleteListYaml, err = readConfigFileYaml(filePath)
		if err != nil {
			return deleteListYaml.Lambdas, err
		}

		output = deleteListYaml.Lambdas
	}

	return output, err
}

// readConfigFileYaml is a function that takes a file path as input and returns a list of Lambdas to be deleted. A YAML file is expected.
func readConfigFileYaml(file string) (CustomDeleteListYaml, error) {
	var (
		list CustomDeleteListYaml
	)

	fileContent, err := os.ReadFile(file)
	if err != nil {
		return list, errors.New("unable to read the input file")
	}

	dc := yaml.NewDecoder(strings.NewReader(string(fileContent)))
	dc.KnownFields(true)

	if err := dc.Decode(&list); err != nil {
		return list, fmt.Errorf("unable to decode the YAML file. Ensure the file is in the correct format and that all fields are correct. %s", err.Error())
	}

	return list, err
}

// readConfigFileJson is a function that takes a file path as input and returns a list of Lambdas to be deleted. A JSON file is expected.
func readConfigFileJson(file string) (CustomDeleteListJson, error) {
	var (
		list CustomDeleteListJson
	)

	fileContent, err := os.ReadFile(file)
	if err != nil {
		return list, errors.New("unable to read the input file")
	}

	err = json.Unmarshal(fileContent, &list)
	if err != nil {
		return list, errors.New("unable to unmarshall the json file")
	}

	return list, err
}

// determineFileType validates the existence of an input file and ensures its prefix is json | yaml | yml.
func determineFileType(file string) (string, error) {
	f, err := os.Stat(file)
	if err != nil {
		return "none", errors.New("unable to read the input file")
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
		err = errors.New("invalid file type provided. Must be of type json, yaml or yml")
	}

	return fileType, err
}

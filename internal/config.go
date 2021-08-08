package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// This returns a list of Lambda names that are read from an input file.
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

// This function reads a yaml input file and returns a list of Lambdas
func readConfigFileYaml(file string) (CustomDeleteListYaml, error) {
	var (
		list CustomDeleteListYaml
	)
	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	yaml.Unmarshal(fileContent, &list)
	if err != nil {
		err = errors.New("unable to unmarshall the YAML file")
	}
	return list, err
}

// This function reads a json input file and returns a list of Lambdas
func readConfigFileJson(file string) (CustomDeleteListJson, error) {
	var (
		list CustomDeleteListJson
	)
	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(fileContent, &list)
	if err != nil {
		fmt.Println("error:", err)
	}

	return list, nil
}

// This function validates the existence of an input file and ensures its prefix is json | yaml | yml
func determineFileType(file string) (string, error) {
	f, err := os.Stat(file)
	var fileType string
	if err == nil {
		switch {
		case strings.HasSuffix(f.Name(), "yaml"):
			fileType = "yaml"

		case strings.HasSuffix(f.Name(), "json"):
			fileType = "json"

		case strings.HasSuffix(f.Name(), "yml"):
			fileType = "yaml"

		default:
			fileType = "none"
			err = errors.New("invalid file type provided. Must be of type json, ymal or yml")
		}
	}
	return fileType, err
}

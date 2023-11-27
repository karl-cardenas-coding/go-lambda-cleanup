// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: MIT

package internal

import (
	"testing"
)

func TestReadConfigFileYaml(t *testing.T) {

	want := 2
	got, err := readConfigFileYaml("../tests/test.yaml")
	if len(got.Lambdas) != want || err != nil {
		t.Fatalf("Failed to read the Yaml file. Expected %d but received %d", want, len(got.Lambdas))
	}

	_, err = readConfigFileYaml("test.json")
	if err == nil {
		t.Fatalf("Failed to read the Yaml file. Expected error but received %d", err)
	}
}

func TestReadConfigFileYml(t *testing.T) {

	want := 2
	got, err := readConfigFileYaml("../tests/test.yml")
	if len(got.Lambdas) != want || err != nil {
		t.Fatalf("Failed to read the Yaml file. Expected %d but received %d", want, len(got.Lambdas))
	}
}

func TestReadConfigFileJson(t *testing.T) {

	want := 2
	got, err := readConfigFileJson("../tests/test.json")
	if len(got.Lambdas) != want || err != nil {
		t.Fatalf("Failed to read the json file. Expected %d but received %d", want, len(got.Lambdas))
	}

	_, err = readConfigFileJson("test.yaml")
	if err == nil {
		t.Fatalf("Failed to read the json file. Expected error but received %d", err)
	}
}

func TestDetermineFileTypeYaml(t *testing.T) {

	want := "yaml"
	got, err := determineFileType("../tests/test.yaml")
	if got != want || err != nil {
		t.Fatalf("Failed to read the json file. Expected %s but received %s", want, got)
	}
}

func TestDetermineFileTypeInvalid(t *testing.T) {

	want := "none"
	got, err := determineFileType("../tests/handler.zip")
	if got != want || err == nil {
		t.Fatalf("Failed to read the json file. Expected %s but received %s", want, got)
	}

}

func TestDetermineFileTypeYml(t *testing.T) {

	want := "yaml"
	got, err := determineFileType("../tests/test.yml")
	if got != want || err != nil {
		t.Fatalf("Failed to read the json file. Expected %s but received %s", want, got)
	}
}

func TestDetermineFileTypeJson(t *testing.T) {

	want := "json"
	got, err := determineFileType("../tests/test.json")
	if got != want || err != nil {
		t.Fatalf("Failed to read the json file. Expected %s but received %s", want, got)
	}
}

func TestInvalidJSON(t *testing.T) {

	_, err := readConfigFileJson("../tests/invalid.json")
	if err == nil {
		t.Fatalf("An error was expected but received %s", err)
	}
}

func TestGenerateLambdaDeleteListJson(t *testing.T) {
	want := []string{
		"stopEC2-instances",
		"putControls",
	}
	got, err := GenerateLambdaDeleteList("../tests/test.json")
	if len(got) != len(want) || err != nil {
		t.Fatalf("Failed to read the json file and expected content. Expected %d but received %d", len(want), len(got))
	}

	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("Failed to read the expected content. Expected %s but received %s", want[index], got[index])
		}
	}
}

func TestInvalidYaml(t *testing.T) {

	_, err := readConfigFileYaml("../tests/invalid.yaml")
	if err == nil {
		t.Fatalf("An error was expected but received %s", err)
	}
}

func TestGenerateLambdaDeleteListYaml(t *testing.T) {
	want := []string{
		"stopEC2-instances",
		"putControls",
	}
	got, err := GenerateLambdaDeleteList("../tests/test.yaml")
	if len(got) != len(want) || err != nil {
		t.Fatalf("Failed to read the json file and expected content. Expected %d but received %d", len(want), len(got))
	}

	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("Failed to read the expected content. Expected %s but received %s", want[index], got[index])
		}
	}
}

func TestGenerateLambdaDeleteListYml(t *testing.T) {
	want := []string{
		"stopEC2-instances",
		"putControls",
	}
	got, err := GenerateLambdaDeleteList("../tests/test.yml")
	if len(got) != len(want) || err != nil {
		t.Fatalf("Failed to read the json file and expected content. Expected %d but received %d", len(want), len(got))
	}

	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("Failed to read the expected content. Expected %s but received %s", want[index], got[index])
		}
	}
}

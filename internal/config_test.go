package internal

import (
	"testing"
)

func TestReadConfigFileYaml(t *testing.T) {

	want := 2
	got, err := readConfigFileYaml("test.yaml")
	if len(got.Lambdas) != want || err != nil {
		t.Fatalf("Failed to read the Yaml file. Expected %d but received %d", want, len(got.Lambdas))
	}
}

func TestReadConfigFileYml(t *testing.T) {

	want := 2
	got, err := readConfigFileYaml("test.yml")
	if len(got.Lambdas) != want || err != nil {
		t.Fatalf("Failed to read the Yaml file. Expected %d but received %d", want, len(got.Lambdas))
	}
}

func TestReadConfigFileJson(t *testing.T) {

	want := 2
	got, err := readConfigFileJson("test.json")
	if len(got.Lambdas) != want || err != nil {
		t.Fatalf("Failed to read the json file. Expected %d but received %d", want, len(got.Lambdas))
	}
}

func TestDetermineFileTypeYaml(t *testing.T) {

	want := "yaml"
	got, err := determineFileType("test.yaml")
	if got != want || err != nil {
		t.Fatalf("Failed to read the json file. Expected %s but received %s", want, got)
	}
}

func TestDetermineFileTypeYml(t *testing.T) {

	want := "yaml"
	got, err := determineFileType("test.yml")
	if got != want || err != nil {
		t.Fatalf("Failed to read the json file. Expected %s but received %s", want, got)
	}
}

func TestDetermineFileTypeJson(t *testing.T) {

	want := "json"
	got, err := determineFileType("test.json")
	if got != want || err != nil {
		t.Fatalf("Failed to read the json file. Expected %s but received %s", want, got)
	}
}

func TestGenerateLambdaDeleteListJson(t *testing.T) {
	want := []string{
		"stopEC2-instances",
		"putControls",
	}
	got, err := GenerateLambdaDeleteList("test.json")
	if len(got) != len(want) || err != nil {
		t.Fatalf("Failed to read the json file and expected content. Expected %d but received %d", len(want), len(got))
	}

	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("Failed to read the expected content. Expected %s but received %s", want[index], got[index])
		}
	}
}

func TestGenerateLambdaDeleteListYaml(t *testing.T) {
	want := []string{
		"stopEC2-instances",
		"putControls",
	}
	got, err := GenerateLambdaDeleteList("test.yaml")
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
	got, err := GenerateLambdaDeleteList("test.yml")
	if len(got) != len(want) || err != nil {
		t.Fatalf("Failed to read the json file and expected content. Expected %d but received %d", len(want), len(got))
	}

	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("Failed to read the expected content. Expected %s but received %s", want[index], got[index])
		}
	}
}

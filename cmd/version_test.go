package cmd

import (
	"fmt"
	"testing"
)

const (
	noVersionAvailable              = "No new version available"
	youAreRunningAPreReleaseVersion = "You are running a pre-release version"
)

func TestCheckForNewRelease(t *testing.T) {

	client := createHTTPClient()
	version := "0.0.0"
	useragent := fmt.Sprintf("go-lambda-cleanup/%s", version)

	want := true
	got, msg, err := checkForNewRelease(client, version, useragent)
	if err != nil {
		t.Fatalf("Error checking for new release: %s", err)
	}
	if got != want && msg == noVersionAvailable || msg == youAreRunningAPreReleaseVersion {
		t.Fatalf("Error checking for new release: %s", err)
	}

	version = "100.0.0"
	want2 := true
	got2, msg2, err := checkForNewRelease(client, version, useragent)
	if err != nil {
		t.Fatalf("Error checking for new release: %s", err)
	}
	if got2 != want2 && msg2 != youAreRunningAPreReleaseVersion {
		t.Fatalf("Error checking for new release: %s", err)
	}
}

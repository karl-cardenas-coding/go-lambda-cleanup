// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
	got, msg, err := checkForNewRelease(client, version, useragent, url)
	if err != nil {
		t.Fatalf("Error checking for new release: %s", err)
	}
	if got != want && msg == noVersionAvailable || msg == youAreRunningAPreReleaseVersion {
		t.Fatalf("Error checking for new release: %s", err)
	}

	version = "100.0.0"
	want2 := true
	got2, msg2, err := checkForNewRelease(client, version, useragent, url)
	if err != nil {
		t.Fatalf("Error checking for new release: %s", err)
	}
	if got2 != want2 && msg2 != youAreRunningAPreReleaseVersion {
		t.Fatalf("Error checking for new release: %s", err)
	}
}

func TestErrorPath(t *testing.T) {

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Mock error response", http.StatusInternalServerError)
	}))
	defer mockServer.Close()

	// Override the base URL to use the mock server URL
	baseURL := mockServer.URL

	client := createHTTPClient()
	version := "------"
	useragent := fmt.Sprintf("go-lambda-cleanup/%s", version)

	_, _, err := checkForNewRelease(client, version, useragent, baseURL)
	if err == nil {
		t.Fatalf("Error Expected: %s", err)
	}
}

func TestErrorPathJSON(t *testing.T) {

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the Content-Type header to indicate JSON response
		w.Header().Set("Content-Type", "application/json")

		// Write an invalid JSON payload to the response writer
		w.Write([]byte("invalid-json-payload"))
	}))
	defer mockServer.Close()

	// Override the base URL to use the mock server URL
	baseURL := mockServer.URL

	client := createHTTPClient()
	version := "------"
	useragent := fmt.Sprintf("go-lambda-cleanup/%s", version)

	_, _, err := checkForNewRelease(client, version, useragent, baseURL)
	if err == nil {
		t.Fatalf("Error Expected: %s", err)
	}
}

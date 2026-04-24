// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(VersionCmd)
}

const (
	url = "https://api.github.com/repos/karl-cardenas-coding/go-lambda-cleanup/releases/latest"
)

var (
	errCreateRequest          = errors.New("failed to create release request")
	errConnectReleaseEndpoint = errors.New("error connecting to release endpoint")
	errDecodeReleasePayload   = errors.New("failed to decode release payload")
	errCreateVersionValue     = errors.New("failed to parse version value")
	errCompareVersions        = errors.New("error comparing versions")
)

// VersionCmd prints the current version and checks for newer releases.
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current version number of go-lambda-cleanup",
	Long:  `Prints the current version number of go-lambda-cleanup`,
	RunE: func(_ *cobra.Command, _ []string) error {
		version := "go-lambda-cleanup " + VersionString
		log.Info(version)

		_, message, err := checkForNewRelease(GlobalHTTPClient, VersionString, UserAgent, url)
		if err != nil {
			log.Error(err)

			return err
		}

		log.Info(message)

		return nil
	},
}

//nolint:cyclop // Network and version-comparison branching is explicit to preserve current behavior.
func checkForNewRelease(client *http.Client, currentVersion, useragent, url string) (bool, string, error) {
	var (
		output  bool
		message string
		release Release
	)

	log.Info("Checking for new releases")

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"package":         "cmd",
			"file":            "version.go",
			"parent_function": "checkForNewRelease",
			"function":        "http.NewRequest",
			"error":           err,
			"data":            nil,
		}).Debug("Error creating the HTTP request", IssueMSG)

		return output, message, fmt.Errorf("%w: %w", errCreateRequest, err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", useragent)

	resp, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"package":         "cmd",
			"file":            "version.go",
			"parent_function": "checkForNewRelease",
			"function":        "client.Do",
			"error":           err,
			"data":            nil,
		}).Debug("Error initaiting connection to, ", url, IssueMSG)

		return output, message, fmt.Errorf("%w: %w", errConnectReleaseEndpoint, err)
	}

	if resp != nil && resp.Body != nil {
		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil {
				log.WithFields(log.Fields{
					"package":         "cmd",
					"file":            "version.go",
					"parent_function": "checkForNewRelease",
					"function":        "resp.Body.Close",
					"error":           closeErr,
					"data":            nil,
				}).Debug("Error closing response body", IssueMSG)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			log.WithFields(log.Fields{
				"package":         "cmd",
				"file":            "version.go",
				"parent_function": "checkForNewRelease",
				"function":        "client.Do",
				"error":           err,
				"data":            nil,
			}).Debug("Error initaiting connection to, ", url, IssueMSG)

			return output, message, fmt.Errorf("%w: %s", errConnectReleaseEndpoint, url)
		}
		// Unmarshal the JSON to the Github Release strcut
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			log.WithFields(log.Fields{
				"package":         "cmd",
				"file":            "version.go",
				"parent_function": "checkForNewRelease",
				"function":        "json.NewDecoder",
				"error":           err,
				"data":            nil,
			}).Debug("Error unmarshalling Github response", IssueMSG)

			return output, message, fmt.Errorf("%w: %w", errDecodeReleasePayload, err)
		}

		cVersion, err := version.NewVersion(currentVersion)
		if err != nil {
			log.WithFields(log.Fields{
				"package":         "cmd",
				"file":            "version.go",
				"parent_function": "checkForNewRelease",
				"function":        "version.NewVersion",
				"error":           err,
				"data":            nil,
			}).Debug("Error creating new version", IssueMSG)

			return output, message, fmt.Errorf("%w: %w", errCreateVersionValue, err)
		}

		latestVersion, err := version.NewVersion(release.TagName[1:])
		if err != nil {
			log.WithFields(log.Fields{
				"package":         "cmd",
				"file":            "version.go",
				"parent_function": "checkForNewRelease",
				"function":        "version.NewVersion",
				"error":           err,
				"data":            nil,
			}).Debug("Error creating new version", IssueMSG)

			return output, message, fmt.Errorf("%w: %w", errCreateVersionValue, err)
		}

		switch cVersion.Compare(latestVersion) {
		case -1:
			message = fmt.Sprintf("There is a new release available: %s \n Download it here - %s", release.TagName, release.HTMLURL)
			output = true
		case 0:
			message = "No new version available"
			output = true
		case 1:
			message = "You are running a pre-release version"
			output = true
		default:
			return output, message, errCompareVersions
		}
	} else {
		return output, message, fmt.Errorf("%w: %s", errConnectReleaseEndpoint, url)
	}

	return output, message, nil
}

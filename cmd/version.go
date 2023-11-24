// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: MIT

package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/go-version"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

const (
	url = "https://api.github.com/repos/karl-cardenas-coding/go-lambda-cleanup/releases/latest"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current version number of go-lambda-cleanup",
	Long:  `Prints the current version number of go-lambda-cleanup`,
	Run: func(cmd *cobra.Command, args []string) {
		version := fmt.Sprintf("go-lambda-cleanup v%s", VersionString)
		log.Info(version)
		_, message, err := checkForNewRelease(GlobalHTTPClient, VersionString, UserAgent, url)
		if err != nil {
			log.Error(err)
			log.Fatal("unable to check for new releases. " + IssueMSG)
		}
		log.Info(message)
	},
}

func checkForNewRelease(client *http.Client, currentVersion, useragent, url string) (bool, string, error) {
	var (
		output  bool
		message string
		release Release
	)

	log.Info("Checking for new releases")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"package":         "cmd",
			"file":            "version.go",
			"parent_function": "checkForNewRelease",
			"function":        "http.NewRequest",
			"error":           err,
			"data":            nil,
		}).Debug("Error creating the HTTP request", IssueMSG)
		return output, message, err
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
		return output, message, err
	}

	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.WithFields(log.Fields{
				"package":         "cmd",
				"file":            "version.go",
				"parent_function": "checkForNewRelease",
				"function":        "client.Do",
				"error":           err,
				"data":            nil,
			}).Debug("Error initaiting connection to, ", url, IssueMSG)
			return output, message, fmt.Errorf("error connecting to %s", url)
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
			return output, message, err
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
			return output, message, err
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
			return output, message, err
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
			return output, message, fmt.Errorf("error comparing versions")
		}
	} else {
		return output, message, fmt.Errorf("error connecting to %s", url)
	}

	return output, message, err
}

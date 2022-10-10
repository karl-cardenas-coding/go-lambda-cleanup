package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current version number of go-lambda-cleanup",
	Long:  `Prints the current version number of go-lambda-cleanup`,
	Run: func(cmd *cobra.Command, args []string) {
		version := fmt.Sprintf("go-lambda-cleanup %s", VersionString)
		log.Info(version)
		_, message, err := checkForNewRelease(GlobalHTTPClient, VersionString, UserAgent)
		if err != nil {
			log.Fatal(err)
		}
		log.Info(message)
	},
}

func checkForNewRelease(client *http.Client, currentVersion, useragent string) (bool, string, error) {
	url := "https://api.github.com/repos/karl-cardenas-coding/go-lambda-cleanup/releases/latest"
	var (
		output  bool
		message string
	)

	log.Info("Checking for new releases")
	req, _ := http.NewRequest("GET", url, nil)
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
		}).Debug("Error initaiting connection to, ", url, "If this error persists, please open up an issue on github")
		return output, message, err
	}
	defer resp.Body.Close()

	var release Release

	// Unmarshal the JSON to the Github Release strcut
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		log.WithFields(log.Fields{
			"package":         "cmd",
			"file":            "version.go",
			"parent_function": "checkForNewRelease",
			"function":        "json.NewDecoder",
			"error":           err,
			"data":            nil,
		}).Debug("Error unmarshalling Github response", "If this error persists, please open up an issue on github")
		return output, message, err
	}
	switch semver.Compare(currentVersion, release.TagName) {
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
		return output, message, errors.New("error comparing versions")
	}

	return output, message, nil
}

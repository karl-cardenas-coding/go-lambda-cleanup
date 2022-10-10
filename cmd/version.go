package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
		checkForNewRelease(GlobalHTTPClient)
	},
}

func checkForNewRelease(client *http.Client) (bool, error) {
	url := "https://api.github.com/repos/karl-cardenas-coding/go-lambda-cleanup/releases/latest"
	// version := VersionString
	version := "v1.0.0"
	var output bool = false

	log.Info("Checking for new releases")
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{
			"package":         "cmd",
			"file":            "version.go",
			"parent_function": "checkForNewRelease",
			"function":        "client.Do",
			"error":           err,
			"data":            nil,
		}).Fatal("Error initaiting connection to, ", url, "If this error persists, please open up an issue on github")
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
		}).Fatal("Error unmarshalling Github response", "If this error persists, please open up an issue on github")
	}
	// Check to see if the current version is equivalent to the latest release
	if version != release.TagName {
		log.Info("New version available - ", release.TagName)
		log.Info("Download it here - ", release.HTMLURL)
		output = true
	}

	if version == release.TagName {
		log.Info("No new version available")
		output = false
	}

	return output, nil
}

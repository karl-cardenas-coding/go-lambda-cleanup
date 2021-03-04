package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	VersionString   string = "No version provided"
	CredentialsFile bool
	RegionFlag      string
	Retain          int32
	Debug           bool
)

const (
	ISSUE_MSG = " Please open up a Github issue to report this error! https://github.com/karl-cardenas-coding/go-clean-lambda"
)

var rootCmd = &cobra.Command{
	Use:   "go-lambda",
	Short: "A CLI tool for cleaning up AWS Lambda Version",
	Long:  `A Golang based CLI tool for determining natural catastrophe near you, or a location specified. Visit https://github.com/karl-cardenas-coding/go-clean-lambda for more information.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			log.WithFields(log.Fields{
				"package":         "cmd",
				"file":            "root.go",
				"parent_function": "generateDocFlag",
				"function":        "cmd.Help",
				"error":           err,
				"data":            nil,
			}).Fatal("Error outputting help!", ISSUE_MSG)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&RegionFlag, "region", "r", "", "Specify the desired AWS region to target.")
	rootCmd.PersistentFlags().BoolVarP(&Debug, "verbose", "v", false, "Set to true to enable debugging")
	rootCmd.PersistentFlags().BoolVarP(&CredentialsFile, "credentialsFile", "d", false, "Leverageds the default ~/.aws/crededentials file")
	cleanCmd.Flags().Int32VarP(&Retain, "count", "c", 1, "The number of versions to retain from $LATEST - n-(x)")

	// Establish logging default
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	log.SetOutput(os.Stdout)

	log.SetLevel(log.WarnLevel)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.WithFields(log.Fields{
			"package":  "cmd",
			"file":     "root.go",
			"function": "Execute",
			"error":    err,
			"data":     nil,
		}).Fatal("Error executing the CLI!", ISSUE_MSG)
		os.Exit(1)
	}
}

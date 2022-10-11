package cmd

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	// VersionString is the version of the CLI
	VersionString string = "No version provided"
	// ProfileFlag is the AWS crendentials profile passed in
	ProfileFlag string
	// CredentialsFile is a boolean for the credentials provider logic
	CredentialsFile bool
	// RegionFlag is the AWS Region to target for the execution
	RegionFlag string
	// Retain is the number of versions to retain excluding $LATEST
	Retain int8
	// Verbose is to enable debug output
	Verbose bool
	// DryRun is to enable a preview of what an actual execution would do
	DryRun bool
	// LambdaListFile points a file that contains a listof Lambdas to delete
	LambdaListFile string
	// MoreLambdaDetails is to show information about the Lambda being worked on
	MoreLambdaDetails bool
	// SizeIEC is used to display the size in IEC units
	SizeIEC bool
	// CliConfig is the struct that holds the CLI configuration
	GlobalCliConfig cliConfig
	// HTTPClient is the HTTP client to use for the AWS API calls
	GlobalHTTPClient *http.Client
	// UserAgent is the value to use for the User-Agent header
	UserAgent string
)

const (
	// IssueMSG is a default message to pass to the user
	IssueMSG = " Please open up a Github issue to report this error! https://github.com/karl-cardenas-coding/go-clean-lambda"
)

var rootCmd = &cobra.Command{
	Use:   "glc",
	Short: "A CLI tool for cleaning up AWS Lambda versions",
	Long:  `A CLI tool for cleaning up AWS Lambda versions`,
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
			}).Fatal("Error outputting help!", IssueMSG)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&RegionFlag, "region", "r", "", "Specify the desired AWS region to target.")
	rootCmd.PersistentFlags().StringVarP(&ProfileFlag, "profile", "p", "", "Specify the AWS profile to leverage for authentication.")
	rootCmd.PersistentFlags().StringVarP(&LambdaListFile, "listFile", "l", "", "Specify a file containing Lambdas to delete.")
	rootCmd.PersistentFlags().BoolVarP(&MoreLambdaDetails, "moreLambdaDetails", "m", false, "Set to true to show Lambda names and count of versions to be removed (bool)")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Set to true to enable debugging (bool)")
	rootCmd.PersistentFlags().BoolVarP(&DryRun, "dryrun", "d", false, "Executes a dry run (bool)")
	rootCmd.PersistentFlags().BoolVarP(&SizeIEC, "size-iec", "i", false, "Displays file sizes in IEC units (bool)")
	cleanCmd.Flags().Int8VarP(&Retain, "count", "c", 1, "The number of versions to retain from $LATEST-(n)")

	GlobalCliConfig.RegionFlag = &RegionFlag
	GlobalCliConfig.ProfileFlag = &ProfileFlag
	GlobalCliConfig.LambdaListFile = &LambdaListFile
	GlobalCliConfig.MoreLambdaDetails = &MoreLambdaDetails
	GlobalCliConfig.Verbose = &Verbose
	GlobalCliConfig.DryRun = &DryRun
	GlobalCliConfig.SizeIEC = &SizeIEC
	GlobalCliConfig.Retain = &Retain
	UserAgent = fmt.Sprintf("go-clean-lambda/%s", VersionString)
	// Establish logging default
	log.SetFormatter(&log.TextFormatter{
		DisableColors:   false,
		TimestampFormat: "01/02/06",
		FullTimestamp:   true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	GlobalHTTPClient = createHTTPClient()

}

// Execute is the main execution function
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		if Verbose {
			log.WithFields(log.Fields{
				"package":  "cmd",
				"file":     "root.go",
				"function": "Execute",
				"error":    err,
				"data":     nil,
			}).Fatal("Error executing the CLI!", IssueMSG)
		} else {
			log.Fatal("Error executing the CLI!", IssueMSG)
		}
	}
}

// createHTTPClient creates an HTTP client with TLS
func createHTTPClient() *http.Client {

	// Setup client header to use TLS 1.2
	tr := &http.Transport{
		// Reads PROXY configuration from environment variables
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	// Needed due to custom client being leveraged, otherwise HTTP2 will not be used.
	tr.ForceAttemptHTTP2 = true

	// Create the client
	return &http.Client{Transport: tr}
}

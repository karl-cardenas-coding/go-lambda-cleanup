package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current version number of disaster-cli",
	Long:  `Prints the current version number of disaster-cli`,
	Run: func(cmd *cobra.Command, args []string) {
		version := fmt.Sprintf("go-lambda-cleanup %s", VersionString)
		fmt.Println(version)
	},
}

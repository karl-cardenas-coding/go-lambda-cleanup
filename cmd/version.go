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
	Short: "Print the current version number of go-lambda-cleanup",
	Long:  `Prints the current version number of go-lambda-cleanup`,
	Run: func(cmd *cobra.Command, args []string) {
		version := fmt.Sprintf("go-lambda-cleanup %s", VersionString)
		fmt.Println(version)
	},
}

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information - these can be overridden during build
var (
	Version = "1.0.0"
	Commit  = "none"
	Date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of ksau-go",
	Long:  `All software has versions. This is ksau-go's.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ksau-go v%s\n", Version)
		fmt.Printf("Commit: %s\n", Commit)
		fmt.Printf("Built: %s\n", Date)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

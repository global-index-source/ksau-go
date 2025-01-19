package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ksau-go",
	Short: "A CLI tool for OneDrive file operations",
	Long: `ksau-go is a command line tool for performing OneDrive operations
like uploading files and checking quota information across multiple
OneDrive configurations.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("remote-config", "c", "oned", "Name of the remote configuration section in rclone.conf")
}

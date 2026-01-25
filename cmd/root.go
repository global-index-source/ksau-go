package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	restarted bool
	rootCmd = &cobra.Command{
		Use:   "ksau-go",
		Short: "A CLI tool for OneDrive file operations",
		Long: `ksau-go is a command line tool for performing OneDrive operations
like uploading files and checking quota information across multiple
OneDrive configurations.`,
	}
)

func Execute() {
	// Initialize auto-update checker if not a restart command
	if !contains(os.Args, "update") && !restarted {
		checker := NewAutoUpdateChecker()
		checker.CheckForUpdatesInBackground(Version)
		
		// Start periodic checks for long-running processes
		go checker.SchedulePeriodicChecks(Version)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.PersistentFlags().StringP("remote-config", "c", "", "Name of the remote configuration section in rclone.conf")
	rootCmd.PersistentFlags().BoolVar(&restarted, "restarted", false, "Internal flag to detect restart after update")
}

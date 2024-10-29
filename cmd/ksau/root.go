package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ksau",
	Short: "A versatile tool for managing file uploads and remote configurations",
	Long: `ksau is a command-line application designed to simplify file uploads and manage remote configurations.
It provides various commands to upload files, list available remotes, refresh configurations, and update the tool itself.
Use the 'ksau help' command to see detailed usage information and examples.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			helpCmd.Run(cmd, args) // No subcommand provided, show help message
			return                 // Help message printed, exit
		}

		// Subcommands logic is handled in the subcommands themselves
		// This is to allow for better organization and separation of concerns
	},
}

func init() {
	// Disable cobra's default help command
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	// Silence usage errors
	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
}

// Execute initializes the CLI and adds all subcommands
func Execute() {
	// Add subcommands here
	// rootCmd.AddCommand(uploadCmd) // Example
	rootCmd.AddCommand(helpCmd) // Help command

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Invalid command. Use 'ksau help' for usage information.")
		os.Exit(1)
	}
}

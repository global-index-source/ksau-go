// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/global-index-source/ksau-go/internal"

	"github.com/goh-chunlin/go-onedrive/onedrive"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var executableName string = internal.GetExecutableName()

var rootCmd = &cobra.Command{
	Use:   "ksau",
	Short: "A versatile tool for managing file uploads and remote configurations",
	Long: fmt.Sprintf(`ksau is a command-line application designed to simplify file uploads and manage remote configurations.
It provides various commands to upload files, list available remotes, refresh configurations, and update the tool itself.
Use the '%s help' command to see detailed usage information and examples.`, executableName),
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

var invalidCommandMessage string = fmt.Sprintf("Invalid command. Use '%s help' for usage information.", executableName)

// Execute initializes the CLI and adds all subcommands
func Execute() {
	// Add subcommands here
	// rootCmd.AddCommand(uploadCmd) // Example
	rootCmd.AddCommand(helpCmd)    // Help command
	rootCmd.AddCommand(refreshCmd) // refresh command
	rootCmd.AddCommand(uploadCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(invalidCommandMessage)
		os.Exit(1)
	}
}

func getClientAndContext(remote *internal.Remote) (*onedrive.Client, *context.Context) {
	var ctx context.Context = context.Background()
	var tokenSource oauth2.TokenSource = oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken:  remote.AccessToken,
			TokenType:    remote.TokenType,
			RefreshToken: remote.RefreshToken,
		},
	)

	var tokenClient *http.Client = oauth2.NewClient(ctx, tokenSource)
	var client onedrive.Client = *onedrive.NewClient(tokenClient)

	return &client, &ctx
}

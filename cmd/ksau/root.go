package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var helpMessage string = `
Usage : ksau [-r] [OPTION] [FILE]

Options:
    -r    Add random string to the end of filename (but before extension) when uploading
    -q    Suppress all output, printing only the link after upload is finished
    -c    Upload to a specific remote

    Note that all options must be passed BEFORE other arguments.

upload [-r] [FILE] [FOLDER] : Uploads the given file to the given folder
                              on index.

list                        : List available remotes and show their usage info.

refresh                     : Refresh rclone config. Try this if upload does not work.

update                      : Fetch and install latest version.
                              available.

help                        : Show this message.

version                     : Show ksau version.

Example: ksau upload test.txt Public
Note: Each time ksau is run, even if not for uploading, it will attempt to refresh
      the remote with the most free space cache. This might cause a few seconds of delay,
      depending on the internet connection.

Tool By Sauraj, Hakimi, and Pratham
Join our Telegram channel for updates:
    https://t.me/ksau_update
`

var rootCmd = &cobra.Command{
	Use:   "ksau",
	Short: "A versatile tool for managing file uploads and remote configurations",
	Long: `ksau is a command-line application designed to simplify file uploads and manage remote configurations.
It provides various commands to upload files, list available remotes, refresh configurations, and update the tool itself.
Use the 'ksau help' command to see detailed usage information and examples.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println(helpMessage)
			return // Help message printed, exit
		}

		// Subcommands logic is handled in the subcommands themselves
		// This is to allow for better organization and separation of concerns

		fmt.Println("Invalid command. Use 'ksau help' for usage information.")
	},
}

// Execute initializes the CLI and adds all subcommands
func Execute() {
	// Add subcommands here
	// rootCmd.AddCommand(uploadCmd) // Example

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var helpMessage string = fmt.Sprintf(`
Usage : %s [-rqc] [SUBCOMMAND]

Options:
    -r    Add random string to the end of filename (but before extension) when uploading
    -q    Suppress all output, printing only the link after upload is finished
    -c    Upload to a specific remote

Subcommands:
upload [-rqc] <FILE> [FOLDER]: Uploads the given file to the given folder
                               on index.

list                         : List available remotes and show their usage info.

refresh                      : Refresh rclone config. Try this if upload does not work.

update                       : Fetch and install latest version.
                               available.

help                         : Show this message.

version                      : Show ksau version.

Example: %s upload test.txt Public
Note: Each time ksau is run, even if not for uploading, it will attempt to refresh
      the remote with the most free space cache. This might cause a few seconds of delay,
      depending on the internet connection.

Tool By Sauraj, Hakimi, and Pratham
Join our Telegram channel for updates:
    https://t.me/ksau_update
`, executableName, executableName)

var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "Display detailed usage information for ksau",
	Long: `The 'ksau help' command provides comprehensive information on how to use the ksau tool.
It includes details on various commands, options, and examples to help users effectively manage file uploads and remote configurations.
Use this command to understand the full capabilities of ksau and how to utilize them.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(helpMessage)
	},
}

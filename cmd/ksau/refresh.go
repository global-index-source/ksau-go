// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"

	"github.com/global-index-source/ksau-go/internal"

	"github.com/spf13/cobra"
)

// GIST_URL is the URL to download the latest tokens.
// This URL should point to a JSON file containing the tokens.
const GIST_URL string = "https://gist.githubusercontent.com/hakimifr/d964b07dff17ff7dc80d39c93ecdfdeb/raw/ksau.json.asc"

func refreshCmdCallback(cmd *cobra.Command, args []string) {
	fmt.Println("Updating tokens...")

	var err error = internal.DownloadConfig(GIST_URL)
	if err != nil {
		panic("Error while updating tokens: " + err.Error())
	}

	fmt.Println("Tokens updated successfully.") // Make sure to let the user know that the tokens have been updated.
}

var refreshCmd *cobra.Command = &cobra.Command{
	Use:   "refresh",
	Short: "Download latest token for each remote",
	Long:  `Download latest token for each remote. Try this if other commands fails to run.`,

	Run: refreshCmdCallback,
}

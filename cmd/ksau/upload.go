// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	// "github.com/goh-chunlin/go-onedrive/onedrive"
	// "golang.org/x/oauth2"

	"fmt"

	"github.com/global-index-source/ksau-go/internal"
	"github.com/spf13/cobra"
)

func uploadCmdCallback(cmd *cobra.Command, args []string) {
	if len(args) > 3 {
		fmt.Printf("%s: Cannot have more than three arguments", internal.GetFunctionName())
		return
	}

	if len(args) == 2 {
		args = append(args, "")
	}

	remote, err := internal.GetRemote(args[0])
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}

	client, ctx := getClientAndContext(remote)
	item, err := client.DriveItems.UploadNewFile(*ctx, remote.DriveId, remote.Prefix+"/"+args[2], args[1])
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	fmt.Printf("item: %s\nmime type: %s\nfile id: %s\n", item.Name, item.File.MIMEType, item.Id)
}

var uploadCmd *cobra.Command = &cobra.Command{
	Use:   "upload",
	Short: "Upload a file",
	Long:  "Upload a file. Optionally to a folder, if provided.",

	Run: uploadCmdCallback,
}

// ctx := context.Background()
// ts := oauth2.StaticTokenSource(
// 	&oauth2.Token{AccessToken: "..."},
// )
// tc := oauth2.NewClient(ctx, ts)

// client := onedrive.NewClient(tc)

// // list all OneDrive drives for the current logged in user
// drives, err := client.Drives.List(ctx)

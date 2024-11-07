// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/global-index-source/ksau-go/internal"

	"github.com/spf13/cobra"
)

var randomFlag bool

func init() {
	uploadCmd.Flags().BoolVarP(&randomFlag, "random", "r", false, "Generate a random file name")
}

func uploadCmdCallback(cmd *cobra.Command, args []string) {
	if len(args) > 3 {
		fmt.Printf("%s: Cannot have more than three arguments", internal.GetFunctionName())
		return
	}

	if len(args) == 2 {
		args = append(args, "")
	}

	// Extract the args into separate variables
	remoteName := args[0]
	filePath := args[1]
	folderPath := args[2]

	// Get the filename from the path
	fileName := filepath.Base(filePath)

	// Check if the file exists
	if fileName == "." || fileName == "/" {
		fmt.Printf("error: %s\n", "Invalid file path")
		return
	}

	// Randomize the file name
	// Syntax: <filename>-<random>.<extension>
	if randomFlag {
		randomString := internal.GenerateRandomString(5)
		ext := filepath.Ext(fileName)
		base := fileName[:len(fileName)-len(ext)]

		fileName = fmt.Sprintf("%s-%s%s", base, randomString, ext)
	}

	remote, err := internal.GetRemote(remoteName)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}

	uploadPath := filepath.Join(remote.Prefix, folderPath)

	// Rename the file if randomFlag is set
	if randomFlag {
		uploadPath = filepath.Join(uploadPath, fileName)
	}

	client, ctx := getClientAndContext(remote)
	item, err := client.DriveItems.UploadNewFile(*ctx, remote.DriveId, uploadPath, filePath)
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

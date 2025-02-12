package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

const UPDATE_URL string = "https://raw.githubusercontent.com/global-index-source/ksau-go/refs/heads/master/updateMetadata.json"

var updateCustomUrl string

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update ksau-go",
	Long:  `Update to the latest ksau-go version available. You can specify custom update metadata url.`,
	Run:   runUpdate,
}

const (
	UpdateTypeMajor string = "major"
	UpdateTypeMinor string = "minor"
	UpdateTypePatch string = "patch"
)

type UpdateMetadata struct {
	Major        int    `json:"major"`
	Minor        int    `json:"minor"`
	Patch        int    `json:"patch"`
	UpdateType   string `json:"update_type"`
	UpdateString string `json:"update_string"`
	UpdateUrl    string `json:"update_url"`
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringVarP(&updateCustomUrl, "url", "u", "", "Sets a custom update metadata url (must be direct.)")
}

func runUpdate(cmd *cobra.Command, args []string) {
	var targetUrl string
	if updateCustomUrl != "" {
		targetUrl = updateCustomUrl
	} else {
		targetUrl = DEFAULT_URL
	}

	fmt.Println("fetching rclone config from", targetUrl)
	resp, err := http.Get(targetUrl)
	if err != nil {
		fmt.Println("failed to fetch config file:", err.Error())
		os.Exit(1)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("something is wrong with the response:", err.Error())
		os.Exit(1)
	}

	userConfigFilePath, err := getConfigPath()
	if err != nil {
		fmt.Println("cannot get your rclone config file path:", err.Error())
	}

	fmt.Println("writing config file to", userConfigFilePath)
	err = os.WriteFile(userConfigFilePath, body, 0644)
	if err != nil {
		fmt.Println("cannot write to your config file:", err.Error())
	}
}

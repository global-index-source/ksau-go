package cmd

import (
	"encoding/json"
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

func download(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("download failed for url '%s': %v", url, err)
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return body, nil
}

func runUpdate(cmd *cobra.Command, args []string) {
	var targetUrl string
	var updateMetadata *UpdateMetadata = new(UpdateMetadata)

	if updateCustomUrl != "" {
		targetUrl = updateCustomUrl
	} else {
		targetUrl = DEFAULT_URL
	}

	body, err := download(targetUrl)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	err = json.Unmarshal(body, updateMetadata)
	if err != nil {
		fmt.Printf("failed to unmarshal update metadata: %v", err)
		os.Exit(1)
	}
}

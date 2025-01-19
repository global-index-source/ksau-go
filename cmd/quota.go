package cmd

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ksauraj/ksau-oned-api/azure"
	"github.com/spf13/cobra"
)

var quotaCmd = &cobra.Command{
	Use:   "quota",
	Short: "Display OneDrive quota information",
	Long:  `Display quota information for all configured OneDrive remotes.`,
	Run:   runQuota,
}

func init() {
	rootCmd.AddCommand(quotaCmd)
}

func runQuota(cmd *cobra.Command, args []string) {
	// Read the rclone config file
	configData, err := getConfigData()
	if err != nil {
		fmt.Println("Failed to read config file:", err)
		return
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	for remote := range rootFolders {
		client, err := azure.NewAzureClientFromRcloneConfigData(configData, remote)
		if err != nil {
			fmt.Printf("Failed to initialize client for remote '%s': %v\n", remote, err)
			continue
		}

		quota, err := client.GetDriveQuota(httpClient)
		if err != nil {
			fmt.Printf("Failed to fetch quota information for remote '%s': %v\n", remote, err)
			continue
		}

		azure.DisplayQuotaInfo(remote, quota)
	}
}

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
		fmt.Println("Failed to read config file:", err.Error())
		return
	}

	rcloneConfigFile, err := azure.ParseRcloneConfigData(configData)
	if err != nil {
		fmt.Println("Failed to parse rclone config file:", err.Error())
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}

	availRemotes := azure.GetAvailableRemotes(&rcloneConfigFile)

	for _, remoteName := range availRemotes {
		fmt.Println("current remote:", remoteName)
		client, err := azure.NewAzureClientFromRcloneConfigData(configData, remoteName)
		if err != nil {
			fmt.Printf("Failed to initialize client for remote '%s': %v\n", remoteName, err)
			continue
		}

		quota, err := client.GetDriveQuota(httpClient)
		if err != nil {
			fmt.Printf("Failed to fetch quota information for remote '%s': %v\n", remoteName, err)
			continue
		}

		azure.DisplayQuotaInfo(remoteName, quota)
	}
}

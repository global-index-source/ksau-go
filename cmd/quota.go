package cmd

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/global-index-source/ksau-go/azure"
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

	var wg = new(sync.WaitGroup)

	for _, remoteName := range availRemotes {
		wg.Add(1)
		go func(rName string) {
			client, err := azure.NewAzureClientFromRcloneConfigData(configData, rName)
			if err != nil {
				fmt.Printf("Failed to initialize client for remote '%s': %v\n", rName, err)
				return
			}

			quota, err := client.GetDriveQuota(httpClient)
			if err != nil {
				fmt.Printf("Failed to fetch quota information for remote '%s': %v\n", rName, err)
				return
			}

			azure.DisplayQuotaInfo(remoteName, quota)
			wg.Done()
		}(remoteName)
	}

	wg.Wait()
}

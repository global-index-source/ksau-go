package cmd

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
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
	configData, _, err := getConfigData()
	if err != nil {
		fmt.Println("Failed to read config file:", err.Error())
		return
	}

	rcloneConfigFile, err := azure.ParseRcloneConfigData(configData)
	if err != nil {
		fmt.Println("Failed to parse rclone config file:", err.Error())
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	exitCode := atomic.Int32{}

	availRemotes := azure.GetAvailableRemotes(&rcloneConfigFile)

	var wg = new(sync.WaitGroup)

	for _, remoteName := range availRemotes {
		wg.Add(1)
		go func(rName string) {
			defer wg.Done()
			client, err := azure.NewAzureClientFromRcloneConfigData(configData, rName)

			if err != nil {
				fmt.Printf("Failed to initialize client for remote '%s': %v\n", rName, err)
				exitCode.Store(1)
				return
			}

			quota, err := client.GetDriveQuota(httpClient)
			if err != nil {
				fmt.Printf("Failed to fetch quota information for remote '%s': %v\n", rName, err)
				exitCode.Store(1)
				return
			}

			azure.DisplayQuotaInfo(remoteName, quota)
		}(remoteName)
	}

	wg.Wait()

	if ec := exitCode.Load(); ec != 0 {
		os.Exit(int(ec))
	}
}

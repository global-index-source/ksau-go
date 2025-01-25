package cmd

import (
	"fmt"
	"os"

	"github.com/global-index-source/ksau-go/azure"
	"github.com/spf13/cobra"
)

var listRemote = &cobra.Command{
	Use:   "list-remote",
	Short: "List available remotes from the configuration file.",
	Long:  "List all available remotes from the configuration file. If the command fails, run refresh.",
	Run:   runListRemote,
}

func init() {
	rootCmd.AddCommand(listRemote)
}

func runListRemote(cmd *cobra.Command, args []string) {
	fmt.Println("reading configuration file...")

	configData, err := getConfigData()
	if err != nil {
		fmt.Println("failed to get configuration file data:", err.Error())
		os.Exit(1)
	}

	parsedConfigData, err := azure.ParseRcloneConfigData(configData)
	if err != nil {
		fmt.Println("failed to parse configuration file data:", err.Error())
		os.Exit(1)
	}

	availableRemotes := azure.GetAvailableRemotes(&parsedConfigData)
	fmt.Println("available remotes:", availableRemotes)
}

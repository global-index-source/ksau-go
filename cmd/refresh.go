package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

const DEFAULT_URL string = "https://gist.githubusercontent.com/hakimifr/34c579f9a35c9da400e4df1ac73cf795/raw/rclone.conf.asc"

var customUrl string

var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh rclone config file",
	Long: `Refresh rclone config. Use this if you get errors
when using this tool, as this possibly fetches renewed tokens.`,
	Run: runRefresh,
}

func init() {
	rootCmd.AddCommand(refreshCmd)

	refreshCmd.Flags().StringVarP(&customUrl, "url", "u", "", "Sets a custom url (must be direct.)")
}

func runRefresh(cmd *cobra.Command, args []string) {
	var targetUrl string
	if customUrl != "" {
		targetUrl = customUrl
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

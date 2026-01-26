package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/global-index-source/ksau-go/internal/config"
	"github.com/spf13/cobra"
)

const (
	DEFAULT_URL         string        = "https://raw.githubusercontent.com/global-index-source/creds/main/rclone.conf.asc"
	defaultRefreshLimit int64         = 5 << 20
	defaultRefreshDelay time.Duration = 15 * time.Second
)

var (
	customUrl    string
	refreshLimit int64
	refreshDelay time.Duration
)

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
	refreshCmd.Flags().Int64Var(&refreshLimit, "max-bytes", defaultRefreshLimit, "Maximum allowed config size in bytes")
	refreshCmd.Flags().DurationVar(&refreshDelay, "timeout", defaultRefreshDelay, "HTTP timeout for download")
}

func runRefresh(cmd *cobra.Command, args []string) {
	var targetUrl string
	if customUrl != "" {
		targetUrl = customUrl
	} else {
		targetUrl = DEFAULT_URL
	}

	fmt.Println("fetching rclone config from", targetUrl)
	client := &http.Client{Timeout: refreshDelay}
	resp, err := client.Get(targetUrl)
	if err != nil {
		fmt.Println("failed to fetch config file:", err.Error())
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("failed to fetch config file: status %d\n", resp.StatusCode)
		os.Exit(1)
	}

	limitedReader := io.LimitReader(resp.Body, refreshLimit+1)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		fmt.Println("something is wrong with the response:", err.Error())
		os.Exit(1)
	}
	if int64(len(body)) > refreshLimit {
		fmt.Println("downloaded config exceeds size limit")
		os.Exit(1)
	}

	downloadFormat, err := config.DetectFormat(body)
	if err != nil {
		fmt.Println("failed to detect config format:", err)
		os.Exit(1)
	}

	userConfigFilePath, err := getConfigPath()
	if err != nil {
		fmt.Println("cannot get your rclone config file path:", err.Error())
		os.Exit(1)
	}

	var existingFormat *config.Format
	if _, format, err := config.ReadConfigFile(userConfigFilePath); err == nil {
		existingFormat = &format
	}

	writeFormat, err := config.ResolveWriteFormat(configFormat, existingFormat)
	if err != nil {
		fmt.Println("cannot determine config format:", err)
		os.Exit(1)
	}

	payload := body
	if downloadFormat != writeFormat {
		decoded, err := config.DecodeConfig(body, downloadFormat)
		if err != nil {
			fmt.Println("failed to decode downloaded config:", err)
			os.Exit(1)
		}
		payload, err = config.EncodeConfig(decoded, writeFormat)
		if err != nil {
			fmt.Println("failed to encode config:", err)
			os.Exit(1)
		}
	}

	fmt.Println("writing config file to", userConfigFilePath)
	if err := config.WriteConfigFile(userConfigFilePath, payload); err != nil {
		fmt.Println("cannot write to your config file:", err.Error())
		os.Exit(1)
	}
}

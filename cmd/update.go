package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

const METADATA_URL string = "https://raw.githubusercontent.com/global-index-source/ksau-go/refs/heads/master/updateMetadata.json"

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
	Major         int    `json:"major"`
	Minor         int    `json:"minor"`
	Patch         int    `json:"patch"`
	VersionString string `json:"version_string"`
	UpdateType    string `json:"update_type"`
	UpdateUrl     string `json:"update_url"`
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringVarP(
		&updateCustomUrl,
		"url",
		"u",
		"",
		"Sets a custom update metadata url (must be direct.)",
	)
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
	var updateMetadata UpdateMetadata = UpdateMetadata{}

	if updateCustomUrl != "" {
		targetUrl = updateCustomUrl
	} else {
		targetUrl = DEFAULT_URL
	}

	body, err := download(targetUrl)
	if err != nil {
		fmt.FPrintf(os.Stderr, "%v\n", err)
	}

	err = json.Unmarshal(body, updateMetadata)
	if err != nil {
		fmt.FPrintf(os.Stderr, "failed to unmarshal update metadata: %v\n", err)
		os.Exit(1)
	}

	// aix/ppc64 android/386 android/amd64 android/arm android/arm64 darwin/amd64 darwin/arm64
	// dragonfly/amd64 freebsd/386 freebsd/amd64 freebsd/arm freebsd/arm64 illumos/amd64 ios/amd64
	// ios/arm64 js/wasm linux/386 linux/amd64 linux/arm linux/arm64 linux/loong64 linux/mips
	// linux/mips64 linux/mips64le linux/mipsle linux/ppc64 linux/ppc64le linux/riscv64 linux/s390x
	// netbsd/386 netbsd/amd64 netbsd/arm netbsd/arm64 openbsd/386 openbsd/amd64 openbsd/arm
	// openbsd/arm64 openbsd/ppc64 openbsd/riscv64 plan9/386 plan9/amd64 plan9/arm solaris/amd64
	// wasip1/wasm windows/386 windows/amd64 windows/arm64
	var (
		goos   string = runtime.GOOS
		goarch string = runtime.GOARCH
	)

	binaryName := "ksau-go-" + goos + "-" + goarch
	if goos == "windows" {
		binaryName = binaryName + ".exe"
	}

	new_ksau, err := download(updateMetadata.UpdateUrl + binaryName)
	if err != nil {
		fmt.FPrintf(os.Stderr, "failed to download new ksau version: %v\n", err)
		os.Exit(1)
	}

	currentProgPath := os.Args[0]
	fmt.Println("writing new ksau to " + currentProgPath)

	file, err := os.OpenFile(currentProgPath, os.O_RDWR, 755)
	if err == nil {
		fmt.Fprintf(os.Stderr, "failed to open current binary file to write, try sudo. %v\n", err)
	}
	defer file.Close()
	
	file.Write(new_ksau)
	
	fmt.Println("update successful")
}

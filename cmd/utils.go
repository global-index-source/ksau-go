package cmd

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ksauraj/ksau-oned-api/azure"
	"github.com/rclone/rclone/backend/onedrive/quickxorhash"
)

// Constants for dynamic chunk size selection
const (
	smallFileSize  = 100 * 1024 * 1024  // 100 MB
	mediumFileSize = 500 * 1024 * 1024  // 500 MB
	largeFileSize  = 1024 * 1024 * 1024 // 1 GB
)

// ANSI color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorRed    = "\033[31m"
)

// Root folders for each remote configuration
var rootFolders = map[string]string{
	"hakimionedrive": "Public",
	"oned":           "",
	"saurajcf":       "MY_BOMT_STUFFS",
}

// Base URLs for each remote configuration
var baseURLs = map[string]string{
	"hakimionedrive": "https://onedrive-vercel-index-kohl-eight-30.vercel.app",
	"oned":           "https://index.sauraj.eu.org",
	"saurajcf":       "https://my-index-azure.vercel.app",
}

// getConfigData reads the rclone.conf file from the user's home directory
func getConfigData() ([]byte, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}

	configPath := filepath.Join(home, ".config", "rclone", "rclone.conf")
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Try Windows path if the Unix path fails
		configPath = filepath.Join(home, "AppData", "Roaming", "rclone", "rclone.conf")
		data, err = os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read rclone config: %v", err)
		}
	}

	return data, nil
}

// getChunkSize dynamically selects a chunk size based on the file size
func getChunkSize(fileSize int64) int64 {
	switch {
	case fileSize <= smallFileSize:
		return 2 * 1024 * 1024 // 2 MB for small files
	case fileSize <= mediumFileSize:
		return 4 * 1024 * 1024 // 4 MB for medium files
	case fileSize <= largeFileSize:
		return 8 * 1024 * 1024 // 8 MB for large files
	default:
		return 16 * 1024 * 1024 // 16 MB for very large files
	}
}

// getRootFolder returns the root folder for a given remote configuration
func getRootFolder(remoteConfig string) string {
	rootFolder, exists := rootFolders[remoteConfig]
	if !exists {
		fmt.Printf("Error: no root folder defined for remote-config '%s'\n", remoteConfig)
		return ""
	}
	return rootFolder
}

// getBaseURL returns the base URL for a given remote configuration
func getBaseURL(remoteConfig string) string {
	baseURL, exists := baseURLs[remoteConfig]
	if !exists {
		fmt.Printf("Error: no base URL defined for remote-config '%s'\n", remoteConfig)
		return ""
	}
	return baseURL
}

// QuickXorHash calculates the QuickXorHash for a file
func QuickXorHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	hash := quickxorhash.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate hash: %v", err)
	}

	hashBytes := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(hashBytes), nil
}

// verifyFileIntegrity verifies the uploaded file's integrity using QuickXorHash
func verifyFileIntegrity(filePath string, fileID string, client *azure.AzureClient, httpClient *http.Client) {
	fmt.Println("Verifying file integrity...")

	localHash, err := QuickXorHash(filePath)
	if err != nil {
		fmt.Printf("Failed to calculate local QuickXorHash: %v\n", err)
		return
	}

	remoteHash, err := getQuickXorHashWithRetry(client, httpClient, fileID, hashRetries, hashRetryDelay)
	if err != nil {
		fmt.Printf("Failed to retrieve remote QuickXorHash: %v\n", err)
		return
	}

	if localHash != remoteHash {
		fmt.Printf("%sQuickXorHash mismatch: File integrity verification failed.%s\n", ColorRed, ColorReset)
	} else {
		fmt.Printf("%sQuickXorHash match: File integrity verified.%s\n", ColorGreen, ColorReset)
	}
}

// getQuickXorHashWithRetry retries fetching the quickXorHash until it succeeds or max retries are reached
func getQuickXorHashWithRetry(client *azure.AzureClient, httpClient *http.Client, fileID string, maxRetries int, retryDelay time.Duration) (string, error) {
	for retry := 0; retry < maxRetries; retry++ {
		remoteHash, err := client.GetQuickXorHash(httpClient, fileID)
		if err == nil {
			return remoteHash, nil
		}

		fmt.Printf("Attempt %d/%d: Failed to retrieve remote QuickXorHash: %v\n", retry+1, maxRetries, err)
		time.Sleep(retryDelay)
	}

	return "", fmt.Errorf("failed to retrieve remote QuickXorHash after %d retries", maxRetries)
}

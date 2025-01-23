package cmd

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"time"

	"github.com/global-index-source/ksau-go/azure"
	"github.com/global-index-source/ksau-go/crypto"

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

// getConfigPath: get user's rclone config path. This function is not responsible
// for checking if the file actually exist, and instead only returns OS-specific path.
func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %w", err)
	}

	var configPath string
	// first let's determine what kind of OS we're in
	if slices.Contains([]string{"android", "linux", "unix"}, runtime.GOOS) {
		configPath = filepath.Join(home, ".config", "rclone", "rclone.conf")
	} else if runtime.GOOS == "windows" {
		configPath = filepath.Join(home, "AppData", "Roaming", "rclone", "rclone.conf")
	} else {
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	return configPath, nil
}

// getConfigData reads the rclone.conf file from the user's home directory
func getConfigData() ([]byte, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get user's config file path: %w", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rclone config: %v", err)
	}

	return crypto.Decrypt(data), nil
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

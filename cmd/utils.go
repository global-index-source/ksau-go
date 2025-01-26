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
)

// ANSI color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
)

func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %w", err)
	}

	var configDir string
	if slices.Contains([]string{"android", "linux", "unix"}, runtime.GOOS) {
		configDir = filepath.Join(home, ".ksau", ".conf")
	} else if runtime.GOOS == "windows" {
		configDir = filepath.Join(home, "AppData", "Roaming", "ksau", ".conf")
	} else {
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	// Create directories if they don't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "rclone.conf")
	return configPath, nil
}

func getConfigData() ([]byte, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config path: %w", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	decryptedConfig, err := crypto.Decrypt(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt user's config file: %w", err)
	}
	return decryptedConfig, nil

}

func getChunkSize(fileSize int64) int64 {
	// Default chunk sizes based on file size
	switch {
	case fileSize < 1024*1024*64: // < 64MB
		return 1024 * 1024 * 4 // 4MB chunks
	case fileSize < 1024*1024*256: // < 256MB
		return 1024 * 1024 * 8 // 8MB chunks
	case fileSize < 1024*1024*1024: // < 1GB
		return 1024 * 1024 * 16 // 16MB chunks
	default:
		return 1024 * 1024 * 32 // 32MB chunks
	}
}

func verifyFileIntegrity(filePath string, fileID string, client *azure.AzureClient, httpClient *http.Client) {
	fmt.Println("Verifying file integrity...")

	var fileHash string
	var err error

	// Retry getting the file hash
	for i := 0; i < hashRetries; i++ {
		fileHash, err = client.GetQuickXorHash(httpClient, fileID)
		if err == nil {
			break
		}
		fmt.Printf("Attempt %d/%d: Failed to get file hash: %v\n", i+1, hashRetries, err)
		if i < hashRetries-1 {
			time.Sleep(hashRetryDelay)
		}
	}

	if err != nil {
		fmt.Printf("%sWarning: Could not verify file integrity: %v%s\n", ColorYellow, err, ColorReset)
		return
	}

	// Calculate local file hash
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("%sWarning: Could not open local file for verification: %v%s\n", ColorYellow, err, ColorReset)
		return
	}
	defer file.Close()

	// Create new quickXorHash instance
	hasher := crypto.New()

	// Copy the file content into the hash
	if _, err := io.Copy(hasher, file); err != nil {
		fmt.Printf("%sWarning: Could not calculate file hash: %v%s\n", ColorYellow, err, ColorReset)
		return
	}

	// Get the hash as a Base64-encoded string
	hashBytes := hasher.Sum(nil)
	localHash := base64.StdEncoding.EncodeToString(hashBytes)

	// fmt.Printf("Local file hash: %s\n", localHash)
	// fmt.Printf("Remote file hash: %s\n", fileHash)

	if localHash == fileHash {
		fmt.Printf("%sFile integrity verified successfully%s\n", ColorGreen, ColorReset)
	} else {
		fmt.Printf("%sWarning: File integrity check failed - hashes do not match%s\n", ColorRed, ColorReset)
	}
}

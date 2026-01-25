package cmd

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
	Prerelease bool `json:"prerelease"`
	Assets  []Asset `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
	Digest             string `json:"digest"`
}

// UpdateConfig holds configuration for the update process
type UpdateConfig struct {
	RepoOwner      string
	RepoName       string
	CurrentVersion string
	CheckInterval  time.Duration
	AutoUpdate     bool
}

var updateCmd = &cobra.Command{
	Use:   "self-update",
	Short: "Update ksau-go to the latest version",
	Long: `Check for updates and automatically upgrade ksau-go to the latest version.
This command downloads the latest release from GitHub, verifies its integrity,
and replaces the current binary with the new version.`,
	RunE: runSelfUpdate,
}

var (
	checkOnly      bool
	forceUpdate    bool
	prerelease     bool
	updateChannel  string
)

func init() {
	rootCmd.AddCommand(updateCmd)
	
	updateCmd.Flags().BoolVar(&checkOnly, "check-only", false, "Only check for updates, don't install")
	updateCmd.Flags().BoolVar(&forceUpdate, "force", false, "Force update even if already on latest version")
	updateCmd.Flags().BoolVar(&prerelease, "prerelease", false, "Include pre-release versions")
	updateCmd.Flags().StringVar(&updateChannel, "channel", "stable", "Update channel (stable, beta, alpha)")
}

func runSelfUpdate(cmd *cobra.Command, args []string) error {
	config := &UpdateConfig{
		RepoOwner:      "global-index-source",
		RepoName:       "ksau-go",
		CurrentVersion: Version,
		CheckInterval:  24 * time.Hour,
		AutoUpdate:     false,
	}

	fmt.Printf("Current version: %s\n", config.CurrentVersion)
	fmt.Println("Checking for updates...")

	latestRelease, err := checkForUpdates(config, prerelease)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if latestRelease == nil {
		fmt.Println("No updates available.")
		return nil
	}

	fmt.Printf("Latest version: %s\n", latestRelease.TagName)
	fmt.Printf("Release notes: %s\n\n", latestRelease.Body)

	if checkOnly {
		fmt.Println("Update available. Run without --check-only to install.")
		return nil
	}

	// Check if we're already on the latest version
	if !forceUpdate && isLatestVersion(config.CurrentVersion, latestRelease.TagName) {
		fmt.Println("You're already on the latest version.")
		return nil
	}

	// Confirm update
	fmt.Printf("Update to version %s? [y/N] ", latestRelease.TagName)
	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println("Update cancelled.")
		return nil
	}

	return performUpdate(latestRelease)
}

func checkForUpdates(config *UpdateConfig, includePrerelease bool) (*GitHubRelease, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Get latest release
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", config.RepoOwner, config.RepoName)
	
	if includePrerelease {
		// Get all releases and filter
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", config.RepoOwner, config.RepoName)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", fmt.Sprintf("ksau-go/%s", config.CurrentVersion))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	if includePrerelease {
		var releases []GitHubRelease
		if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
			return nil, fmt.Errorf("failed to decode releases: %w", err)
		}

		// Find the latest release (including prereleases if requested)
		for _, release := range releases {
			if !includePrerelease && release.Prerelease {
				continue
			}
			return &release, nil
		}
		return nil, nil
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release: %w", err)
	}

	return &release, nil
}

func isLatestVersion(current, latest string) bool {
	// Remove 'v' prefix if present
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")
	
	return current == latest
}

func performUpdate(release *GitHubRelease) error {
	// Determine the correct asset for this platform
	assetName := getAssetName()
	var targetAsset *Asset
	
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, assetName) {
			targetAsset = &asset
			break
		}
	}

	if targetAsset == nil {
		return fmt.Errorf("no suitable binary found for %s-%s", runtime.GOOS, runtime.GOARCH)
	}

	fmt.Printf("Downloading %s...\n", targetAsset.Name)

	// Create temporary directory for update
	tempDir, err := os.MkdirTemp("", "ksau-go-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download the new binary
	downloadPath := filepath.Join(tempDir, targetAsset.Name)
	if err := downloadFile(targetAsset.BrowserDownloadURL, downloadPath); err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}

	// Verify the download
	if err := verifyBinary(downloadPath, targetAsset); err != nil {
		return fmt.Errorf("binary verification failed: %w", err)
	}

	// Extract if it's a tar.gz
	if strings.HasSuffix(targetAsset.Name, ".tar.gz") {
		extractedPath, err := extractTarGz(downloadPath, tempDir)
		if err != nil {
			return fmt.Errorf("failed to extract binary: %w", err)
		}
		downloadPath = extractedPath
	}

	// Make the binary executable
	if err := os.Chmod(downloadPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	// Get current executable path
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	// Create backup before updating
	rollbackManager := NewRollbackManager()
	backupPath, err := rollbackManager.CreateBackup(currentExe)
	if err != nil {
		fmt.Printf("Warning: failed to create backup: %v\n", err)
		backupPath = currentExe + ".backup"
		if err := os.Rename(currentExe, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	} else {
		fmt.Printf("Created backup: %s\n", filepath.Base(backupPath))
	}

	// Atomic update using rename
	if err := atomicUpdate(downloadPath, currentExe, backupPath); err != nil {
		return fmt.Errorf("failed to update binary: %w", err)
	}

	// Clean up temporary backup if it exists
	if backupPath == currentExe+".backup" {
		defer os.Remove(backupPath)
	}

	fmt.Println("Update completed successfully!")
	fmt.Printf("Updated to version %s\n", release.TagName)

	// Clean up old backups
	go func() {
		if err := rollbackManager.CleanupOldBackups(); err != nil {
			fmt.Printf("Warning: failed to cleanup old backups: %v\n", err)
		}
	}()

	// Restart with new binary
	return restartWithNewBinary(currentExe)
}

func getAssetName() string {
	os := runtime.GOOS
	arch := runtime.GOARCH
	
	switch os {
	case "windows":
		return "ksau-go-windows-amd64.exe"
	case "linux":
		switch arch {
		case "amd64":
			return "ksau-go-linux-amd64"
		case "arm64":
			return "ksau-go-linux-arm64"
		default:
			return fmt.Sprintf("ksau-go-linux-%s", arch)
		}
	case "darwin":
		switch arch {
		case "amd64":
			return "ksau-go-darwin-amd64"
		case "arm64":
			return "ksau-go-darwin-arm64"
		default:
			return fmt.Sprintf("ksau-go-darwin-%s", arch)
		}
	default:
		return fmt.Sprintf("ksau-go-%s-%s", os, arch)
	}
}

func downloadFile(url, path string) error {
	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func verifyBinary(path string, asset *Asset) error {
	// Calculate SHA256 hash
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open binary: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("failed to calculate hash: %w", err)
	}

	calculatedHash := fmt.Sprintf("sha256:%x", hasher.Sum(nil))

	// If GitHub provides a digest, verify it
	if asset.Digest != "" {
		if calculatedHash != asset.Digest {
			return fmt.Errorf("hash mismatch: expected %s, got %s", asset.Digest, calculatedHash)
		}
		fmt.Println("Binary signature verified successfully.")
	} else {
		fmt.Printf("Warning: No digest provided by GitHub. Calculated hash: %s\n", calculatedHash)
	}

	return nil
}

func extractTarGz(archivePath, destDir string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	var extractedBinary string
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar entry: %w", err)
		}

		// Look for the ksau-go binary
		if strings.Contains(header.Name, "ksau-go") && !strings.HasSuffix(header.Name, ".md5") {
			extractedPath := filepath.Join(destDir, filepath.Base(header.Name))
			
			outFile, err := os.Create(extractedPath)
			if err != nil {
				return "", fmt.Errorf("failed to create extracted file: %w", err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return "", fmt.Errorf("failed to extract file: %w", err)
			}
			outFile.Close()

			extractedBinary = extractedPath
			break
		}
	}

	if extractedBinary == "" {
		return "", fmt.Errorf("no binary found in archive")
	}

	return extractedBinary, nil
}

func atomicUpdate(newPath, currentPath, backupPath string) error {
	// Create backup of current binary
	if err := os.Rename(currentPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Move new binary to current location
	if err := os.Rename(newPath, currentPath); err != nil {
		// Restore backup if update fails
		os.Rename(backupPath, currentPath)
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	return nil
}

func restartWithNewBinary(exePath string) error {
	fmt.Println("Restarting with new version...")

	// Get current arguments
	args := os.Args[1:] // Skip the program name
	
	// Add a flag to indicate this is a restart
	args = append(args, "--restarted")

	// Start new process
	cmd := exec.Command(exePath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start new process: %w", err)
	}

	// Exit current process
	os.Exit(0)
	return nil
}
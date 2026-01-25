package cmd

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName    string  `json:"tag_name"`
	Name       string  `json:"name"`
	Body       string  `json:"body"`
	Prerelease bool    `json:"prerelease"`
	Draft      bool    `json:"draft"`
	Assets     []Asset `json:"assets"`
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
	RepoOwner    string
	RepoName     string
	CurrentVersion string
	CheckInterval time.Duration
	AutoUpdate   bool
}

var updateCmd = &cobra.Command{
	Use:   "update",
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
	
	// Add check-updates command
	checkUpdatesCmd := &cobra.Command{
		Use:   "check-updates",
		Short: "Check for available updates",
		Long: `Check if a newer version of ksau-go is available.
This command will check for updates and optionally notify you if a new version is found.`,
		RunE: runCheckUpdates,
	}
	rootCmd.AddCommand(checkUpdatesCmd)
	
	// Add config command for update settings
	configCmd := &cobra.Command{
		Use:   "update-config",
		Short: "Manage update configuration",
		Long: `Configure automatic update settings for ksau-go.
This allows you to enable automatic updates and set update preferences.`,
		RunE: runUpdateConfig,
	}
	configCmd.Flags().Bool("enable-auto-update", false, "Enable automatic updates")
	configCmd.Flags().Bool("disable-auto-update", false, "Disable automatic updates")
	configCmd.Flags().String("channel", "", "Set update channel (stable, beta, alpha)")
	configCmd.Flags().Bool("include-prereleases", false, "Include pre-release versions")
	configCmd.Flags().Bool("exclude-prereleases", false, "Exclude pre-release versions")
	rootCmd.AddCommand(configCmd)
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

	// Always get all releases to have better control
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", config.RepoOwner, config.RepoName)

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

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to decode releases: %w", err)
	}

	// Find the latest suitable release
	var latestRelease *GitHubRelease
	for _, release := range releases {
		// Skip drafts
		if release.Draft {
			continue
		}
		
		// Skip prereleases unless requested
		if !includePrerelease && release.Prerelease {
			continue
		}
		
		// If this is the first suitable release or newer than current latest
		if latestRelease == nil || isNewerVersion(release.TagName, latestRelease.TagName) {
			latestRelease = &release
		}
	}

	return latestRelease, nil
}

func isLatestVersion(current, latest string) bool {
	// Remove 'v' prefix if present
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")
	
	return current == latest
}

// isNewerVersion checks if version1 is newer than version2
func isNewerVersion(version1, version2 string) bool {
	// Remove 'v' prefix if present
	version1 = strings.TrimPrefix(version1, "v")
	version2 = strings.TrimPrefix(version2, "v")
	
	// Split version numbers
	v1Parts := strings.Split(version1, ".")
	v2Parts := strings.Split(version2, ".")
	
	// Compare each part
	maxLen := len(v1Parts)
	if len(v2Parts) > maxLen {
		maxLen = len(v2Parts)
	}
	
	for i := 0; i < maxLen; i++ {
		var v1Num, v2Num int
		
		if i < len(v1Parts) {
			if num, err := strconv.Atoi(v1Parts[i]); err == nil {
				v1Num = num
			}
		}
		
		if i < len(v2Parts) {
			if num, err := strconv.Atoi(v2Parts[i]); err == nil {
				v2Num = num
			}
		}
		
		if v1Num > v2Num {
			return true
		} else if v1Num < v2Num {
			return false
		}
	}
	
	return false
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

// testBinary validates that a binary can run and show version
func testBinary(binaryPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("binary test failed: %w, output: %s", err, string(output))
	}

	// Check if output contains version information
	if !strings.Contains(string(output), "ksau-go v") {
		return fmt.Errorf("binary output doesn't contain expected version info: %s", string(output))
	}

	return nil
}

// RollbackManager handles backup and rollback operations
type RollbackManager struct {
	backupDir string
}

// NewRollbackManager creates a new rollback manager
func NewRollbackManager() *RollbackManager {
	homeDir, _ := os.UserHomeDir()
	backupDir := filepath.Join(homeDir, ".ksau", "backups")
	return &RollbackManager{backupDir: backupDir}
}

// CreateBackup creates a backup of the current binary
func (rm *RollbackManager) CreateBackup(currentPath string) (string, error) {
	if err := os.MkdirAll(rm.backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("ksau-go-backup-%s", timestamp)
	backupPath := filepath.Join(rm.backupDir, backupName)

	if err := os.Rename(currentPath, backupPath); err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	return backupPath, nil
}

// CleanupOldBackups removes old backup files (keeps last 5)
func (rm *RollbackManager) CleanupOldBackups() error {
	entries, err := os.ReadDir(rm.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Directory doesn't exist, nothing to clean
		}
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	// Sort entries by modification time (newest first)
	var backupFiles []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && strings.Contains(entry.Name(), "ksau-go-backup-") {
			backupFiles = append(backupFiles, entry)
		}
	}

	// Keep only the 5 most recent backups
	if len(backupFiles) > 5 {
		for i := 5; i < len(backupFiles); i++ {
			backupPath := filepath.Join(rm.backupDir, backupFiles[i].Name())
			if err := os.Remove(backupPath); err != nil {
				fmt.Printf("Warning: failed to remove old backup %s: %v\n", backupFiles[i].Name(), err)
			}
		}
	}

	return nil
}

// runCheckUpdates checks for updates without installing
func runCheckUpdates(cmd *cobra.Command, args []string) error {
	config := &UpdateConfig{
		RepoOwner:      "global-index-source",
		RepoName:       "ksau-go",
		CurrentVersion: Version,
		CheckInterval:  24 * time.Hour,
		AutoUpdate:     false,
	}

	// Load update configuration
	configManager := NewConfigManager()
	updateConfig, err := configManager.LoadConfig()
	if err != nil {
		fmt.Printf("Warning: Could not load update config: %v\n", err)
	} else {
		config.AutoUpdate = updateConfig.AutoUpdate
		config.Prerelease = updateConfig.Prerelease
	}

	fmt.Printf("Current version: %s\n", config.CurrentVersion)
	fmt.Println("Checking for updates...")

	latestRelease, err := checkForUpdates(config, config.Prerelease)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if latestRelease == nil {
		fmt.Println("No updates available.")
		return nil
	}

	fmt.Printf("Latest version: %s\n", latestRelease.TagName)
	
	if isLatestVersion(config.CurrentVersion, latestRelease.TagName) {
		fmt.Println("✓ You're on the latest version.")
	} else {
		fmt.Println("⚠️  A new version is available!")
		fmt.Printf("Release notes: %s\n\n", latestRelease.Body)
		fmt.Println("Run 'ksau-go update' to install the new version.")
	}

	// Update last check time
	if err := configManager.UpdateLastCheck(); err != nil {
		fmt.Printf("Warning: Could not update last check time: %v\n", err)
	}

	return nil
}

// runUpdateConfig manages update configuration
func runUpdateConfig(cmd *cobra.Command, args []string) error {
	configManager := NewConfigManager()
	config, err := configManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("Current update configuration:\n")
	fmt.Printf("  Auto-update: %v\n", config.AutoUpdate)
	fmt.Printf("  Update channel: %s\n", config.UpdateChannel)
	fmt.Printf("  Include pre-releases: %v\n", config.Prerelease)
	fmt.Printf("  Last check: %s\n", config.LastCheck.Format(time.RFC3339))

	// If no arguments, just show current config
	if len(args) == 0 {
		fmt.Println("\nTo modify settings, use:")
		fmt.Println("  ksau-go update-config --enable-auto-update")
		fmt.Println("  ksau-go update-config --disable-auto-update")
		fmt.Println("  ksau-go update-config --channel stable|beta|alpha")
		fmt.Println("  ksau-go update-config --include-prereleases")
		fmt.Println("  ksau-go update-config --exclude-prereleases")
		return nil
	}

	// Process configuration changes
	changed := false
	
	if cmd.Flag("enable-auto-update").Changed {
		config.AutoUpdate = true
		changed = true
		fmt.Println("✓ Auto-update enabled")
	}
	
	if cmd.Flag("disable-auto-update").Changed {
		config.AutoUpdate = false
		changed = true
		fmt.Println("✓ Auto-update disabled")
	}
	
	if cmd.Flag("channel").Changed {
		newChannel, _ := cmd.Flags().GetString("channel")
		if newChannel == "stable" || newChannel == "beta" || newChannel == "alpha" {
			config.UpdateChannel = newChannel
			changed = true
			fmt.Printf("✓ Update channel set to %s\n", newChannel)
		} else {
			return fmt.Errorf("invalid channel. Use stable, beta, or alpha")
		}
	}
	
	if cmd.Flag("include-prereleases").Changed {
		config.Prerelease = true
		changed = true
		fmt.Println("✓ Pre-releases included")
	}
	
	if cmd.Flag("exclude-prereleases").Changed {
		config.Prerelease = false
		changed = true
		fmt.Println("✓ Pre-releases excluded")
	}

	if changed {
		if err := configManager.SaveConfig(config); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Println("Configuration saved successfully.")
	}

	return nil
}

// AutoUpdateChecker handles automatic update checking
type AutoUpdateChecker struct {
	configManager *ConfigManager
}

// NewAutoUpdateChecker creates a new auto-update checker
func NewAutoUpdateChecker() *AutoUpdateChecker {
	return &AutoUpdateChecker{
		configManager: NewConfigManager(),
	}
}

// CheckForUpdatesInBackground checks for updates in the background
func (auc *AutoUpdateChecker) CheckForUpdatesInBackground(currentVersion string) {
	config, err := auc.configManager.LoadConfig()
	if err != nil {
		return // Silently fail
	}

	// Check if we should check for updates
	shouldCheck, err := auc.configManager.ShouldCheck()
	if err != nil || !shouldCheck {
		return
	}

	// Check in background
	go func() {
		updateConfig := &UpdateConfig{
			RepoOwner:      "global-index-source",
			RepoName:       "ksau-go",
			CurrentVersion: currentVersion,
			CheckInterval:  24 * time.Hour,
			AutoUpdate:     config.AutoUpdate,
		}

		latestRelease, err := checkForUpdates(updateConfig, config.Prerelease)
		if err != nil {
			return // Silently fail
		}

		if latestRelease != nil && !isLatestVersion(currentVersion, latestRelease.TagName) {
			fmt.Printf("\n⚠️  A new version of ksau-go is available: %s\n", latestRelease.TagName)
			fmt.Printf("Current version: %s\n", currentVersion)
			fmt.Println("Run 'ksau-go update' to update.")
			if latestRelease.Body != "" {
				fmt.Printf("Release notes:\n%s\n", latestRelease.Body)
			}
			fmt.Println()
		}

		// Update last check time
		auc.configManager.UpdateLastCheck()
	}()
}

// SchedulePeriodicChecks schedules periodic update checks
func (auc *AutoUpdateChecker) SchedulePeriodicChecks(currentVersion string) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			auc.CheckForUpdatesInBackground(currentVersion)
		}
	}
}
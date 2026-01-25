package update

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

// Config holds configuration for the update process
type Config struct {
	RepoOwner     string
	RepoName      string
	CurrentVersion string
	CheckInterval time.Duration
	AutoUpdate    bool
	Prerelease    bool
}

// Updater handles the self-update process
type Updater struct {
	config *Config
	client *http.Client
}

// NewUpdater creates a new updater instance
func NewUpdater(config *Config) *Updater {
	if config == nil {
		config = &Config{
			CheckInterval: 24 * time.Hour,
		}
	}
	
	return &Updater{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CheckForUpdates checks if there are available updates
func (u *Updater) CheckForUpdates(includePrerelease bool) (*GitHubRelease, error) {
	// Get latest release
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", u.config.RepoOwner, u.config.RepoName)
	
	if includePrerelease {
		// Get all releases and filter
		url = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", u.config.RepoOwner, u.config.RepoName)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if u.config.CurrentVersion != "" {
		req.Header.Set("User-Agent", fmt.Sprintf("ksau-go/%s", u.config.CurrentVersion))
	}

	resp, err := u.client.Do(req)
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

// IsLatestVersion checks if the current version is the latest
func (u *Updater) IsLatestVersion(current, latest string) bool {
	// Remove 'v' prefix if present
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")
	
	return current == latest
}

// PerformUpdate downloads and installs the latest release
func (u *Updater) PerformUpdate(release *GitHubRelease) error {
	// Determine the correct asset for this platform
	assetName := u.getAssetName()
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
	if err := u.downloadFile(targetAsset.BrowserDownloadURL, downloadPath); err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}

	// Verify the download
	if err := u.verifyBinary(downloadPath, targetAsset); err != nil {
		return fmt.Errorf("binary verification failed: %w", err)
	}

	// Extract if it's a tar.gz
	if strings.HasSuffix(targetAsset.Name, ".tar.gz") {
		extractedPath, err := u.extractTarGz(downloadPath, tempDir)
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

	// Atomic update using rename
	backupPath := currentExe + ".backup"
	if err := u.atomicUpdate(downloadPath, currentExe, backupPath); err != nil {
		return fmt.Errorf("failed to update binary: %w", err)
	}

	// Clean up backup
	defer os.Remove(backupPath)

	fmt.Println("Update completed successfully!")
	fmt.Printf("Updated to version %s\n", release.TagName)

	// Restart with new binary
	return u.restartWithNewBinary(currentExe)
}

// getAssetName returns the appropriate asset name for the current platform
func (u *Updater) getAssetName() string {
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

// downloadFile downloads a file from URL to path
func (u *Updater) downloadFile(url, path string) error {
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

// verifyBinary verifies the integrity of the downloaded binary
func (u *Updater) verifyBinary(path string, asset *Asset) error {
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

// extractTarGz extracts a tar.gz archive and returns the path to the binary
func (u *Updater) extractTarGz(archivePath, destDir string) (string, error) {
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

// atomicUpdate performs an atomic update of the binary
func (u *Updater) atomicUpdate(newPath, currentPath, backupPath string) error {
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

// restartWithNewBinary restarts the application with the new binary
func (u *Updater) restartWithNewBinary(exePath string) error {
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
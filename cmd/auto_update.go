package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// AutoUpdateChecker handles automatic update checks
type AutoUpdateChecker struct {
	configManager *ConfigManager
}

// NewAutoUpdateChecker creates a new auto-update checker
func NewAutoUpdateChecker() *AutoUpdateChecker {
	return &AutoUpdateChecker{
		configManager: NewConfigManager(),
	}
}

// CheckForUpdatesInBackground checks for updates in background
func (auc *AutoUpdateChecker) CheckForUpdatesInBackground(currentVersion string) {
	shouldCheck, err := auc.configManager.ShouldCheck()
	if err != nil {
		log.Printf("Failed to check update config: %v", err)
		return
	}

	if !shouldCheck {
		return
	}

	// Update the last check timestamp
	if err := auc.configManager.UpdateLastCheck(); err != nil {
		log.Printf("Failed to update last check timestamp: %v", err)
	}

	// Load config to get update preferences
	config, err := auc.configManager.LoadConfig()
	if err != nil {
		log.Printf("Failed to load update config: %v", err)
		return
	}

	// Check for updates
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic in update checker: %v", r)
			}
		}()

		release, err := checkForUpdates(&UpdateConfig{
			RepoOwner:      "global-index-source",
			RepoName:       "ksau-go",
			CurrentVersion: currentVersion,
		}, config.Prerelease)
		
		if err != nil {
			log.Printf("Failed to check for updates: %v", err)
			return
		}

		if release != nil && !isLatestVersion(currentVersion, release.TagName) {
			fmt.Printf("\n🚀 Update available! %s → %s\n", currentVersion, release.TagName)
			fmt.Printf("Run 'ksau-go update' to upgrade.\n")
			
			if len(release.Body) > 100 {
				fmt.Printf("Release notes: %s...\n", strings.TrimSpace(release.Body[:100]))
			} else if len(release.Body) > 0 {
				fmt.Printf("Release notes: %s\n", strings.TrimSpace(release.Body))
			}
		}
	}()
}

// SchedulePeriodicChecks sets up periodic update checking
func (auc *AutoUpdateChecker) SchedulePeriodicChecks(currentVersion string) {
	config, err := auc.configManager.LoadConfig()
	if err != nil {
		return
	}

	// Only set up periodic checks if auto-update is enabled
	if !config.AutoUpdate {
		return
	}

	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		defer ticker.Stop()
		
		for range ticker.C {
			release, err := checkForUpdates(&UpdateConfig{
				RepoOwner:      "global-index-source",
				RepoName:       "ksau-go",
				CurrentVersion: currentVersion,
			}, config.Prerelease)
			
			if err != nil {
				log.Printf("Failed to check for updates: %v", err)
				continue
			}

			if release != nil && !isLatestVersion(currentVersion, release.TagName) {
				if config.AutoUpdate {
					fmt.Printf("🚀 Auto-updating to %s...\n", release.TagName)
					if err := performUpdate(release); err != nil {
						log.Printf("Auto-update failed: %v", err)
					}
				} else {
					fmt.Printf("🚀 Update available! %s → %s\n", currentVersion, release.TagName)
					fmt.Printf("Run 'ksau-go update' to upgrade.\n")
				}
			}

			// Update last check timestamp
			auc.configManager.UpdateLastCheck()
		}
	}()
}
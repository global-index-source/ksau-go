package update

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// Checker handles automatic update checks
type Checker struct {
	manager *Manager
	config  *Config
}

// NewChecker creates a new auto-update checker
func NewChecker(config *Config) *Checker {
	return &Checker{
		manager: NewManager(),
		config:  config,
	}
}

// CheckForUpdatesInBackground checks for updates in background
func (c *Checker) CheckForUpdatesInBackground(currentVersion string) {
	shouldCheck, err := c.manager.ShouldCheck()
	if err != nil {
		log.Printf("Failed to check update config: %v", err)
		return
	}

	if !shouldCheck {
		return
	}

	// Update the last check timestamp
	if err := c.manager.UpdateLastCheck(); err != nil {
		log.Printf("Failed to update last check timestamp: %v", err)
	}

	// Load config to get update preferences
	configFile, err := c.manager.LoadConfig()
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

		updater := NewUpdater(c.config)
		release, err := updater.CheckForUpdates(configFile.Prerelease)
		
		if err != nil {
			log.Printf("Failed to check for updates: %v", err)
			return
		}

		if release != nil && !updater.IsLatestVersion(currentVersion, release.TagName) {
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
func (c *Checker) SchedulePeriodicChecks(currentVersion string) {
	configFile, err := c.manager.LoadConfig()
	if err != nil {
		return
	}

	// Only set up periodic checks if auto-update is enabled
	if !configFile.AutoUpdate {
		return
	}

	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		defer ticker.Stop()
		
		for range ticker.C {
			updater := NewUpdater(c.config)
			release, err := updater.CheckForUpdates(configFile.Prerelease)
			
			if err != nil {
				log.Printf("Failed to check for updates: %v", err)
				continue
			}

			if release != nil && !updater.IsLatestVersion(currentVersion, release.TagName) {
				if configFile.AutoUpdate {
					fmt.Printf("🚀 Auto-updating to %s...\n", release.TagName)
					if err := updater.PerformUpdate(release); err != nil {
						log.Printf("Auto-update failed: %v", err)
					}
				} else {
					fmt.Printf("🚀 Update available! %s → %s\n", currentVersion, release.TagName)
					fmt.Printf("Run 'ksau-go update' to upgrade.\n")
				}
			}

			// Update last check timestamp
			c.manager.UpdateLastCheck()
		}
	}()
}
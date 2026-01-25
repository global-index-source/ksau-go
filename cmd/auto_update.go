package cmd

import (
	"time"

	"github.com/global-index-source/ksau-go/pkg/update"
)

// AutoUpdateChecker handles automatic update checks
type AutoUpdateChecker struct {
	checker *update.Checker
	config  *update.Config
}

// NewAutoUpdateChecker creates a new auto-update checker
func NewAutoUpdateChecker() *AutoUpdateChecker {
	config := &update.Config{
		RepoOwner:      "TronPkr",  // Updated to match the actual repo owner
		RepoName:       "ksau-go",
		CheckInterval:  24 * time.Hour,
		AutoUpdate:     false,
	}
	
	return &AutoUpdateChecker{
		checker: update.NewChecker(config),
		config:  config,
	}
}

// CheckForUpdatesInBackground checks for updates in background
func (auc *AutoUpdateChecker) CheckForUpdatesInBackground(currentVersion string) {
	auc.checker.CheckForUpdatesInBackground(currentVersion)
}

// SchedulePeriodicChecks sets up periodic update checking
func (auc *AutoUpdateChecker) SchedulePeriodicChecks(currentVersion string) {
	auc.checker.SchedulePeriodicChecks(currentVersion)
}
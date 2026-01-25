package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RollbackManager handles rollback operations
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
func (rm *RollbackManager) CreateBackup(currentExe string) (string, error) {
	if err := os.MkdirAll(rm.backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("ksau-go-backup-%s", timestamp)
	backupPath := filepath.Join(rm.backupDir, backupName)

	// Copy current binary to backup location
	if err := copyFile(currentExe, backupPath); err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	// Make backup executable
	if err := os.Chmod(backupPath, 0755); err != nil {
		return "", fmt.Errorf("failed to make backup executable: %w", err)
	}

	return backupPath, nil
}

// ListBackups lists available backups
func (rm *RollbackManager) ListBackups() ([]string, error) {
	entries, err := os.ReadDir(rm.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "ksau-go-backup-") {
			backups = append(backups, entry.Name())
		}
	}

	return backups, nil
}

// RestoreBackup restores a backup
func (rm *RollbackManager) RestoreBackup(backupName, currentExe string) error {
	backupPath := filepath.Join(rm.backupDir, backupName)
	
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup not found: %s", backupName)
	}

	// Create backup of current binary before restoring
	tempBackup := currentExe + ".temp-backup"
	if err := os.Rename(currentExe, tempBackup); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	// Restore backup
	if err := os.Rename(backupPath, currentExe); err != nil {
		// Try to restore the temp backup
		os.Rename(tempBackup, currentExe)
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	// Clean up temp backup
	os.Remove(tempBackup)
	
	return nil
}

// CleanupOldBackups removes backups older than 30 days
func (rm *RollbackManager) CleanupOldBackups() error {
	entries, err := os.ReadDir(rm.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	cutoffTime := time.Now().AddDate(0, 0, -30) // 30 days ago
	
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "ksau-go-backup-") {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			
			if info.ModTime().Before(cutoffTime) {
				backupPath := filepath.Join(rm.backupDir, entry.Name())
				if err := os.Remove(backupPath); err != nil {
					fmt.Printf("Warning: failed to remove old backup %s: %v\n", entry.Name(), err)
				}
			}
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = destination.ReadFrom(source)
	return err
}
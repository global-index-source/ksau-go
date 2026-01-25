package cmd

import (
	"fmt"
	"os"

	"github.com/global-index-source/ksau-go/pkg/update"
	"github.com/spf13/cobra"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback to a previous version",
	Long: `Rollback ksau-go to a previously backed up version.
This is useful if an update causes issues.`,
	RunE: runRollback,
}

var (
	listBackups bool
	backupName  string
)

func init() {
	rootCmd.AddCommand(rollbackCmd)
	
	rollbackCmd.Flags().BoolVar(&listBackups, "list", false, "List available backups")
	rollbackCmd.Flags().StringVar(&backupName, "to", "", "Specific backup to restore to")
}

func runRollback(cmd *cobra.Command, args []string) error {
	rollbackManager := NewRollbackManager()
	
	if listBackups {
		backups, err := rollbackManager.ListBackups()
		if err != nil {
			return fmt.Errorf("failed to list backups: %w", err)
		}
		
		if len(backups) == 0 {
			fmt.Println("No backups available.")
			return nil
		}
		
		fmt.Println("Available backups:")
		for _, backup := range backups {
			fmt.Printf("  %s\n", backup)
		}
		return nil
	}

	// Get current executable path
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get current executable path: %w", err)
	}

	// If no specific backup is provided, use the most recent one
	if backupName == "" {
		backups, err := rollbackManager.ListBackups()
		if err != nil {
			return fmt.Errorf("failed to list backups: %w", err)
		}
		
		if len(backups) == 0 {
			return fmt.Errorf("no backups available for rollback")
		}
		
		backupName = backups[len(backups)-1] // Get the most recent backup
	}

	fmt.Printf("Rolling back to %s...\n", backupName)
	
	if err := rollbackManager.RestoreBackup(backupName, currentExe); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	fmt.Println("Rollback completed successfully!")
	fmt.Printf("Restored to backup: %s\n", backupName)
	
	// Restart with the rolled back binary
	config := &update.Config{
		RepoOwner:      "TronPkr",
		RepoName:       "ksau-go",
		CurrentVersion: Version,
	}
	updater := update.NewUpdater(config)
	return updater.RestartWithNewBinary(currentExe)
}
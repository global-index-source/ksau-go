package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var updateConfigCmd = &cobra.Command{
	Use:   "update-config",
	Short: "Configure automatic update settings",
	Long: `Configure how ksau-go handles automatic updates.
You can enable/disable auto-update, set the update channel, and configure prerelease preferences.`,
	RunE: runUpdateConfig,
}

var (
	setAutoUpdate    bool
	setUpdateChannel string
	setPrerelease    bool
	showConfig      bool
)

func init() {
	rootCmd.AddCommand(updateConfigCmd)
	
	updateConfigCmd.Flags().BoolVar(&setAutoUpdate, "auto-update", false, "Enable or disable automatic updates")
	updateConfigCmd.Flags().StringVar(&setUpdateChannel, "channel", "", "Set update channel (stable, beta, alpha)")
	updateConfigCmd.Flags().BoolVar(&setPrerelease, "prerelease", false, "Include pre-release versions")
	updateConfigCmd.Flags().BoolVar(&showConfig, "show", false, "Show current update configuration")
}

func runUpdateConfig(cmd *cobra.Command, args []string) error {
	configManager := NewConfigManager()
	
	// Show current configuration
	config, err := configManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if showConfig || (!cmd.Flags().Changed("auto-update") && !cmd.Flags().Changed("channel") && !cmd.Flags().Changed("prerelease")) {
		fmt.Printf("Update Configuration:\n")
		fmt.Printf("  Auto Update: %t\n", config.AutoUpdate)
		fmt.Printf("  Update Channel: %s\n", config.UpdateChannel)
		fmt.Printf("  Include Prereleases: %t\n", config.Prerelease)
		fmt.Printf("  Last Check: %s\n", config.LastCheck.Format(time.RFC3339))
		return nil
	}

	// Update configuration
	if cmd.Flags().Changed("auto-update") {
		config.AutoUpdate = setAutoUpdate
		fmt.Printf("Set auto-update to %t\n", setAutoUpdate)
	}

	if cmd.Flags().Changed("channel") {
		validChannels := map[string]bool{
			"stable": true,
			"beta":   true,
			"alpha":  true,
		}
		if !validChannels[setUpdateChannel] {
			return fmt.Errorf("invalid channel '%s'. Valid channels: stable, beta, alpha", setUpdateChannel)
		}
		config.UpdateChannel = setUpdateChannel
		fmt.Printf("Set update channel to %s\n", setUpdateChannel)
	}

	if cmd.Flags().Changed("prerelease") {
		config.Prerelease = setPrerelease
		fmt.Printf("Set prerelease inclusion to %t\n", setPrerelease)
	}

	// Save updated configuration
	if err := configManager.SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("Configuration updated successfully.")
	return nil
}
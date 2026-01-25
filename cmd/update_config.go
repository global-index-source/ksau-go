package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// UpdateConfigFile represents the configuration file for updates
type UpdateConfigFile struct {
	LastCheck     time.Time `json:"last_check"`
	AutoUpdate    bool      `json:"auto_update"`
	UpdateChannel string    `json:"update_channel"`
	Prerelease    bool      `json:"prerelease"`
}

// ConfigManager handles update configuration
type ConfigManager struct {
	configPath string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".ksau")
	return &ConfigManager{
		configPath: filepath.Join(configDir, "update-config.json"),
	}
}

// LoadConfig loads the update configuration
func (cm *ConfigManager) LoadConfig() (*UpdateConfigFile, error) {
	config := &UpdateConfigFile{
		LastCheck:     time.Time{},
		AutoUpdate:    false,
		UpdateChannel: "stable",
		Prerelease:    false,
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(cm.configPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load existing config if it exists
	if _, err := os.Stat(cm.configPath); err == nil {
		data, err := os.ReadFile(cm.configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := json.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	return config, nil
}

// SaveConfig saves the update configuration
func (cm *ConfigManager) SaveConfig(config *UpdateConfigFile) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cm.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ShouldCheck determines if we should check for updates
func (cm *ConfigManager) ShouldCheck() (bool, error) {
	config, err := cm.LoadConfig()
	if err != nil {
		return false, err
	}

	// Check if we haven't checked in the last 24 hours
	if time.Since(config.LastCheck) < 24*time.Hour {
		return false, nil
	}

	return true, nil
}

// UpdateLastCheck updates the last check timestamp
func (cm *ConfigManager) UpdateLastCheck() error {
	config, err := cm.LoadConfig()
	if err != nil {
		return err
	}

	config.LastCheck = time.Now()
	return cm.SaveConfig(config)
}
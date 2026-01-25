package update

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ConfigFile represents the configuration file for updates
type ConfigFile struct {
	LastCheck     time.Time `json:"last_check"`
	AutoUpdate    bool      `json:"auto_update"`
	UpdateChannel string    `json:"update_channel"`
	Prerelease    bool      `json:"prerelease"`
}

// Manager handles update configuration
type Manager struct {
	configPath string
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".ksau")
	return &Manager{
		configPath: filepath.Join(configDir, "update-config.json"),
	}
}

// LoadConfig loads the update configuration
func (m *Manager) LoadConfig() (*ConfigFile, error) {
	config := &ConfigFile{
		LastCheck:     time.Time{},
		AutoUpdate:    false,
		UpdateChannel: "stable",
		Prerelease:    false,
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(m.configPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load existing config if it exists
	if _, err := os.Stat(m.configPath); err == nil {
		data, err := os.ReadFile(m.configPath)
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
func (m *Manager) SaveConfig(config *ConfigFile) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ShouldCheck determines if we should check for updates
func (m *Manager) ShouldCheck() (bool, error) {
	config, err := m.LoadConfig()
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
func (m *Manager) UpdateLastCheck() error {
	config, err := m.LoadConfig()
	if err != nil {
		return err
	}

	config.LastCheck = time.Now()
	return m.SaveConfig(config)
}

// SetAutoUpdate enables or disables auto-update
func (m *Manager) SetAutoUpdate(enabled bool) error {
	config, err := m.LoadConfig()
	if err != nil {
		return err
	}

	config.AutoUpdate = enabled
	return m.SaveConfig(config)
}

// SetPrerelease enables or disables prerelease updates
func (m *Manager) SetPrerelease(enabled bool) error {
	config, err := m.LoadConfig()
	if err != nil {
		return err
	}

	config.Prerelease = enabled
	return m.SaveConfig(config)
}
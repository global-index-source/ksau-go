package azure

import (
	"fmt"
	"slices"
	"strings"
)

// ParseRcloneConfigData parses rclone configuration data from a byte slice and returns an array of configuration maps.
// Each configuration map represents a remote section in the rclone config, containing key-value pairs of settings.
//
// The function processes the config data line by line, handling:
// - Section headers in [section-name] format
// - Key-value pairs in "key = value" format
// - Empty values in "key =" format
// - Blank lines and comments (lines starting with #) are ignored
//
// Each remote section is converted into a map with its settings, including a special "remote_name" key
// containing the section name.
//
// Parameters:
//   - configData: []byte containing the rclone configuration data
//
// Returns:
//   - []map[string]string: Array of maps, each containing config settings for one remote
//   - error: Returns error if parsing fails
func ParseRcloneConfigData(configData []byte) ([]map[string]string, error) {
	// fmt.Println("Parsing rclone config data for remote:", remoteConfig)
	content := string(configData)
	lines := strings.Split(content, "\n")
	var configMaps []map[string]string
	configMap := make(map[string]string)

	var currentSection string
	for linenum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if len(configMap) > 0 {
				configMaps = append(configMaps, configMap)
				// clear out the configMap
				configMap = make(map[string]string)
			}
			currentSection = strings.Trim(line, "[]")
			configMap["remote_name"] = currentSection
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			configMap[key] = value
		} else if len(parts) == 1 { // allow config to set empty value e.g. "foo ="
			key := strings.TrimSpace(parts[0])
			configMap[key] = ""
		} else {
			return nil, fmt.Errorf("error parsing line %d of rclone config", linenum)
		}
	}

	configMaps = append(configMaps, configMap)
	return configMaps, nil
}

// GetAvailableRemotes extracts and returns a slice of remote names from the parsed rclone configuration.
// It takes a pointer to a slice of string maps representing the parsed rclone config and iterates through
// each map's keys to collect all remote names.
//
// Parameters:
//   - parsedRcloneConfig: A pointer to a slice of maps containing the parsed rclone configuration
//
// Returns:
//   - []string: A slice containing all available remote names from the configuration
func GetAvailableRemotes(parsedRcloneConfig *[]map[string]string) []string {
	var remotes []string
	for _, elem := range *parsedRcloneConfig {
		remotes = append(remotes, elem["remote_name"])
	}

	return remotes
}

// GetRemoteConfig retrieves the configuration map for a specified remote from parsed rclone config.
// It takes a pointer to a slice of string maps containing parsed rclone configurations and a remote name as input.
// Returns the configuration map for the specified remote if found, or an error if the remote doesn't exist.
//
// Parameters:
//   - parsedRcloneConfig: Pointer to slice of maps containing parsed rclone configurations
//   - remoteConfig: Name of the remote configuration to retrieve
//
// Returns:
//   - map[string]string: Configuration map for the specified remote
//   - error: Error if remote is not found or any other error occurs
func GetRemoteConfig(parsedRcloneConfig *[]map[string]string, remoteConfig string) (map[string]string, error) {
	availableRemotes := GetAvailableRemotes(parsedRcloneConfig)

	if !slices.Contains(availableRemotes, remoteConfig) {
		return nil, fmt.Errorf("remote %s does not exist", remoteConfig)
	}

	for _, elem := range *parsedRcloneConfig {
		for key := range elem {
			if key == remoteConfig {
				return elem, nil
			}
		}
	}

	return nil, fmt.Errorf("this shouldn't be reachable(?)")
}

// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

// Name of the config file that will be used to store user's configuration
const CONFIG_FILE_NAME string = ".ksau.json"

// File access permissions based on the read-only flag
// Keep it all in one place for easier maintenance
var fileAccessPermissions = map[bool]int{
	true:  os.O_RDONLY,
	false: os.O_RDWR | os.O_CREATE | os.O_TRUNC,
}

func GetUserConfigFile(readOnly bool) (*os.File, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot find your home dir")
	}

	userConfigFilePath := filepath.Join(userHome, CONFIG_FILE_NAME)

	userConfigFile, err := os.OpenFile(userConfigFilePath, fileAccessPermissions[readOnly], 0644)
	if err != nil {
		return nil, fmt.Errorf("error while trying to open your config file: %w", err)
	}

	return userConfigFile, nil
}

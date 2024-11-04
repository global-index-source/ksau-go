// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"fmt"
	"io"
	"net/http"
)

func DownloadConfig(url string) error {
	// Download the config from the given URL
	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download config: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	configData, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	configFile, err := GetUserConfigFile(false)
	if err != nil {
		return fmt.Errorf("failed to get config file: %w", err)
	}

	// Write the downloaded config data to the config file
	if _, err := configFile.Write(configData); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

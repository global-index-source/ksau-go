// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"os"
	"path/filepath"
)

func GetExecutableName() string {
	executable, err := os.Executable()

	if err != nil {
		return "ksau" // Fallback name if there is an error
	}

	return filepath.Base(executable) // Get just the executable name
}

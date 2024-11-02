package internal

import (
	"fmt"
	"os"
)

func GetUserConfigFile(readOnly bool) (*os.File, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot find your home dir")
	}

	var userConfigFile *os.File
	if readOnly {
		userConfigFile, err = os.OpenFile(userHome+"/"+ConfigFileName, os.O_RDONLY, 0644)
	} else {
		userConfigFile, err = os.OpenFile(userHome+"/"+ConfigFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	}
	if err != nil {
		return nil, fmt.Errorf("error while trying to open your config file: %s", err.Error())
	}

	return userConfigFile, nil
}

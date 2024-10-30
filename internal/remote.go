package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

const ConfigFileName string = ".ksau.json"

type Remote struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	DriveId      string `json:"drive_id"`
	DriveType    string `json:"drive_type"`

	// Token stuff
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Expiry       int    `json:"expiry"`
}

// The (encrypted) json config file that we'll be shipping.
// It will contain all three remotes, and their respective id, token, etc.
type Remotes map[string]Remote

func check(err error, msg string) {
	if err != nil {
		log.Panic(msg)
	}
}

// Not to be exported
func getUserConfigFile() (*os.File, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot find your home dir")
	}

	userConfigFile, err := os.OpenFile(userHome+"/"+ConfigFileName, os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("error while trying to open your config file: %s", err.Error())
	}

	return userConfigFile, nil
}

// Returns a pointer to a an array of Remote (Remotes type alias), with config file already parsed for you
// TODO(hakimi): Do not just panic but instead return relevant error, so that CLI part can let the user know what went wrong
func GetRemotes() (*Remotes, error) {
	userConfigFile, err := getUserConfigFile()
	check(err, err.Error())
	defer userConfigFile.Close()

	userConfigFileContent, err := io.ReadAll(userConfigFile)
	check(err, err.Error())

	var remotes *Remotes
	err = json.Unmarshal(userConfigFileContent, remotes)
	check(err, err.Error())

	return remotes, nil
}

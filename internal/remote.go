package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"slices"
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

// Returns a pointer to a an array of Remote (Remotes type alias), with config file already parsed for you
// TODO(hakimi): Do not just panic but instead return relevant error, so that CLI part can let the user know what went wrong
func GetRemotes() (*Remotes, error) {
	userConfigFile, err := GetUserConfigFile()
	check(err, err.Error())
	defer userConfigFile.Close()

	userConfigFileContent, err := io.ReadAll(userConfigFile)
	check(err, err.Error())

	var remotes *Remotes
	err = json.Unmarshal(userConfigFileContent, remotes)
	check(err, err.Error())

	return remotes, nil
}

// Return a pointer to a remote, provided that the remote name given exists.
func GetRemote(name string) (*Remote, error) {
	remotes, err := GetRemotes()
	check(err, err.Error())

	var remoteNames []string = make([]string, len(*remotes))
	if !slices.Contains(remoteNames, name) {
		return nil, fmt.Errorf("Remote '%s' remote does not exist", name)
	}

	var remote Remote = (*remotes)[name]
	return &remote, nil
}

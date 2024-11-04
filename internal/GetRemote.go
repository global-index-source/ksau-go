// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"maps"
	"slices"

	"github.com/global-index-source/ksau-go/crypto"
)

type Remote struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	DriveId      string `json:"drive_id"`
	DriveType    string `json:"drive_type"`

	// Also present in original ksau, just wasn't obvious at first sight.
	// this allows file to be uploaded not to root of the drive,
	// but rather to this particular folder.
	Prefix string `json:"prefix"`

	// Token stuff
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	Expiry       string `json:"expiry"`
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
func GetRemotes() (*Remotes, error) {
	userConfigFile, err := GetUserConfigFile(true)
	if err != nil {
		return nil, fmt.Errorf("cannot get user's config file: %w", err)
	}
	defer userConfigFile.Close()

	encryptedUserConfigFileContent, err := io.ReadAll(userConfigFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read user's config file: %w", err)
	}

	var decrypted string = crypto.Decrypt(encryptedUserConfigFileContent)

	var remotes *Remotes = &Remotes{}
	err = json.Unmarshal([]byte(decrypted), remotes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return remotes, nil
}

// Return a pointer to a remote, provided that the remote name given exists.
func GetRemote(name string) (*Remote, error) {
	remotes, err := GetRemotes()
	check(err, "could not get remotes")

	var remoteNames []string = make([]string, len(*remotes))
	for key := range maps.Keys(*remotes) {
		remoteNames = append(remoteNames, key)
	}

	if !slices.Contains(remoteNames, name) {
		return nil, fmt.Errorf("Remote '%s' remote does not exist", name)
	}

	var remote Remote = (*remotes)[name]
	return &remote, nil
}

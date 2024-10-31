package cmd

import (
	"fmt"
	"io"
	"net/http"

	"github.com/global-index-source/ksau-go/internal"
	"github.com/spf13/cobra"
)

// This always points to the latest file
const gistUrl string = "https://gist.githubusercontent.com/hakimifr/d964b07dff17ff7dc80d39c93ecdfdeb/raw/ksau.json.asc"

// hell
// Download a given url and return a File pointer
// not meant for large files
func download(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseInByteSlice, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return responseInByteSlice, nil
}

func downloadAndWriteConfigFromGist() error {
	configData, err := download(gistUrl)
	if err != nil {
		return err
	}

	configFile, err := internal.GetUserConfigFile()
	if err != nil {
		return err
	}

	_, err = configFile.Write(configData)
	if err != nil {
		return err
	}

	return nil
}

func refreshCmdCallback(cmd *cobra.Command, args []string) {
	fmt.Println("Updating tokens...")

	var err error = downloadAndWriteConfigFromGist()
	if err != nil {
		panic("Error while updating tokens: " + err.Error())
	}

}

var refreshCmd *cobra.Command = &cobra.Command{
	Use:   "refresh",
	Short: "Download latest token for each remote",
	Long:  `Download latest token for each remote. Try this if other commands fails to run.`,

	Run: refreshCmdCallback,
}

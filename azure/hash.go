package azure

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GetQuickXorHash retrieves the QuickXorHash value for a specified file from Microsoft Graph API.
//
// Parameters:
//   - httpClient: *http.Client - The HTTP client used to make the request
//   - fileID: string - The unique identifier of the file in Microsoft OneDrive
//
// Returns:
//   - string: The QuickXorHash value of the file
//   - error: An error object that indicates if the operation was unsuccessful
//
// The function performs the following steps:
// 1. Validates the access token
// 2. Makes a GET request to Microsoft Graph API to fetch file metadata
// 3. Parses the response to extract the QuickXorHash value
//
// Error cases:
//   - Invalid or expired access token
//   - Failed HTTP request
//   - Non-200 HTTP response
//   - Missing QuickXorHash in metadata
//   - JSON parsing errors
func (client *AzureClient) GetQuickXorHash(httpClient *http.Client, fileID string) (string, error) {
	// Ensure the access token is valid
	if err := client.EnsureTokenValid(httpClient); err != nil {
		return "", err
	}

	// Construct the URL to get the file's metadata
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/items/%s", fileID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+client.AccessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch file metadata: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to fetch file metadata, status: %d, response: %s", resp.StatusCode, responseBody)
	}

	// Parse the response to extract the quickXorHash
	var metadata struct {
		File struct {
			Hashes struct {
				QuickXorHash string `json:"quickXorHash"`
			} `json:"hashes"`
		} `json:"file"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return "", fmt.Errorf("failed to parse metadata: %v", err)
	}

	if metadata.File.Hashes.QuickXorHash == "" {
		return "", fmt.Errorf("quickXorHash not found in metadata")
	}

	return metadata.File.Hashes.QuickXorHash, nil
}

package azure

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// AzureClient represents a client for interacting with Microsoft Azure services.
// It manages authentication credentials and access tokens for Azure API operations.
//
// Fields:
//   - ClientID: The application (client) ID registered in Azure Active Directory
//   - ClientSecret: The client secret key for authentication
//   - AccessToken: The current OAuth access token for API requests
//   - RefreshToken: Token used to obtain a new access token when expired
//   - Expiration: Timestamp indicating when the current access token expires
//   - DriveID: The identifier for the specific OneDrive instance
//   - DriveType: The type of drive (personal, business, sharepoint)
//   - mu: Mutex for handling concurrent access to client fields
type AzureClient struct {
	ClientID     string
	ClientSecret string
	AccessToken  string
	RefreshToken string
	Expiration   time.Time
	DriveID      string
	DriveType    string

	// Root folder of the remote. Sometimes a remote may not want the tool from
	// uploading directly to the root folder, but instead into a custom folder.
	RemoteRootFolder string

	// Base url from which user can download the file.
	RemoteBaseUrl string

	mu sync.Mutex
}

// NewAzureClientFromRcloneConfigData creates a new AzureClient instance using rclone configuration data.
// It takes a byte slice containing rclone config data and a remote configuration name as input.
//
// The function parses the rclone configuration and extracts necessary Azure credentials including:
// - Client ID and Secret
// - Access and Refresh tokens
// - Token expiration time
// - Drive ID and Drive type
//
// Parameters:
//   - configData: []byte containing the rclone configuration data
//   - remoteConfig: string specifying which remote configuration to use
//
// Returns:
//   - *AzureClient: Pointer to initialized AzureClient instance
//   - error: Error if configuration parsing or client creation fails
func NewAzureClientFromRcloneConfigData(configData []byte, remoteConfig string) (*AzureClient, error) {
	// fmt.Println("Reading rclone config from embedded data for remote:", remoteConfig)
	configMaps, err := ParseRcloneConfigData(configData)
	var configMap map[string]string
	if err != nil {
		return nil, fmt.Errorf("failed to parse rclone config: %v", err)
	}

	for _, elem := range configMaps {
		if elem["remote_name"] == remoteConfig {
			configMap = elem
		}
	}

	var client AzureClient

	client.ClientID = configMap["client_id"]
	client.ClientSecret = configMap["client_secret"]
	client.RemoteRootFolder = configMap["root_folder"]
	client.RemoteBaseUrl = configMap["base_url"]

	// Extract token information
	var tokenData struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		Expiry       string `json:"expiry"`
	}
	err = json.Unmarshal([]byte(configMap["token"]), &tokenData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token JSON: %v", err)
	}

	client.AccessToken = tokenData.AccessToken
	client.RefreshToken = tokenData.RefreshToken

	expiration, err := time.Parse(time.RFC3339, tokenData.Expiry)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token expiration time: %v", err)
	}
	client.Expiration = expiration

	client.DriveID = configMap["drive_id"]
	client.DriveType = configMap["drive_type"]

	return &client, nil
}

// EnsureTokenValid ensures the Azure access token is valid by checking its expiration
// and refreshing it if necessary. It uses a mutex to ensure thread-safe token updates.
//
// The function performs the following steps:
// 1. Checks if the current token is still valid
// 2. If expired, requests a new token using the refresh token
// 3. Updates the client's access token, refresh token, and expiration time
//
// Parameters:
//   - httpClient: *http.Client - The HTTP client used to make the token refresh request
//
// Returns:
//   - error: Returns nil if token is valid or successfully refreshed, error otherwise
//
// Thread-safety: This method is thread-safe as it uses a mutex to protect token updates.
func (client *AzureClient) EnsureTokenValid(httpClient *http.Client) error {
	client.mu.Lock()
	defer client.mu.Unlock()

	if time.Now().Before(client.Expiration) {
		return nil
	}

	tokenURL := "https://login.microsoftonline.com/common/oauth2/v2.0/token"
	data := url.Values{}
	data.Set("client_id", client.ClientID)
	data.Set("client_secret", client.ClientSecret)
	data.Set("refresh_token", client.RefreshToken)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return fmt.Errorf("failed to refresh token, status code: %v", res.StatusCode)
	}

	var responseData struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	err = json.NewDecoder(res.Body).Decode(&responseData)
	if err != nil {
		return err
	}

	client.AccessToken = responseData.AccessToken
	client.RefreshToken = responseData.RefreshToken
	client.Expiration = time.Now().Add(time.Duration(responseData.ExpiresIn) * time.Second)

	return nil
}

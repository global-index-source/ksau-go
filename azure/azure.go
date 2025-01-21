// Package azure provides functionality for interacting with Microsoft Azure services,
// specifically focusing on OneDrive and SharePoint storage operations through Microsoft Graph API.
//
// The package implements various features including:
//   - Authentication and token management for Azure services
//   - Large file uploads with chunked transfer and parallel processing
//   - File metadata retrieval and management
//   - Storage quota information
//   - Hash verification
//   - Integration with rclone configuration
//
// Core Components:
//
// AzureClient: The main client struct that handles authentication and API operations.
// It manages access tokens, refresh tokens, and provides methods for file operations.
//
// UploadParams: Configuration struct for customizing file upload behavior including
// chunk size, parallel processing, and retry mechanisms.
//
// DriveQuota: Represents storage quota information including total, used, and remaining space.
//
// Key Features:
//   - Automatic token refresh and management
//   - Parallel chunk upload with configurable workers
//   - Retry mechanism for failed operations
//   - Progress tracking and error handling
//   - Storage quota management
//   - QuickXorHash verification
//
// Usage Example:
//
//	client, err := NewAzureClientFromRcloneConfigData(configData, "remote")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	params := UploadParams{
//	    FilePath: "local/path/file.txt",
//	    RemoteFilePath: "remote/path/file.txt",
//	    ChunkSize: 10 * 1024 * 1024, // 10MB chunks
//	    ParallelChunks: 4,
//	    MaxRetries: 3,
//	    RetryDelay: time.Second * 5,
//	}
//
//	fileID, err := client.Upload(httpClient, params)
//
// The package is designed to handle large file transfers efficiently and provides
// robust error handling and retry mechanisms for reliable file operations.
package azure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"slices"
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
	//fmt.Println("Reading rclone config from embedded data for remote:", remoteConfig)
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

// ParseRcloneConfigData parses rclone configuration data from a byte slice and returns an array of configuration maps.
// Each configuration map represents a remote section in the rclone config, containing key-value pairs of settings.
//
// The function processes the config data line by line, handling:
// - Section headers in [section-name] format
// - Key-value pairs in "key = value" format
// - Empty values in "key =" format
// - Blank lines and comments (lines starting with #) are ignored
//
// Each remote section is converted into a map with its settings, including a special "remote_name" key
// containing the section name.
//
// Parameters:
//   - configData: []byte containing the rclone configuration data
//
// Returns:
//   - []map[string]string: Array of maps, each containing config settings for one remote
//   - error: Returns error if parsing fails
func ParseRcloneConfigData(configData []byte) ([]map[string]string, error) {
	//fmt.Println("Parsing rclone config data for remote:", remoteConfig)
	content := string(configData)
	lines := strings.Split(content, "\n")
	var configMaps []map[string]string
	configMap := make(map[string]string)

	var currentSection string
	for linenum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if len(configMap) > 0 {
				configMaps = append(configMaps, configMap)
				// clear out the configMap
				configMap = make(map[string]string)
			}
			currentSection = strings.Trim(line, "[]")
			configMap["remote_name"] = currentSection
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			configMap[key] = value
		} else if len(parts) == 1 { // allow config to set empty value e.g. "foo ="
			key := strings.TrimSpace(parts[0])
			configMap[key] = ""
		} else {
			return nil, fmt.Errorf("error parsing line %d of rclone config", linenum)
		}
	}

	return configMaps, nil
}

// GetAvailableRemotes extracts and returns a slice of remote names from the parsed rclone configuration.
// It takes a pointer to a slice of string maps representing the parsed rclone config and iterates through
// each map's keys to collect all remote names.
//
// Parameters:
//   - parsedRcloneConfig: A pointer to a slice of maps containing the parsed rclone configuration
//
// Returns:
//   - []string: A slice containing all available remote names from the configuration
func GetAvailableRemotes(parsedRcloneConfig *[]map[string]string) []string {
	var remotes []string
	for _, elem := range *parsedRcloneConfig {
		for key := range elem {
			remotes = append(remotes, key)
		}
	}

	return remotes
}

// GetRemoteConfig retrieves the configuration map for a specified remote from parsed rclone config.
// It takes a pointer to a slice of string maps containing parsed rclone configurations and a remote name as input.
// Returns the configuration map for the specified remote if found, or an error if the remote doesn't exist.
//
// Parameters:
//   - parsedRcloneConfig: Pointer to slice of maps containing parsed rclone configurations
//   - remoteConfig: Name of the remote configuration to retrieve
//
// Returns:
//   - map[string]string: Configuration map for the specified remote
//   - error: Error if remote is not found or any other error occurs
func GetRemoteConfig(parsedRcloneConfig *[]map[string]string, remoteConfig string) (map[string]string, error) {
	availableRemotes := GetAvailableRemotes(parsedRcloneConfig)

	if !slices.Contains(availableRemotes, remoteConfig) {
		return nil, fmt.Errorf("remote %s does not exist", remoteConfig)
	}

	for _, elem := range *parsedRcloneConfig {
		for key := range elem {
			if key == remoteConfig {
				return elem, nil
			}
		}
	}

	return nil, fmt.Errorf("this shouldn't be reachable(?)")
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

// Upload performs a large file upload to Azure storage using chunked upload with parallel processing.
// It creates an upload session, splits the file into chunks, and uploads them in parallel using a worker pool.
//
// Parameters:
//   - httpClient: The HTTP client to use for requests
//   - params: UploadParams struct containing:
//   - FilePath: Local path of file to upload
//   - RemoteFilePath: Destination path in Azure storage
//   - ChunkSize: Size of each upload chunk in bytes
//   - ParallelChunks: Number of chunks to upload in parallel
//   - MaxRetries: Maximum number of retry attempts per chunk
//   - RetryDelay: Delay between retry attempts
//
// Returns:
//   - string: The file ID of the uploaded file
//   - error: Any error that occurred during upload
//
// The function implements the following features:
//   - Automatic token refresh
//   - Parallel chunk upload using worker pools
//   - Configurable chunk size and parallel upload count
//   - Retry mechanism for failed chunk uploads
//   - Progress tracking and error handling
func (client *AzureClient) Upload(httpClient *http.Client, params UploadParams) (string, error) {
	fmt.Println("Starting file upload with upload session...")

	// Ensure the access token is valid
	if err := client.EnsureTokenValid(httpClient); err != nil {
		return "", err
	}

	// Create an upload session
	uploadURL, err := client.createUploadSession(httpClient, params.RemoteFilePath, client.AccessToken)
	if err != nil {
		return "", fmt.Errorf("failed to create upload session: %v", err)
	}
	fmt.Println("Upload session created successfully.")

	// Open the file to upload
	file, err := os.Open(params.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Get file information
	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %v", err)
	}
	fileSize := fileInfo.Size()
	fmt.Printf("File size: %d bytes\n", fileSize)

	// Define chunk size and calculate the number of chunks
	chunkSize := params.ChunkSize
	numChunks := (fileSize + chunkSize - 1) / chunkSize

	// Create a worker pool for parallel uploads
	var wg sync.WaitGroup
	chunkChan := make(chan int64, numChunks)
	errChan := make(chan error, numChunks)

	// Start workers
	for i := 0; i < params.ParallelChunks; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for start := range chunkChan {
				end := start + chunkSize - 1
				if end >= fileSize {
					end = fileSize - 1
				}

				// Read the current chunk from the file
				chunk := make([]byte, end-start+1)
				_, err := file.ReadAt(chunk, start)
				if err != nil && err != io.EOF {
					errChan <- fmt.Errorf("failed to read chunk %d-%d: %v", start, end, err)
					continue
				}

				// Retry logic for chunk upload
				for retry := 0; retry < params.MaxRetries; retry++ {
					success, err := client.uploadChunk(httpClient, uploadURL, chunk, start, end, fileSize)
					if success {
						break
					}

					fmt.Printf("Error uploading chunk %d-%d: %v\n", start, end, err)
					fmt.Printf("Retrying chunk upload (attempt %d/%d)...\n", retry+1, params.MaxRetries)
					time.Sleep(params.RetryDelay)
				}
			}
		}()
	}

	// Send chunk start positions to the workers
	for start := int64(0); start < fileSize; start += chunkSize {
		chunkChan <- start
	}
	close(chunkChan)

	// Wait for all workers to finish
	wg.Wait()

	// Check for errors
	select {
	case err := <-errChan:
		return "", fmt.Errorf("failed to upload file: %v", err)
	default:
		fileID, err := client.getFileID(httpClient, params.RemoteFilePath)
		if err != nil {
			return "", fmt.Errorf("failed to fetch file ID: %v", err)
		}

		return fileID, nil
	}

}

// getFileID retrieves the unique identifier of a file from Microsoft OneDrive using the Microsoft Graph API.
// It takes an HTTP client and the remote path of the file as parameters.
//
// Parameters:
//   - httpClient: *http.Client - The HTTP client used to make the request
//   - remotePath: string - The path to the file in OneDrive
//
// Returns:
//   - string: The unique identifier of the file
//   - error: An error if the request fails, if the file is not found, or if the response cannot be parsed
//
// The function makes a GET request to the Microsoft Graph API, authenticating with the client's access token.
// It expects a JSON response containing the file's metadata, from which it extracts the ID.
// If the file is not found or any other error occurs during the process, it returns an appropriate error.
func (client *AzureClient) getFileID(httpClient *http.Client, remotePath string) (string, error) {
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/root:/%s", remotePath)
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

	var metadata struct {
		ID string `json:"id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return "", fmt.Errorf("failed to parse metadata: %v", err)
	}

	if metadata.ID == "" {
		return "", fmt.Errorf("file ID not found in metadata")
	}

	return metadata.ID, nil
}

// createUploadSession creates an upload session for a large file upload to OneDrive/SharePoint through Microsoft Graph API.
// It takes an HTTP client, the remote path where the file will be stored, and an access token for authentication.
//
// Parameters:
//   - httpClient: *http.Client - The HTTP client to make the request
//   - remotePath: string - The destination path in OneDrive where the file will be uploaded
//   - accessToken: string - OAuth2 access token for Microsoft Graph API authentication
//
// Returns:
//   - string: The upload URL to be used for subsequent chunk uploads
//   - error: An error object if the operation fails, nil otherwise
//
// The function implements Microsoft Graph API's large file upload protocol by creating
// an upload session with conflict behavior set to "rename" if a file with the same name exists.
// It returns an upload URL that can be used to upload the file in chunks.
func (client *AzureClient) createUploadSession(httpClient *http.Client, remotePath string, accessToken string) (string, error) {
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/root:/%s:/createUploadSession", remotePath)
	requestBody := map[string]interface{}{
		"item": map[string]string{
			"@microsoft.graph.conflictBehavior": "rename",
		},
	}
	body, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create upload session request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create upload session: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create upload session, status: %d, response: %s", resp.StatusCode, responseBody)
	}

	var response struct {
		UploadUrl string `json:"uploadUrl"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to parse upload session response: %v", err)
	}

	return response.UploadUrl, nil
}

// uploadChunk uploads a single chunk of data to Azure Blob Storage using the provided URL.
// It takes an HTTP client, the upload URL, the chunk data, start and end byte positions,
// and the total file size.
//
// Parameters:
//   - httpClient: The HTTP client to use for the request
//   - uploadURL: The URL to upload the chunk to
//   - chunk: The byte slice containing the chunk data
//   - start: The starting byte position of this chunk
//   - end: The ending byte position of this chunk
//   - totalSize: The total size of the complete file
//
// Returns:
//   - bool: true if upload was successful (status 201 Created or 202 Accepted)
//   - error: nil if successful, otherwise contains the error details with response body
//
// The function sets the Content-Range header according to Azure Blob Storage requirements
// and performs the upload using a PUT request.
func (client *AzureClient) uploadChunk(httpClient *http.Client, uploadURL string, chunk []byte, start, end, totalSize int64) (bool, error) {
	req, err := http.NewRequest("PUT", uploadURL, bytes.NewReader(chunk))
	if err != nil {
		return false, fmt.Errorf("failed to create chunk upload request: %v", err)
	}

	rangeHeader := fmt.Sprintf("bytes %d-%d/%d", start, end, totalSize)
	req.Header.Set("Content-Range", rangeHeader)

	resp, err := httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to upload chunk: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusAccepted {
		return true, nil
	}

	responseBody, _ := io.ReadAll(resp.Body)
	return false, fmt.Errorf("failed to upload chunk, status: %d, response: %s", resp.StatusCode, responseBody)
}

// itemByPath retrieves a DriveItem from Microsoft OneDrive by its file path.
// It makes a GET request to the Microsoft Graph API using the provided HTTP client and access token.
//
// Parameters:
//   - httpClient: An *http.Client to make the HTTP request
//   - accessToken: A valid Microsoft Graph API access token
//   - path: The file path in OneDrive to retrieve
//
// Returns:
//   - *DriveItem: The retrieved drive item if successful
//   - error: Any error encountered during the request or processing
//
// The function will return an error if:
//   - The HTTP request fails
//   - The response status code is not in the 2xx range
//   - The response body cannot be decoded into a DriveItem
func itemByPath(httpClient *http.Client, accessToken, path string) (*DriveItem, error) {
	fmt.Println("Retrieving item by path:", path)
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/root:/%s", path)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	fmt.Println("Item by path response status code:", res.StatusCode)

	if res.StatusCode < 200 || res.StatusCode > 299 {
		responseBody, _ := ioutil.ReadAll(res.Body)
		return nil, fmt.Errorf("failed to retrieve item, status code: %v, response: %s", res.StatusCode, string(responseBody))
	}

	var item DriveItem
	err = json.NewDecoder(res.Body).Decode(&item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

// DriveItem represents an item in a Microsoft OneDrive or SharePoint drive.
// It contains basic properties such as the unique identifier and name of the item.
type DriveItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// UploadParams contains configuration parameters for file upload operations to Azure Blob Storage.
//
// Fields:
//   - FilePath: Local path of the file to be uploaded
//   - RemoteFilePath: Destination path in Azure Blob Storage
//   - ChunkSize: Size of each upload chunk in bytes
//   - ParallelChunks: Number of chunks to upload concurrently
//   - MaxRetries: Maximum number of retry attempts for failed uploads
//   - RetryDelay: Duration to wait between retry attempts
//   - AccessToken: Azure authentication token for the upload operation
type UploadParams struct {
	FilePath       string
	RemoteFilePath string
	ChunkSize      int64
	ParallelChunks int
	MaxRetries     int
	RetryDelay     time.Duration
	AccessToken    string
}

// DriveQuota represents storage quota information for a drive.
// It contains details about the total storage space, used space,
// remaining space, and space used by items in the recycle bin.
type DriveQuota struct {
	Total     int64 `json:"total"`
	Used      int64 `json:"used"`
	Remaining int64 `json:"remaining"`
	Deleted   int64 `json:"deleted"`
}

// GetDriveQuota retrieves the quota information for the user's OneDrive storage using the Microsoft Graph API.
// It returns a DriveQuota struct containing total storage space, used space, remaining space and deleted space in bytes.
//
// The function automatically ensures the access token is valid before making the request.
// If any error occurs during the process (token validation, HTTP request, response parsing),
// it returns nil for DriveQuota and the corresponding error.
//
// Parameters:
//   - httpClient: *http.Client - The HTTP client to use for making the request
//
// Returns:
//   - *DriveQuota: Contains quota information (total, used, remaining, and deleted space)
//   - error: Any error encountered during the process
func (client *AzureClient) GetDriveQuota(httpClient *http.Client) (*DriveQuota, error) {
	// Ensure the access token is valid
	if err := client.EnsureTokenValid(httpClient); err != nil {
		return nil, err
	}

	// Construct the URL to get the drive's quota information
	url := fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/quota")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create quota request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+client.AccessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch quota information: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch quota information, status: %d, response: %s", resp.StatusCode, responseBody)
	}

	var quotaResponse struct {
		Total     int64 `json:"total"`
		Used      int64 `json:"used"`
		Remaining int64 `json:"remaining"`
		Deleted   int64 `json:"deleted"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&quotaResponse); err != nil {
		return nil, fmt.Errorf("failed to parse quota response: %v", err)
	}

	return &DriveQuota{
		Total:     quotaResponse.Total,
		Used:      quotaResponse.Used,
		Remaining: quotaResponse.Remaining,
		Deleted:   quotaResponse.Deleted,
	}, nil
}

// formatBytes converts a size in bytes to a human-readable string representation.
// It automatically chooses the appropriate unit (B, KiB, MiB, GiB, TiB, PiB, or EiB)
// and formats the number with three decimal places.
//
// Parameters:
//   - bytes: The size in bytes to be formatted
//
// Returns:
//   - A string representing the size with the appropriate binary unit suffix
//     For example:
//   - 1024 bytes -> "1.000 KiB"
//   - 1048576 bytes -> "1.000 MiB"
//   - 2000000000 bytes -> "1.863 GiB"
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.3f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// DisplayQuotaInfo prints quota information for a given remote drive to standard output.
// It displays the remote name and formatted storage values for total, used, free and trashed space.
//
// Parameters:
//   - remote: string representing the remote drive name/path
//   - quota: pointer to DriveQuota struct containing storage quota information
//
// The output is formatted as follows:
//   - Remote: <remote name>
//   - Total: <formatted total space>
//   - Used: <formatted used space>
//   - Free: <formatted remaining space>
//   - Trashed: <formatted deleted space>
func DisplayQuotaInfo(remote string, quota *DriveQuota) {
	fmt.Printf("Remote: %s\n", remote)
	fmt.Printf("Total:   %s\n", formatBytes(quota.Total))
	fmt.Printf("Used:    %s\n", formatBytes(quota.Used))
	fmt.Printf("Free:    %s\n", formatBytes(quota.Remaining))
	fmt.Printf("Trashed: %s\n", formatBytes(quota.Deleted))
	fmt.Println()
}

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

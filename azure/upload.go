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
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Upload performs a large file upload to Azure storage using chunked upload with parallel processing.
// It creates an upload session, splits the file into chunks, and uploads them in parallel using a worker pool.
//
// Parameters:
//   - httpClient: The HTTP client to use for requests
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

	// Set up channels for upload management
	var wg sync.WaitGroup
	chunkChan := make(chan int64, numChunks)
	errChan := make(chan error, numChunks)

	// Track total uploaded bytes with thread-safety
	var totalUploaded int64
	var progressMu sync.Mutex

	// Use a single worker to avoid session conflicts
	wg.Add(1)
	go func() {
		defer wg.Done()
		for start := range chunkChan {
			end := start + chunkSize - 1
			if end >= fileSize {
				end = fileSize - 1
			}
			actualChunkSize := end - start + 1

			// Read the current chunk from the file
			chunk := make([]byte, actualChunkSize)
			_, err := file.ReadAt(chunk, start)
			if err != nil && err != io.EOF {
				errChan <- fmt.Errorf("failed to read chunk %d-%d: %v", start, end, err)
				continue
			}

			// Retry logic for chunk upload with session refresh
			for retry := 0; retry < params.MaxRetries; retry++ {
				uploadSuccess, err := client.uploadChunk(httpClient, uploadURL, chunk, start, end, fileSize)
				if uploadSuccess {
					// Update progress
					progressMu.Lock()
					totalUploaded += actualChunkSize
					if params.ProgressCallback != nil {
						params.ProgressCallback(totalUploaded)
					}
					progressMu.Unlock()
					break
				}

				if retry < params.MaxRetries-1 {
					if strings.Contains(err.Error(), "resourceModified") || strings.Contains(err.Error(), "invalidRange") {
						// Session expired or range error, create new session
						newUploadURL, sessionErr := client.createUploadSession(httpClient, params.RemoteFilePath, client.AccessToken)
						if sessionErr != nil {
							fmt.Printf("Failed to create new upload session: %v\n", sessionErr)
							continue
						}
						uploadURL = newUploadURL
						fmt.Println("Created new upload session after error")
					}

					fmt.Printf("Error uploading chunk %d-%d: %v\n", start, end, err)
					fmt.Printf("Retrying chunk upload (attempt %d/%d)...\n", retry+1, params.MaxRetries)
					time.Sleep(params.RetryDelay)
				} else {
					errChan <- fmt.Errorf("failed to upload chunk after %d retries: %v", params.MaxRetries, err)
				}
			}
		}
	}()

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
			"@microsoft.graph.conflictBehavior": "replace",
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
	// Validate chunk parameters
	if start < 0 || end < start || end >= totalSize {
		return false, fmt.Errorf("invalid chunk range: start=%d, end=%d, total=%d", start, end, totalSize)
	}

	expectedSize := end - start + 1
	if int64(len(chunk)) != expectedSize {
		return false, fmt.Errorf("chunk size mismatch: got %d bytes, expected %d bytes", len(chunk), expectedSize)
	}

	// Create request with validated chunk
	req, err := http.NewRequest("PUT", uploadURL, bytes.NewReader(chunk))
	if err != nil {
		return false, fmt.Errorf("failed to create chunk upload request: %v", err)
	}

	// Set required headers for chunk upload
	rangeHeader := fmt.Sprintf("bytes %d-%d/%d", start, end, totalSize)
	req.Header.Set("Content-Range", rangeHeader)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", expectedSize))
	req.Header.Set("Content-Type", "application/octet-stream")

	// Perform upload
	resp, err := httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to upload chunk: %v", err)
	}
	defer resp.Body.Close()

	// Handle response based on status code
	switch resp.StatusCode {
	case http.StatusCreated, http.StatusAccepted, http.StatusOK:
		return true, nil
	case http.StatusRequestedRangeNotSatisfiable:
		responseBody, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("invalidRange: status %d, response: %s", resp.StatusCode, responseBody)
	case http.StatusConflict:
		responseBody, _ := io.ReadAll(resp.Body)
		if strings.Contains(string(responseBody), "resourceModified") {
			return false, fmt.Errorf("resourceModified: session expired")
		}
		return false, fmt.Errorf("conflict error: status %d, response: %s", resp.StatusCode, responseBody)
	default:
		responseBody, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("upload failed: status %d, response: %s", resp.StatusCode, responseBody)
	}
}

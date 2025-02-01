package azure

import "time"

// DriveItem represents an item in a Microsoft OneDrive or SharePoint drive.
// It contains basic properties such as the unique identifier and name of the item.
type DriveItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ProgressCallback is a function that gets called with progress updates
type ProgressCallback func(uploadedBytes int64)

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
	FilePath         string
	RemoteFilePath   string
	ChunkSize        int64
	ParallelChunks   int
	MaxRetries       int
	RetryDelay       time.Duration
	AccessToken      string
	ProgressCallback ProgressCallback
}

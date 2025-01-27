package cmd

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/global-index-source/ksau-go/azure"
	"github.com/spf13/cobra"
)

var (
	filePath       string
	remoteFolder   string
	remoteFileName string
	chunkSize      int64
	parallelChunks int
	maxRetries     int
	retryDelay     time.Duration
	skipHash       bool
	hashRetries    int
	hashRetryDelay time.Duration
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a file to OneDrive",
	Long: `Upload a file to OneDrive with support for chunked uploads,
parallel processing, and integrity verification.`,
	Run: runUpload,
}

func init() {
	rootCmd.AddCommand(uploadCmd)

	uploadCmd.Flags().StringVarP(&filePath, "file", "f", "", "Path to the local file to upload (required)")
	uploadCmd.Flags().StringVarP(&remoteFolder, "remote", "r", "", "Remote folder on OneDrive to upload the file (required)")
	uploadCmd.Flags().StringVarP(&remoteFileName, "remote-name", "n", "", "Optional: Remote filename (defaults to local filename)")
	uploadCmd.Flags().Int64VarP(&chunkSize, "chunk-size", "s", 0, "Chunk size for uploads in bytes (0 for automatic selection)")
	uploadCmd.Flags().IntVarP(&parallelChunks, "parallel", "p", 1, "Number of parallel chunks to upload")
	uploadCmd.Flags().IntVar(&maxRetries, "retries", 3, "Maximum number of retries for uploading chunks")
	uploadCmd.Flags().DurationVar(&retryDelay, "retry-delay", 5*time.Second, "Delay between retries")
	uploadCmd.Flags().BoolVar(&skipHash, "skip-hash", false, "Skip QuickXorHash verification")
	uploadCmd.Flags().IntVar(&hashRetries, "hash-retries", 5, "Maximum number of retries for fetching QuickXorHash")
	uploadCmd.Flags().DurationVar(&hashRetryDelay, "hash-retry-delay", 10*time.Second, "Delay between QuickXorHash retries")

	uploadCmd.MarkFlagRequired("file")
	uploadCmd.MarkFlagRequired("remote")
}

func runUpload(cmd *cobra.Command, args []string) {
	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Println("Failed to get file info:", err)
		return
	}
	fileSize := fileInfo.Size()

	// Get the remote config from persistent flags
	remoteConfig, _ := cmd.Flags().GetString("remote-config")
	if remoteConfig == "" {
		remoteConfig, err = selectRemoteAutomatically(fileSize)
		if err != nil {
			fmt.Println("cannot automatically determine remote to be used:", err.Error())
			os.Exit(1)
		}
	}

	// Dynamically select chunk size if not specified
	if chunkSize == 0 {
		chunkSize = getChunkSize(fileSize)
		fmt.Printf("Selected chunk size: %d bytes (based on file size: %d bytes)\n", chunkSize, fileSize)
	} else {
		fmt.Printf("Using user-specified chunk size: %d bytes\n", chunkSize)
	}

	// Determine remote filename and path
	localFileName := filepath.Base(filePath)
	remoteFilePath := filepath.Join(remoteFolder, localFileName)
	if remoteFileName != "" {
		remoteFilePath = filepath.Join(remoteFolder, remoteFileName)
	}

	// Read the rclone config file
	configData, err := getConfigData()
	if err != nil {
		fmt.Println("Failed to read config file:", err)
		return
	}

	client, err := azure.NewAzureClientFromRcloneConfigData(configData, remoteConfig)
	if err != nil {
		fmt.Println("Failed to initialize client:", err)
		return
	}

	// Add root folder for the selected remote configuration
	// rootFolder := getRootFolder(remoteConfig)
	rootFolder := client.RemoteRootFolder
	fullRemotePath := filepath.Join(rootFolder, remoteFilePath)
	fmt.Printf("Full remote path: %s\n", fullRemotePath)

	// Prepare upload parameters
	params := azure.UploadParams{
		FilePath:       filePath,
		RemoteFilePath: fullRemotePath,
		ChunkSize:      chunkSize,
		ParallelChunks: parallelChunks,
		MaxRetries:     maxRetries,
		RetryDelay:     retryDelay,
		AccessToken:    client.AccessToken,
	}

	httpClient := &http.Client{Timeout: 10 * time.Second}
	fileID, err := client.Upload(httpClient, params)
	if err != nil {
		fmt.Println("Failed to upload file:", err)
		return
	}

	if fileID != "" {
		fmt.Println("File uploaded successfully.")

		// Generate download URL
		baseURL := client.RemoteBaseUrl
		urlPath := strings.ReplaceAll(filepath.Join(remoteFolder, localFileName), "\\", "/")
		if remoteFileName != "" {
			urlPath = filepath.Join(remoteFolder, remoteFileName)
		}

		urlPath = strings.ReplaceAll(urlPath, " ", "%20")
		downloadURL := fmt.Sprintf("%s/%s", baseURL, urlPath)
		fmt.Printf("%sDownload URL:%s %s%s%s\n", ColorGreen, ColorReset, ColorGreen, downloadURL, ColorReset)

		if !skipHash {
			verifyFileIntegrity(filePath, fileID, client, httpClient)
		}
	} else {
		fmt.Println("File upload failed.")
	}
}

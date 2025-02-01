package azure

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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

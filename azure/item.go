package azure

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

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

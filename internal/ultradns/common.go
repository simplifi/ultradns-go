// Package ultradns Contains common shared code for http requests.
package ultradns

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// GetError checks the response for HTTP errors. If the status code is >= 400, it attempts to parse the response body
// to create the proper error struct.
// If no error is detected, returns nil.
func GetError(response *http.Response) error {
	if response.StatusCode < 400 {
		return nil
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("API call returned HTTP Status Code %d. Unable to read body of response", response.StatusCode)
	}

	errorJSON := ErrorResponse{}
	if err := json.Unmarshal(bodyBytes, &errorJSON); err != nil {
		return fmt.Errorf("API call returned HTTP Status Code %d. JSON parsing failed for body '%s'", response.StatusCode, string(bodyBytes))
	}

	return errorJSON
}

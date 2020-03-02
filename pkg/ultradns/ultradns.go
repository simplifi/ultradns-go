// Package ultradns holds the code for abstracting the UltraDNS API for manipulating Sitebacker pools.
package ultradns

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/simplifi/ultradns-go/internal/ultradns"
)

// APIConnection defines a connection to the UltraDNS API.
type APIConnection struct {
	Client        *http.Client
	Authorization *ultradns.Authorization
	BaseURL       string
}

// APIOptions is an options struct for passing into NewAPIConnection()
type APIOptions struct {
	// API Username
	Username string

	// API Password
	Password string

	// RefreshToken can be set in lieu of a Username/Password. By default, APIConnection will attempt to authenticate
	// using the RefreshToken if available, and fall back to the Username/Password only if it is unavailable or rejected.
	RefreshToken string

	// BaseURL is the first part of the API URL (default "https://api.ultradns.com")
	BaseURL string

	// Timeout is the underlying HTTP client timeout. Default is 5 seconds.
	Timeout time.Duration
}

func (options *APIOptions) setDefaults() {
	if options.Timeout == 0 {
		options.Timeout = 5 * time.Second
	}
	if options.BaseURL == "" {
		options.BaseURL = "https://api.ultradns.com"
	}
}

// NewAPIConnection Creates an APIConnection using the passed in APIOptions struct.
func NewAPIConnection(options *APIOptions) *APIConnection {
	options.setDefaults()

	httpClient := &http.Client{
		Timeout: options.Timeout,
	}
	auth := ultradns.NewAuthorization(options.Username, options.Password)
	auth.BaseURL = options.BaseURL

	return &APIConnection{
		Client:        httpClient,
		Authorization: auth,
		BaseURL:       options.BaseURL,
	}
}

// Get executes a GET request at the given path using the APIConnection's client and credentials
// error will be non-nil when:
// * encountering an error authorizing
// * Failing to connect to the API server
// * When getting an HTTP status code of >= 400
// TODO: Accept parameters
func (apiConn *APIConnection) Get(path string) (*http.Response, error) {
	err := apiConn.Authorization.Authorize(apiConn.Client)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", apiConn.BaseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+apiConn.Authorization.AccessToken)

	resp, err := apiConn.Client.Do(req)

	if err != nil {
		return resp, err
	}

	if resp.StatusCode >= 400 {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("API returned HTTP status code %d. Could not parse response body", resp.StatusCode)
		}
		errResp := ultradns.ErrorResponse{}

		err = json.Unmarshal(bodyBytes, &errResp)
		if err != nil {
			return resp, fmt.Errorf("API returned HTTP status code %d. Could not parse response body", resp.StatusCode)
		}

		fmt.Println(resp.StatusCode)
		fmt.Println(string(bodyBytes))
		return resp, &errResp
	}

	return resp, nil
}

// Pool stuff
// GET https://api.ultradns.com/zones/{zoneName}/rrsets/{rrType}/{ownerName}

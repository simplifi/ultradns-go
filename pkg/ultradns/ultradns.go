// Package ultradns holds the code for interacting with the UltraDNS REST API.
package ultradns

import (
	"io"
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

// Get executes a GET request at the given url using the APIConnection's client and credentials
// The url can have parameters on it, e.g. "/foo?bar=baz"
// error will be non-nil when:
// * encountering an error authorizing
// * Failing to connect to the API server
// * When getting an HTTP status code of >= 400
func (apiConn *APIConnection) Get(url string) (resp *http.Response, err error) {
	if err = apiConn.Authorization.Authorize(apiConn.Client); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", apiConn.BaseURL+url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+apiConn.Authorization.AccessToken)

	resp, err = apiConn.Client.Do(req)
	if err == nil {
		err = ultradns.GetError(resp)
	}

	return resp, err
}

// Post executes a POST request at the given url using the APIConnection's client and credentials.
// This function is similar to http.Post(), but does not require a Content-Type as the type is always set to
// 'application/json'
//
// error will be non-nil when:
// * encountering an error authorizing
// * Failing to connect to the API server
// * When getting an HTTP status code of >= 400
func (apiConn *APIConnection) Post(url string, body io.Reader) (resp *http.Response, err error) {
	if err = apiConn.Authorization.Authorize(apiConn.Client); err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", apiConn.BaseURL+url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+apiConn.Authorization.AccessToken)
	req.Header.Add("Content-Type", "application/json")
	resp, err = apiConn.Client.Do(req)
	if err == nil {
		err = ultradns.GetError(resp)
	}
	return resp, err
}

// Put executes a PUT request at the given url using the APIConnection's client and credentials.
// This function imitates the http.Post API, but does not require a Content-Type as the type is always set to
// 'application/json'
//
// error will be non-nil when:
// * encountering an error authorizing
// * Failing to connect to the API server
// * When getting an HTTP status code of >= 400
func (apiConn *APIConnection) Put(url string, body io.Reader) (resp *http.Response, err error) {
	if err = apiConn.Authorization.Authorize(apiConn.Client); err != nil {
		return nil, err
	}
	req, err := http.NewRequest("PUT", apiConn.BaseURL+url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+apiConn.Authorization.AccessToken)
	req.Header.Add("Content-Type", "application/json")
	resp, err = apiConn.Client.Do(req)
	if err == nil {
		err = ultradns.GetError(resp)
	}
	return resp, err
}

// Patch executes a PATCH request at the given url using the APIConnection's client and credentials.
// This function imitates the http.Post API, but does not require a Content-Type as the type is always set to
// 'application/json'
//
// error will be non-nil when:
// * encountering an error authorizing
// * Failing to connect to the API server
// * When getting an HTTP status code of >= 400
func (apiConn *APIConnection) Patch(url string, body io.Reader) (resp *http.Response, err error) {
	if err = apiConn.Authorization.Authorize(apiConn.Client); err != nil {
		return nil, err
	}
	req, err := http.NewRequest("PATCH", apiConn.BaseURL+url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+apiConn.Authorization.AccessToken)
	req.Header.Add("Content-Type", "application/json")
	resp, err = apiConn.Client.Do(req)
	if err == nil {
		err = ultradns.GetError(resp)
	}
	return resp, err
}

// JSONPatch executes a PATCH request at the given url using the APIConnection's client and credentials.
// JSON Patch is a special form of PATCH request that UltraDNS provides to allow partialy altering
// a record set. See the UltraDNS REST API for more detail.
// This function imitates the http.Post API, but does not require a Content-Type as the type is always set to
// 'application/json'
//
// error will be non-nil when:
// * encountering an error authorizing
// * Failing to connect to the API server
// * When getting an HTTP status code of >= 400
func (apiConn *APIConnection) JSONPatch(url string, body io.Reader) (resp *http.Response, err error) {
	if err = apiConn.Authorization.Authorize(apiConn.Client); err != nil {
		return nil, err
	}
	req, err := http.NewRequest("PATCH", apiConn.BaseURL+url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+apiConn.Authorization.AccessToken)
	req.Header.Add("Content-Type", "application/json-patch+json")
	resp, err = apiConn.Client.Do(req)
	if err == nil {
		err = ultradns.GetError(resp)
	}
	return resp, err
}

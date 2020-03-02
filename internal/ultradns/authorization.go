package ultradns

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

// Authorization encapsulates the data needed to use UltraDNS' authorization endpoints.
type Authorization struct {
	sync.Mutex
	Username     string
	Password     string
	AccessToken  string
	TokenExpires int64
	RefreshToken string
	BaseURL      string
}

// NewAuthorization returns an initialized Authorization struct
func NewAuthorization(username string, password string) *Authorization {
	auth := &Authorization{
		Username:     username,
		Password:     password,
		AccessToken:  "",
		TokenExpires: 0,
		RefreshToken: "",
		BaseURL:      "https://api.ultradns.com",
	}
	return auth
}

// Authorize retrives new tokens if necessary.
// Threadsafe.
// Checks the expiration of the current AccessToken. Does nothing if the AccessToken is not close to expiration.
// If the token has expired, then it will ask for a new token using the RefreshToken.
// If the RefreshToken is not available, then authorizes using the username/password.
// In any cases where authorization is performed, the struct is locked and updated with the new Tokens.
func (auth *Authorization) Authorize(client *http.Client) error {
	if auth.tokenIsValid() {
		return nil
	}

	resp, err := client.PostForm(auth.BaseURL+"/authorization/token", auth.authQuery())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 || resp.StatusCode < 200 {
		errorBodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("Authorization call returned HTTP Status Code %d. Unable to read body of response", resp.StatusCode)
		}

		errorJSON := ErrorResponse{}
		err = json.Unmarshal(errorBodyBytes, &errorJSON)
		if err != nil {
			return fmt.Errorf("Authorization call returned HTTP Status Code %d. JSON parsing failed for body '%s'", resp.StatusCode, string(errorBodyBytes))
		}

		return fmt.Errorf("Authorization call returned HTTP Status Code %d. Error was %s", resp.StatusCode, errorJSON.ErrorDescription())
	}

	currentEpoch := time.Now().Unix()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Pull the token data from the response and store it in the struct.
	authJSON := tokenResponse{}
	json.Unmarshal(bodyBytes, &authJSON)
	validSeconds, err := strconv.ParseInt(authJSON.ExpiresIn, 10, 64)
	if err != nil {
		return err
	}

	// Get the http.Client timeout in seconds to use as a margin of error.
	clientTimeout := int64(math.Ceil(client.Timeout.Seconds()))
	expiration := validSeconds + currentEpoch - clientTimeout

	// Writes to the struct should be controlled by locking.
	auth.Lock()

	auth.AccessToken = authJSON.AccessToken
	auth.RefreshToken = authJSON.RefreshToken
	auth.TokenExpires = expiration

	auth.Unlock()

	return nil
}

// authQuery returns the correct url.Values struct based on whether the RefreshToken is available
func (auth *Authorization) authQuery() url.Values {
	if auth.RefreshToken == "" {
		return url.Values{
			"grant_type": {"password"},
			"username":   {auth.Username},
			"password":   {auth.Password},
		}
	}

	return url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {auth.RefreshToken},
	}
}

// tokenIsValid Returns true if the Authorization struct has an unexpired AccessToken
func (auth *Authorization) tokenIsValid() bool {
	if auth.AccessToken == "" {
		return false
	}
	return time.Now().Unix() < auth.TokenExpires
}

func (auth *Authorization) String() string {
	return fmt.Sprintf("Authorization{\n  Username: '%s'\n  Password: '********',\n  AccessToken: '%s',\n  RefreshToken: '%s',\n  TokenExpires: %d\n}\n", auth.Username, auth.AccessToken, auth.RefreshToken, auth.TokenExpires)
}

// tokenResponse encapulates the Authorization response from UltraDNS
type tokenResponse struct {
	RefreshToken string `json:"refreshToken"`
	AccessToken  string `json:"accessToken"`
	ExpiresIn    string `json:"expiresIn"`
}

// ErrorResponse is a representation of the UltraDNS' API JSON error messages.
// API calls can return this type as an error.
// UltraDNS' API return values vary between snake and camel case. This attempts to handle that.
// TODO: move to common code.
type ErrorResponse struct {
	ErrorResponseI
	// Numerical code
	ErrorCodeCC int `json:"errorCode"`
	ErrorCodeSC int `json:"error_code"`
	// human-readable error message
	ErrorMessageCC string `json:"errorMessage"`
	ErrorMessageSC string `json:"error_message"`
	// Specific error type, e.g. 'unsupported_grant_type'
	ErrorTypeValue string `json:"error"`
	// ErrorCode + ErrorMessage
	ErrorDescriptionCC string `json:"errorDescription"`
	ErrorDescriptionSC string `json:"error_description"`
}

// ErrorResponseI is the error reponse interface
type ErrorResponseI interface {
	ErrorCode() int
	ErrorMessage() string
	ErrorType() string
	ErrorDescription() string
}

// ErrorCode returns the error code from the error response
// The code is a numerical representation. `0` means no error.
func (e *ErrorResponse) ErrorCode() int {
	if e.ErrorCodeCC > 0 {
		return e.ErrorCodeCC
	}
	return e.ErrorCodeSC
}

// ErrorMessage returns the error message from the error response
// The error message is a human-readable message
func (e *ErrorResponse) ErrorMessage() string {
	if e.ErrorMessageCC != "" {
		return e.ErrorMessageCC
	}
	return e.ErrorMessageSC
}

// ErrorType returns a string error type. This is useful for being able to
// get a more concise string to do switching on for custom error handling.
// This just returns the ErrorTypeValue field to provide a consistent API for this error struct.
func (e *ErrorResponse) ErrorType() string {
	return e.ErrorTypeValue
}

// ErrorDescription returns the error description from the error response
// The error description is typically a combination of the error code and the error message.
func (e *ErrorResponse) ErrorDescription() string {
	if e.ErrorDescriptionCC != "" {
		return e.ErrorDescriptionCC
	}
	return e.ErrorDescriptionSC
}

// Error is the interface for the error type.
func (e *ErrorResponse) Error() string {
	switch {
	case e.ErrorDescription() != "":
		return e.ErrorDescription()
	case e.ErrorMessage() != "":
		return fmt.Sprintf("%d: %s", e.ErrorCode(), e.ErrorMessage())
	default:
		panic(e)
	}
}

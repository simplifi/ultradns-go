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
	return &Authorization{
		Username:     username,
		Password:     password,
		AccessToken:  "",
		TokenExpires: 0,
		RefreshToken: "",
		BaseURL:      "https://api.ultradns.com",
	}
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

	var bodyBytes []byte

	resp, err := client.PostForm(auth.BaseURL+"/authorization/token", auth.authQuery())

	if err != nil {
		return err
	}
	if err = GetError(resp); err != nil {
		return err
	}
	defer resp.Body.Close()

	if bodyBytes, err = ioutil.ReadAll(resp.Body); err != nil {
		return err
	}

	// Pull the token data from the response and store it in the struct.
	authJSON := tokenResponse{}
	json.Unmarshal(bodyBytes, &authJSON)
	// Pad the expiration with the client timeout to ensure we re-authorize well before the expiration
	clientTimeout := int64(math.Ceil(client.Timeout.Seconds()))
	return auth.updateTokens(&authJSON, clientTimeout)
}

// Update the tokens from the UltraDNS response. Locks the auth.
func (auth *Authorization) updateTokens(response *tokenResponse, padding int64) error {
	currentEpoch := time.Now().Unix()
	validSeconds, err := strconv.ParseInt(response.ExpiresIn, 10, 64)
	if err != nil {
		return err
	}

	expiration := validSeconds + currentEpoch - padding

	// Writes to the struct should be controlled by locking.
	auth.Lock()

	auth.AccessToken = response.AccessToken
	auth.RefreshToken = response.RefreshToken
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

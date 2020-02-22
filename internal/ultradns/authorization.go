package ultradns

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
// Checks the expiration of the current AccessToken. Does nothing if the AccessToken has not expired.
// If the token has expired, then it will ask for a new token using the RefreshToken.
// If the RefreshToken is not available, then authorizes using the username/password.
// In any cases where authorization is performed, the struct is locked and updated with the new Tokens.
func (auth *Authorization) Authorize(client *http.Client) error {
	if auth.TokenIsValid() {
		return nil
	}
	values := url.Values{
		// TODO: This should support refresh_token as well
		"grant_type": {"password"},
		"username":   {auth.Username},
		"password":   {auth.Password},
	}
	resp, err := client.PostForm(auth.BaseURL+"/authorization/token", values)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	currentEpoch := time.Now().Unix()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Pull the token data from the response and store it in the struct.
	authJSON := tokenResponse{}
	json.Unmarshal(bodyBytes, &authJSON)
	i, err := strconv.ParseInt(authJSON.ExpiresIn, 10, 64)
	if err != nil {
		return err
	}
	expiration := i + currentEpoch

	// Writes to the struct should be controlled by locking.
	auth.Lock()

	auth.AccessToken = authJSON.AccessToken
	auth.RefreshToken = authJSON.RefreshToken
	auth.TokenExpires = expiration

	auth.Unlock()

	return nil
}

// TokenIsValid Returns true if the Authorization struct has a current AccessToken
func (auth *Authorization) TokenIsValid() bool {
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
package ultradns

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const validUsername = "superuser"
const validPassword = "secretcode"

// Defines a mock server that handles the /authorization/token endpoint.
func ultradnsAuthMockServer(t *testing.T) *httptest.Server {
	// Container for the response
	var resp string
	var refreshToken string

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/authorization/token" {
			err := r.ParseForm()
			assert.NoError(t, err)

			switch r.Form.Get("grant_type") {
			case "refresh_token":
				// Using a refresh token for re-authentication. Check the refresh token for a match, and then generate a replacement one.
				if refreshToken != "" && r.Form.Get("refresh_token") == refreshToken {
					refreshToken = fmt.Sprintf("%0x", rand.Int63())
					resp = successResponse(refreshToken)
				} else {
					resp = failureResponse()
				}
			case "password":
				// Authentication via username/password. Check for a match, then generate a new refresh token to send back for future authentications.
				if r.Form.Get("username") == validUsername && r.Form.Get("password") == validPassword {
					refreshToken = fmt.Sprintf("%0x", rand.Int63())
					resp = successResponse(refreshToken)
				} else {
					resp = failureResponse()
				}
			default:
				resp = failureResponse()
			}
		}

		// Output the JSON to the client
		w.Write([]byte(resp))
	}))
}

func successResponse(refreshToken string) string {
	accessToken := fmt.Sprintf("%0x", rand.Int63())
	return `{"tokenType":"Bearer","refreshToken":"` + refreshToken + `","accessToken":"` + accessToken + `","expiresIn":"3600","username":"` + validUsername + `","refresh_token":"` + refreshToken + `","access_token":"` + accessToken + `","expires_in":"3600","token_type":"Bearer"}`
}

func failureResponse() string {
	// This is the general form of the errors; the specific failure reasons can vary, but it seems that the complexity in making that exactly
	// 1-to-1 with the UltraDNS API just obfuscates the code and provides no benefit.
	return `{"errorCode":60001,"errorMessage":"invalid_grant:Invalid username & password combination.","error":"invalid_grant","error_description":"60001: invalid_grant:Invalid username & password combination."}`
}

// Test that we do not re-authenticate until we are past the expiration time.
func TestAuthorizationPreservesTokensUntilExpiration(t *testing.T) {
	server := ultradnsAuthMockServer(t)
	defer server.Close()
	auth := NewAuthorization(validUsername, validPassword)

	// Inject the fake server's URL to capture requests.
	auth.BaseURL = server.URL

	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	// Initial authorization
	auth.Authorize(client)
	accessToken := auth.AccessToken
	// Authorize again, ensure that we aren't actually calling the API to get a new token.
	auth.Authorize(client)
	assert.Equal(t, accessToken, auth.AccessToken)
	// Set the expiration to the past so that calling the API is forced.
	auth.Lock()
	auth.TokenExpires = 0
	auth.Unlock()
	auth.Authorize(client)
	assert.NotEqual(t, accessToken, auth.AccessToken)
}

// Test that re-authorization uses the refresh token.
func TestRefreshTokenAuthorization(t *testing.T) {
	server := ultradnsAuthMockServer(t)
	defer server.Close()

	auth := NewAuthorization(validUsername, validPassword)
	auth.BaseURL = server.URL
	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	err := auth.Authorize(client)
	assert.NoError(t, err)
	refreshToken := auth.RefreshToken
	assert.NotEqual(t, refreshToken, "")

	// Clear out username/password to ensure that we can't use them.
	auth.Lock()
	auth.Username = ""
	auth.Password = ""
	auth.TokenExpires = 0
	auth.Unlock()

	// Reauthorize, ensure that the refresh token has changed
	err = auth.Authorize(client)
	assert.NoError(t, err)
	assert.NotEqual(t, refreshToken, "")
	assert.NotEqual(t, refreshToken, auth.RefreshToken)
}

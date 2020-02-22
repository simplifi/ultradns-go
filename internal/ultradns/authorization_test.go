package ultradns

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
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
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/authorization/token" {
			err := r.ParseForm()
			assert.NoError(t, err)

			// Generate a random fake accessToken and refreshToken
			refreshToken := fmt.Sprintf("%0x", rand.Int63())
			accessToken := fmt.Sprintf("%0x", rand.Int63())

			if r.Form.Get("grant_type") == "password" && r.Form.Get("username") == validUsername && r.Form.Get("password") == validPassword {
				resp = `{"tokenType":"Bearer","refreshToken":"` + refreshToken + `","accessToken":"` + accessToken + `","expiresIn":"3600","username":"` + validUsername + `","refresh_token":"` + refreshToken + `","access_token":"` + accessToken + `","expires_in":"3600","token_type":"Bearer"}
				`
			} else {
				resp = `{"errorCode":60001,"errorMessage":"invalid_grant:Invalid username & password combination.","error":"invalid_grant","error_description":"60001: invalid_grant:Invalid username & password combination."}`
			}
		}
		// Output the JSON to the client
		w.Write([]byte(resp))
	}))
}

func TestAuthorizationPreservesTokensUntilExpiration(t *testing.T) {
	server := ultradnsAuthMockServer(t)
	defer server.Close()
	auth := NewAuthorization(validUsername, validPassword)

	// Inject the fake server's URL to capture requests.
	auth.BaseURL = server.URL

	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	auth.Authorize(client)
	accessToken := strings.Repeat(auth.AccessToken, 1)
	auth.Authorize(client)
	assert.Equal(t, accessToken, auth.AccessToken)
	auth.TokenExpires = 0
	auth.Authorize(client)
	assert.NotEqual(t, accessToken, auth.AccessToken)
}

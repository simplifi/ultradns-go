package ultradns

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/simplifi/ultradns-go/internal/ultradns"
	"github.com/stretchr/testify/assert"
)

const validUsername string = "good_user"
const validPassword string = "password123!"

var validAccessToken = fmt.Sprintf("%0x", rand.Int63())
var validRefreshToken = fmt.Sprintf("%0x", rand.Int63())

var correctPostBody = []byte(`{"probing":"enable"}`)

func ultradnsMockServer(t *testing.T) *httptest.Server {
	var resp string
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+validAccessToken {
			w.WriteHeader(400)
			w.Write([]byte(`{"errorCode":60004,"errorMessage":"Authorization Header required"}`))
			return
		}
		switch {
		case r.RequestURI == "/foo":
			resp = `{"fooBar":"isFooBar"}`
		case r.RequestURI == "/post/endpoint" && r.Method == "POST":
			body, err := ioutil.ReadAll(r.Body)
			assert.NoError(t, err)

			if bytes.Compare(body, correctPostBody) == 0 {
				resp = `{"yep":true}`
			} else {
				resp = `{"yep":false}`
			}
		default:
			// This default case makes it easier to diagnose when new tests fail.
			w.WriteHeader(400)
			resp = `{"error":"wrong URL","url":"` + r.RequestURI + `"}`
		}

		w.Write([]byte(resp))
	}))
}

// Stub out an authorization that will avoid trying to make Auth calls.
func validAuthorization() *ultradns.Authorization {
	return &ultradns.Authorization{
		Username:     validUsername,
		Password:     validPassword,
		AccessToken:  validAccessToken,
		RefreshToken: validRefreshToken,
		TokenExpires: time.Now().Unix() + 3600,
	}
}

// Create the mock server and return it and stubbed APIConnection that points to it.
func stubbedServerAndAPIConn(t *testing.T) (*httptest.Server, *APIConnection) {
	server := ultradnsMockServer(t)

	auth := validAuthorization()
	auth.BaseURL = server.URL

	apiConn := NewAPIConnection(&APIOptions{})
	apiConn.BaseURL = server.URL
	apiConn.Authorization = auth

	return server, apiConn
}

func TestClientGetSendsAuthToken(t *testing.T) {
	server, apiConn := stubbedServerAndAPIConn(t)
	defer server.Close()

	resp, err := apiConn.Get("/foo")
	assert.NoError(t, err)
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, `{"fooBar":"isFooBar"}`, string(bodyBytes))
}

func TestClientGetInvalidTokenReturnsError(t *testing.T) {
	server := ultradnsMockServer(t)
	auth := validAuthorization()
	auth.BaseURL = server.URL
	auth.AccessToken = "invalid"

	apiConn := NewAPIConnection(&APIOptions{})
	apiConn.BaseURL = server.URL
	apiConn.Authorization = auth

	_, err := apiConn.Get("/foo")
	assert.Error(t, err)
	// Ensure that the error message is parsed correctly from the expected error JSON:
	assert.Equal(t, err.Error(), "60004: Authorization Header required")
}

func TestClientPostReturnsYepTrue(t *testing.T) {
	server, apiConn := stubbedServerAndAPIConn(t)
	defer server.Close()

	resp, err := apiConn.Post("/post/endpoint", bytes.NewBuffer(correctPostBody))
	assert.NoError(t, err)
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, `{"yep":true}`, string(bodyBytes))
}

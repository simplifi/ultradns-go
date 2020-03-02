package ultradns

import (
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

func ultradnsMockServer(t *testing.T) *httptest.Server {
	var resp string
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+validAccessToken {
			fmt.Println(r.Header)
			w.WriteHeader(400)
			w.Write([]byte(`{"errorCode":60004,"errorMessage":"Authorization Header required"}`))
			return
		}
		switch r.RequestURI {
		case "/foo":
			resp = `{"fooBar":"isFooBar"}`
		default:
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
func TestClientGetSendsAuthToken(t *testing.T) {
	server := ultradnsMockServer(t)
	defer server.Close()

	auth := validAuthorization()
	auth.BaseURL = server.URL

	apiConn := APIConnection{
		Client: &http.Client{
			Timeout: 1 * time.Second,
		},
		Authorization: auth,
		BaseURL:       server.URL,
	}

	resp, err := apiConn.Get("/foo")
	assert.NoError(t, err)
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, `{"fooBar":"isFooBar"}`, string(bodyBytes))
}

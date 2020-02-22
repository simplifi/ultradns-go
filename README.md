# ultradns-go

Provides an UltraDNS API package called `ultradns` for go

## Usage

Import:
```go
import "github.com/simplifi/ultradns-go"
```

To do a basic API connection and call, construct an ultradns.APIConnection object.

ultradns.APIOptions has reasonable defaults for most things.
You must pass in either the Username/Password or RefreshToken. If you pass both, the RefreshToken takes precedence.

This project only supports the JSON API request/response for UltraDNS, not the optional XML format.

```go
apiConn := ultradns.NewAPIConnection(&ultradns.APIOptions{
  // API Username.
  Username: "",

  // API Password. 
  Password: "",

  // This refresh token can be used in lieu of username/password. To initially get a RefreshToken requires a
  // Username/Password, however.
  RefreshToken: "",

  // Defaults to "https://api.ultradns.com". Typically only overridden to allow testing or to use a sandbox endpoint.
  BaseURL: "https://api.ultradns.com",

  // Timeout is a time.Duration given to the underlying http.Client
  Timeout: 5 * time.Second,
})

// apiConn has a similar API to go's net/http library

// Send a GET request to the given path. Returns a http.Response pointer and error.
resp, err := apiConn.Get("/some/api/path")

// Send a POST request to the given path. Returns a http.Response pointer and error.
resp, err := apiConn.Post("/some/api/path", json_to_send)
```
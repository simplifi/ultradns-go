// This main package is mostly for debugging or verifying that you have the correct credentials.
package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/simplifi/ultradns-go/pkg/ultradns"
)

func main() {
	userPtr := flag.String("user", "", "Username for authenticating")
	passPtr := flag.String("pass", "", "Password for authenticating")
	//verbPtr := flag.String("request", "GET", "Specify an HTTP verb (GET/POST)")
	//pathPtr := flag.String("path", "/status", "Path for the HTTP request")
	baseURLPtr := flag.String("base-url", "https://api.ultradns.com", "UltraDNS API base URL (optional)")
	//payloadPtr := flag.String("payload", "", "Payload for POST requests")
	timeoutPtr := flag.Int("timeout", 5, "Timeout in seconds for HTTP requests")

	flag.Parse()

	timeout := time.Second * time.Duration(*timeoutPtr)

	apiConn := ultradns.NewAPIConnection(&ultradns.APIOptions{
		Username: *userPtr,
		Password: *passPtr,
		BaseURL:  *baseURLPtr,
		Timeout:  timeout,
	})

	// TODO: This is just a hack for now because the API isn't complete.
	err := apiConn.Authorization.Authorize(apiConn.Client)

	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully authenticated:")
	fmt.Println(apiConn.Authorization)
}

// Calls the status endpoint and outputs the result
package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/simplifi/ultradns-go/pkg/ultradns"
)

func main() {
	// Use a couple of flags to allow passing in the username/password.
	userPtr := flag.String("user", "", "Username for UltraDNS API")
	passPtr := flag.String("pass", "", "Password for UltraDNS API")

	flag.Parse()

	if *userPtr == "" || *passPtr == "" {
		flag.PrintDefaults()
		return
	}

	// Create an APIConnection with the username/password provided.
	apiConn := ultradns.NewAPIConnection(&ultradns.APIOptions{
		Username: *userPtr,
		Password: *passPtr,
	})

	// Make a GET request to the /status endpoint
	resp, err := apiConn.Get("/status")
	if err != nil {
		fmt.Print("Error in apiConn.Get: ")
		fmt.Println(err)
		return
	}
	// The body of the response is an http.Response.Body ReadCloser. It needs to be closed after being read.
	defer resp.Body.Close()
	// Read the body into a byte array.
	var bodyBytes []byte
	if bodyBytes, err = ioutil.ReadAll(resp.Body); err != nil {
		fmt.Printf("Error in ioutil.ReadAll: %s :\n", err)
		return
	}

	// Print out the body of the response (should be `{"message":"Good"}`)
	fmt.Println("Success:")
	fmt.Println(string(bodyBytes))
}

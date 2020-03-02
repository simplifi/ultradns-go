// Calls the status endpoint and outputs the result
package main

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/simplifi/ultradns-go/pkg/ultradns"
)

func main() {
	userPtr := flag.String("user", "", "Username for UltraDNS API")
	passPtr := flag.String("pass", "", "Password for UltraDNS API")

	flag.Parse()

	if *userPtr == "" || *passPtr == "" {
		flag.PrintDefaults()
		return
	}

	apiConn := ultradns.NewAPIConnection(&ultradns.APIOptions{
		Username: *userPtr,
		Password: *passPtr,
	})

	resp, err := apiConn.Get("/status")
	if err != nil {
		fmt.Print("Error in apiConn.Get: ")
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	var bodyBytes []byte
	if bodyBytes, err = ioutil.ReadAll(resp.Body); err != nil {
		fmt.Printf("Error in ioutil.ReadAll: %s :\n", err)
		return
	}

	if resp.StatusCode >= 400 {
		fmt.Printf("HTTP Error Response %d\n", resp.StatusCode)
	} else {
		fmt.Println("Success:")
		fmt.Println(string(bodyBytes))
	}
}

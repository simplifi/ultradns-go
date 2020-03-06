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
	zonePtr := flag.String("zone", "", "Zone to look at, e.g. your main domain, like 'example.com'")
	trafficControllerPtr := flag.String("tc-name", "", "Address of Traffic Controller to query, e.g. 'my.example.com'")

	flag.Parse()

	if *userPtr == "" || *passPtr == "" || *zonePtr == "" {
		flag.PrintDefaults()
		return
	}

	var url string
	// If the traffic controller name
	if *trafficControllerPtr == "" {
		fmt.Println("No -tc-name option passed, listing all TrafficController pools.")
		url = "/zones/" + *zonePtr + "/rrsets?q=kind:TC_POOLS"
	} else {
		// This just assumes an 'A' record for simplicity.
		url = "/zones/" + *zonePtr + "/rrsets/A/" + *trafficControllerPtr
	}

	// Create an APIConnection with the username/password provided.
	apiConn := ultradns.NewAPIConnection(&ultradns.APIOptions{
		Username: *userPtr,
		Password: *passPtr,
	})
	resp, err := apiConn.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var bodyBytes []byte
	if bodyBytes, err = ioutil.ReadAll(resp.Body); err != nil {
		fmt.Printf("Error in ioutil.ReadAll: %s\n", err)
	}

	fmt.Println(string(bodyBytes))
}

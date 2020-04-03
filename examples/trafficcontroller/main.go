// Example of how to update a Trafficcontroller.
// Compile with `make trafficcontroller`

package main

import (
	"bytes"
	"encoding/json"
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
	addIPPtr := flag.String("add-ip", "", "IP Address to add to the trafficcontroller")

	flag.Parse()

	if *userPtr == "" || *passPtr == "" || *zonePtr == "" {
		flag.PrintDefaults()
		return
	}

	// Return all the pools if a specific one wasn't requested.
	var url string
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

	// Some ugliness here, but it's just an example of API usage.
	switch {
	case *addIPPtr != "":
		patchArr := make([]Patch, 2)
		patchArr[0] = Patch{
			Op:    "add",
			Path:  "/rdata/0",
			Value: *addIPPtr,
		}
		patchArr[1] = Patch{
			Op:   "add",
			Path: "/profile/rdataInfo/0",
			Value: map[string]interface{}{
				"state":    "NORMAL",
				"priority": 1,
			},
		}

		body, err := json.Marshal(patchArr)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Sending JSON PATCH request to %s with body %s\n", url, body)
		resp, err := apiConn.JSONPatch(url, bytes.NewReader(body))
		if err != nil {
			panic(err)
		}
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Update returned %s\n", string(bodyBytes))
	}

	resp, err := apiConn.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var bodyBytes []byte
	if bodyBytes, err = ioutil.ReadAll(resp.Body); err != nil {
		fmt.Printf("Error in ioutil.ReadAll: %s\n", err)
	}

	fmt.Printf("New TrafficController configuration: %s", string(bodyBytes))
}

// Patch is the object structure required for the TrafficController update JSON Patch API.
type Patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

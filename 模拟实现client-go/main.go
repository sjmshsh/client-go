package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const token = "eyJhbGciOiJSUzI1NiIsImtpZCI6ImFRM2J0Z3NmUk1hR2VhV2VRbE5vbkVHbGRSMUIwdEdTU3ZPb21TSXEtMkUifQ"
const apiServer = "https://127.0.0.1:55429"

type Pod struct {
	Metadata struct {
		Name              string    `json:"name"`
		Namespace         string    `json:"namespace"`
		CreationTimestamp time.Time `json:"creationTimestamp"`
	} `json:"metadata"`
}

type Event struct {
	EventType string `json:"type"`
	Object    Pod    `json:"object"`
}

func main() {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	req, err := http.NewRequest(http.MethodGet, apiServer+"/api/v1/namespaces/default/pods?watch=true", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "Brearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var event Event
	decoder := json.NewDecoder(resp.Body)

	for {
		if err := decoder.Decde(&event); err != nil {
			panic(err)
		}
		fmt.Printf("%s Pod %s \n", event.EventType, event.Object.Metadata)
	}
}

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Update with your configuration
const (
	updateFrequencyMins = 2
	dnsZone             = ""
	authEmail           = ""
	authKey             = ""
	identifier          = ""
	recordType          = ""
	subdomain           = ""
)

const (
	noIPAddressFound      = "No external IP address found"
	iPAddressUpdated      = "IP address updated to %s - %s"
	errorupdatingIP       = "Error updating IP address - %s"
	externalIPProviderURL = "http://ipecho.net/plain"
	cloudflareURL         = "https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s"
)

func main() {
	for range time.Tick(time.Minute * updateFrequencyMins) {
		IPAddress, err := getIPAddress()
		if err != nil {
			log.Fatal(err)
		}

		result, err := updateDNS(IPAddress)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(result)
	}
}

func getIPAddress() (string, error) {
	res, err := http.Get(externalIPProviderURL)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return "", err
		}

		return string(bodyBytes), nil
	}

	return "", errors.New(noIPAddressFound)
}

func updateDNS(IPAddress string) (string, error) {
	type Payload struct {
		Type    string `json:"type"`
		Name    string `json:"name"`
		Content string `json:"content"`
		TTL     int    `json:"ttl"`
		Proxied bool   `json:"proxied"`
	}

	data := Payload{recordType, subdomain, IPAddress, 120, false}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	body := bytes.NewReader(payloadBytes)
	reqURL := fmt.Sprintf(cloudflareURL, dnsZone, identifier)

	req, err := http.NewRequest("PUT", reqURL, body)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth-Email", authEmail)
	req.Header.Set("X-Auth-Key", authKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	t := time.Now()

	if res.StatusCode == http.StatusOK {
		successMsg := fmt.Sprintf(iPAddressUpdated, IPAddress, t)
		return successMsg, nil
	}

	errorMsg := fmt.Sprintf(errorupdatingIP, t)
	return "", errors.New(errorMsg)

}

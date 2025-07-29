package monitor

import (
	"NFT_Bot/src/misc"
	"NFT_Bot/src/structs"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"time"
)

type attributes struct {
	Attributes []struct {
		TraitType string      `json:"trait_type"`
		Value     interface{} `json:"value"`
	} `json:"attributes"`
}

var httpClient = http.Client{Timeout: 2000 * time.Millisecond}

func monitorOriginalAPI(monitorChan chan monitorResult, terminateChan chan bool, tokenURI structs.TokenURI) {

	att, err := sendApiRequest(tokenURI.BaseURI, tokenURI.Appender)

	if err != nil {
		misc.PrintRed("Error starting Original API monitor: "+err.Error(), true)
		return
	}

	for {
		select {
		case <-terminateChan:
			return
		default:
			newAtt, err := sendApiRequest(tokenURI.BaseURI, tokenURI.Appender)

			if err != nil {
				misc.PrintRed("Error monitoring original API: "+err.Error(), true)
				continue
			}

			if !isSimilar(newAtt, att) {
				monitorChan <- monitorResult{
					TokenURI: tokenURI.BaseURI,
					Appender: tokenURI.Appender,
					IsIPFS:   tokenURI.IsIpfs,
					Mode:     "By Original Api",
				}
			}

			time.Sleep(200 * time.Millisecond)
		}
	}
}

func isSimilar(attributeOne attributes, attributeTwo attributes) bool {
	if len(attributeOne.Attributes) != len(attributeTwo.Attributes) {
		return false
	}
	for i := 0; i != len(attributeOne.Attributes); i++ {
		if attributeOne.Attributes[i].TraitType != attributeTwo.Attributes[i].TraitType {
			return false
		}
		if attributeOne.Attributes[i].Value != attributeTwo.Attributes[i].Value {
			return false
		}
	}
	return true
}

func sendApiRequest(tokenURI string, appender string) (attributes, error) {
	url := tokenURI + "1"

	if len(appender) > 0 {
		tokenURI += appender
	}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return attributes{}, err
	}

	req.Header.Add("accept", "application/json")
	res, err := httpClient.Do(req)

	if err != nil {
		if os.IsTimeout(err) {
			return sendApiRequest(tokenURI, appender)
		}
		return attributes{}, err
	}

	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		time.Sleep(1000 * time.Millisecond)
		return attributes{}, errors.New(res.Status)
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return attributes{}, err
	}

	parsedBody, err := parseResponse(body)

	if err != nil {
		return attributes{}, err
	}

	return parsedBody, nil
}

// Converts the response to a struct and returns it
func parseResponse(body []byte) (attributes, error) {
	var response attributes
	err := json.Unmarshal(body, &response)
	if err != nil {
		return attributes{}, err
	}
	return response, nil
}

package requests

import (
	"NFT_Bot/src/api/api_structs"
	"NFT_Bot/src/api/request_handler/http_extra"
	"NFT_Bot/src/api/request_handler/http_parsers"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Sends a request to opensea with the passed method and parameters.
// The request will be unmarshalled as an openseaResponse and returned
func SendOpenSeaRequest(method string, params string, openSeaKey string) (api_structs.OpenSeaResponse, error) {
	url := fmt.Sprintf("https://api.opensea.io/%s/%s", method, params)

	req, err := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("X-API-KEY", openSeaKey)

	if err != nil {
		return api_structs.OpenSeaResponse{}, err
	}

	res, err := http_extra.HttpClient.Do(req) // httpClient defined in etherscan.go

	if err != nil {
		if os.IsTimeout(err) {
			return SendOpenSeaRequest(method, params, openSeaKey)
		}
		return api_structs.OpenSeaResponse{}, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		if res.StatusCode == 429 {
			time.Sleep(1000 * time.Millisecond)
			return SendOpenSeaRequest(method, params, openSeaKey)
		}
		return api_structs.OpenSeaResponse{}, errors.New("error sending request: " + strconv.Itoa(res.StatusCode))
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		return api_structs.OpenSeaResponse{}, fmt.Errorf("statuscode: %d", res.StatusCode)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		return api_structs.OpenSeaResponse{}, err
	}

	parsedBody, err := http_parsers.ParseOpenseaResponse(body)

	if err != nil {
		return api_structs.OpenSeaResponse{}, err
	}
	if !parsedBody.Success {
		return api_structs.OpenSeaResponse{}, errors.New("opensea request error")
	}

	return parsedBody, nil
}

func SendOpenseaPostRequest(domain string, payload string, openSeaKey string) (api_structs.OpenseaPostResponse, error) {
	payloadReader := strings.NewReader(payload)
	url := "https://api.opensea.io/" + domain
	req, err := http.NewRequest("POST", url, payloadReader)

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("X-API-KEY", openSeaKey)

	if err != nil {
		return api_structs.OpenseaPostResponse{}, err
	}

	res, err := http_extra.HttpClient.Do(req) // httpClient defined in etherscan.go

	if err != nil {
		if os.IsTimeout(err) {
			return SendOpenseaPostRequest(domain, payload, openSeaKey)
		}
		return api_structs.OpenseaPostResponse{}, err
	}

	if res.StatusCode == 429 {
		time.Sleep(1000 * time.Millisecond)
		return SendOpenseaPostRequest(domain, payload, openSeaKey)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		return api_structs.OpenseaPostResponse{}, err
	}

	return http_parsers.ParseOpenseaPostResponse(body)
}

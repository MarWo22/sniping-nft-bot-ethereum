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
)

// Sends a request to opensea with the passed method and parameters.
// The request will be unmarshalled as an openseaResponse and returned
func SendBlurGetRequest(method string, params string, blurKey string, authToken string, walletAddress string) (api_structs.BlurResponse, error) {
	url := fmt.Sprintf("https://blur.p.rapidapi.com/%s/%s", method, params)

	req, err := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("X-RapidAPI-Key", blurKey)
	req.Header.Add("X-RapidAPI-Host", "blur.p.rapidapi.com")

	if authToken != "" {
		req.Header.Add("authToken", authToken)
	}
	if walletAddress != "" {
		req.Header.Add("walletAddress", walletAddress)
	}

	if err != nil {
		return api_structs.BlurResponse{}, err
	}

	res, err := http_extra.HttpClientHighTimeout.Do(req) // httpClient defined in etherscan.go

	if err != nil {
		if os.IsTimeout(err) {
			return SendBlurGetRequest(method, params, blurKey, authToken, walletAddress)
		}
		return api_structs.BlurResponse{}, err
	}

	if res.StatusCode >= 400 {
		return api_structs.BlurResponse{}, errors.New("status code " + strconv.Itoa(res.StatusCode))
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		return api_structs.BlurResponse{}, err
	}

	parsedBody, err := http_parsers.ParseBlurResponse(body)

	if err != nil {
		return api_structs.BlurResponse{}, err
	}

	return parsedBody, nil
}

func SendBlurPostRequest(domain string, payload string, blurKey string, authToken string, walletAddress string) (api_structs.BlurPostResponse, error) {
	payloadReader := strings.NewReader(payload)
	url := "https://blur.p.rapidapi.com/" + domain
	req, err := http.NewRequest("POST", url, payloadReader)

	req.Header.Add("content-type", "application/json")
	req.Header.Add("X-RapidAPI-Host", "blur.p.rapidapi.com")
	req.Header.Add("X-RapidAPI-Key", blurKey)
	if authToken != "" {
		req.Header.Add("authToken", authToken)
	}
	if walletAddress != "" {
		req.Header.Add("walletAddress", walletAddress)
	}

	if err != nil {
		return api_structs.BlurPostResponse{}, err
	}

	res, err := http_extra.HttpClientHighTimeout.Do(req) // httpClient defined in etherscan.go

	if err != nil {
		if os.IsTimeout(err) {
			return SendBlurPostRequest(domain, payload, blurKey, authToken, walletAddress)
		}
		return api_structs.BlurPostResponse{}, err
	}

	if res.StatusCode >= 400 {
		return api_structs.BlurPostResponse{}, errors.New("status code " + strconv.Itoa(res.StatusCode))
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return api_structs.BlurPostResponse{}, err
	}
	return http_parsers.ParseBlurPostResponse(body)
}

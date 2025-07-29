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
	"time"
)

// Sends a request to the etherscan api
func SendEtherscanRequest(apiKey string, module string, action string, parameters string) (api_structs.EtherscanResponse, error) {
	url := fmt.Sprintf("https://api.etherscan.io/api/?module=%s&action=%s&apikey=%s%s", module, action, apiKey, parameters)
	req, err := http.NewRequest("GET", url, nil)
	//Hopefully stops from getting no response. These are the same headers as used in chrome, where it does get a response
	req.Header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Add("accept-language", "nl-NL,nl;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Add("cache-control", "max-age=0")
	req.Header.Add("sec-ch-ua", "\"Not?A_Brand\";v=\"8\", \"Chromium\";v=\"108\", \"Google Chrome\";v=\"108\"")
	req.Header.Add("sec-ch-ua-mobile", "?0")
	req.Header.Add("sec-ch-ua-platform", "\"Windows\"")
	req.Header.Add("sec-fetch-dest", "document")
	req.Header.Add("sec-fetch-mode", "navigate")
	req.Header.Add("sec-fetch-site", "none")
	req.Header.Add("sec-fetch-user", "?1")
	req.Header.Add("upgrade-insecure-requests", "1")

	if err != nil {
		return api_structs.EtherscanResponse{}, err
	}
	res, err := http_extra.HttpClient.Do(req)

	if err != nil {
		if os.IsTimeout(err) {
			return SendEtherscanRequest(apiKey, module, action, parameters)
		}
		return api_structs.EtherscanResponse{}, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		return api_structs.EtherscanResponse{}, err
	}

	parsedBody, err := http_parsers.ParseEtherscanResponse(body)

	if err != nil {
		return api_structs.EtherscanResponse{}, err
	}

	if parsedBody.Status == "0" {
		if parsedBody.Result == "Max rate limit reached" {
			time.Sleep(1000 * time.Millisecond)
			return SendEtherscanRequest(apiKey, module, action, parameters)
		}
		return api_structs.EtherscanResponse{}, errors.New(parsedBody.Message)
	}

	if parsedBody.Error != (api_structs.EtherscanError{}) {
		return api_structs.EtherscanResponse{}, errors.New(parsedBody.Error.Message)
	}

	return parsedBody, nil
}

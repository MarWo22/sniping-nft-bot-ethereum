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
	"strings"
)

// Sends a request to the Alchemy RPC endpoint using the given method and params
// Returns the result string output
func SendAlchemyRequest(method string, params string, apiKey string) (api_structs.AlchemyResponse, error) {
	var payload *strings.Reader

	if len(params) > 0 {
		payload = strings.NewReader(fmt.Sprintf("{\"id\":1,\"jsonrpc\":\"2.0\",\"params\":%s,\"method\":\"%s\"}", params, method))
	} else {
		payload = strings.NewReader(fmt.Sprintf("{\"id\":1,\"jsonrpc\":\"2.0\",\"method\":\"%s\"}", method))
	}
	url := "https://eth-mainnet.g.alchemy.com/v2/" + apiKey
	req, err := http.NewRequest("POST", url, payload)

	if err != nil {
		return api_structs.AlchemyResponse{}, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	res, err := http_extra.HttpClient.Do(req)

	if err != nil {
		if os.IsTimeout(err) {
			return SendAlchemyRequest(method, params, apiKey)
		}
		return api_structs.AlchemyResponse{}, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)

	if err != nil {
		return api_structs.AlchemyResponse{}, err
	}

	parsedBody, err := http_parsers.ParseAlchemyResponse(body)

	if err != nil {
		return api_structs.AlchemyResponse{}, err
	}

	if parsedBody.Error != (api_structs.ErrorAlchemy{}) {
		return api_structs.AlchemyResponse{}, errors.New(parsedBody.Error.Message)
	}

	return parsedBody, nil
}

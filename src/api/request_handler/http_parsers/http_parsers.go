package http_parsers

import (
	"NFT_Bot/src/api/api_structs"
	"encoding/json"
)

// Converts the response to a struct and returns it
func ParseAlchemyResponse(body []byte) (api_structs.AlchemyResponse, error) {
	var response api_structs.AlchemyResponse
	err := json.Unmarshal(body, &response)
	if err != nil {
		return api_structs.AlchemyResponse{}, err
	}
	return response, nil
}

// Converts the response to a struct and returns it
func ParseEtherscanResponse(body []byte) (api_structs.EtherscanResponse, error) {
	var response api_structs.EtherscanResponse
	err := json.Unmarshal(body, &response)
	if err != nil {
		return api_structs.EtherscanResponse{}, err
	}
	return response, nil
}

// Converts the response to a struct and returns it
func ParseBlurResponse(body []byte) (api_structs.BlurResponse, error) {
	var response api_structs.BlurResponse
	err := json.Unmarshal(body, &response)
	if err != nil {
		return api_structs.BlurResponse{}, err
	}
	return response, nil
}

// Converts the response to a struct and returns it
func ParseBlurPostResponse(body []byte) (api_structs.BlurPostResponse, error) {
	var response api_structs.BlurPostResponse
	err := json.Unmarshal(body, &response)
	if err != nil {
		return api_structs.BlurPostResponse{}, err
	}

	return response, nil
}

// Converts the response to a struct and returns it
func ParseOpenseaResponse(body []byte) (api_structs.OpenSeaResponse, error) {
	response := api_structs.OpenSeaResponse{Success: true}
	err := json.Unmarshal(body, &response)
	if err != nil {
		return api_structs.OpenSeaResponse{}, err
	}
	return response, nil
}

func ParseOpenseaPostResponse(body []byte) (api_structs.OpenseaPostResponse, error) {
	var response api_structs.OpenseaPostResponse
	err := json.Unmarshal(body, &response)
	if err != nil {
		return api_structs.OpenseaPostResponse{}, err
	}

	return response, nil
}

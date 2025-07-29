package websocket_parsers

import (
	api "NFT_Bot/src/api/api_structs"
	"encoding/json"
)

// Converts a slice of bytes to an AlchemyWebsocketResponse
// It returns an error if it fails to do so
func ParseAlchemyResponse(body []byte) (api.AlchemyWebsocketResponse, error) {
	var response api.AlchemyWebsocketResponse
	err := json.Unmarshal(body, &response)
	if err != nil {
		return api.AlchemyWebsocketResponse{}, err
	}
	return response, nil
}

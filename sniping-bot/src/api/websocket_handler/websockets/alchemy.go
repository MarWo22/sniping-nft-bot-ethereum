package websockets

import (
	"NFT_Bot/src/api/api_structs"
	"NFT_Bot/src/api/websocket_handler/websocket_parsers"
	"NFT_Bot/src/misc"

	"github.com/gorilla/websocket"
)

// Creates a websocket connection to Alchemy's websocket API
// Uses params as the parameters of the connection
// Any detected outputs are converted to AlchemyWebsocketResponse structs and redirected to the output channel
// The websocket can be terminated by sending 'true' to the terminator channel
func CreateAlchemyWebsocketClient(params string, output chan api_structs.AlchemyWebsocketResponse, apiKey string, terminator chan bool) {
	socketURL := "wss://eth-mainnet.g.alchemy.com/v2/" + apiKey
	conn, _, err := websocket.DefaultDialer.Dial(socketURL, nil)
	if err != nil {
		misc.PrintRed("Error connecting to Alchemy websocket Server: "+err.Error(), true)
		return
	}
	defer conn.Close()
	payload := "{\"jsonrpc\":\"2.0\",\"id\": 2, \"method\": \"eth_subscribe\", \"params\": " + params + "}"
	err = conn.WriteMessage(websocket.TextMessage, []byte(payload))
	go receiveHandler(conn, output)
	if err != nil {
		misc.PrintRed("Error during writing to websocket: "+err.Error(), true)
		return
	}

	// terminates when read from channel, and adds true again to prevent other routines who may rely on it from getting stuck
	<-terminator
	close(output)
}

// Reads input received from the websocket connection, and redirects
// them to the channel as AlchemyWebsocketResponse structs
func receiveHandler(connection *websocket.Conn, output chan api_structs.AlchemyWebsocketResponse) {
	// Skip first read message (subcription active message)
	connection.ReadMessage()
	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			return
		}
		parsedMsg, err := websocket_parsers.ParseAlchemyResponse(msg)
		if err != nil {
			misc.PrintRed("Error parsing Alchemy response: "+err.Error(), true)
		} else {
			output <- parsedMsg
		}
	}
}

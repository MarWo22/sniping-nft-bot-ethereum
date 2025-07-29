package api

import (
	"NFT_Bot/src/api/api_structs"
	"NFT_Bot/src/api/websocket_handler/websockets"
	"strings"
)

func PendingWebsocket(receiver string, sender string, alchemyKey string, terminateChan chan bool) chan api_structs.AlchemyWebsocketResponse {
	var params string
	if sender != "" {
		params = "[\"alchemy_pendingTransactions\", {\"toAddress\": \"" + receiver + "\", \"fromAddress\": \"" + sender + "\", \"hashesOnly\": false}]"
	} else {
		params = "[\"alchemy_pendingTransactions\", {\"toAddress\": \"" + receiver + "\", \"hashesOnly\": false}]"
	}
	pendingTx := make(chan api_structs.AlchemyWebsocketResponse)

	go websockets.CreateAlchemyWebsocketClient(params, pendingTx, alchemyKey, terminateChan)
	return pendingTx
}

func MinedWebsocket(receiver string, sender string, alchemyKey string, terminateChan chan bool) chan api_structs.AlchemyWebsocketResponse {
	var params string
	if sender != "" && receiver != "" {
		params = "[\"alchemy_minedTransactions\", {\"addresses\": [{\"to\": \"" + strings.ToLower(receiver) + "\", \"from\": \"" + strings.ToLower(sender) + "\"}],\"includeRemoved\": false}]"
	} else if receiver == "" {
		params = "[\"alchemy_minedTransactions\", {\"addresses\": [{\"from\": \"" + strings.ToLower(sender) + "\"}],\"includeRemoved\": true}]"
	} else if sender == "" {
		params = "[\"alchemy_minedTransactions\", {\"addresses\": [{\"to\": \"" + strings.ToLower(receiver) + "\"}],\"includeRemoved\": false}]"
	}
	minedTx := make(chan api_structs.AlchemyWebsocketResponse)

	go websockets.CreateAlchemyWebsocketClient(params, minedTx, alchemyKey, terminateChan)
	return minedTx
}

func waitForHash(hash string, minedTx chan api_structs.AlchemyWebsocketResponse, responseChan chan api_structs.AlchemyWebsocketResponse, terminateChan chan bool) {
	for {
		tx := <-minedTx
		if tx.Params.Result.Transaction.Hash == hash {
			responseChan <- tx
			terminateChan <- true
			return
		}
	}
}

func ListenForTx(sender string, hash string, alchemyKey string) chan api_structs.AlchemyWebsocketResponse {
	terminateChan := make(chan bool)
	responseChan := make(chan api_structs.AlchemyWebsocketResponse)
	minedTx := MinedWebsocket("", sender, alchemyKey, terminateChan)
	go waitForHash(hash, minedTx, responseChan, terminateChan)
	return responseChan
}

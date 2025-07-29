package api

import (
	"NFT_Bot/src/api/request_handler/requests"
	"NFT_Bot/src/structs"
	"encoding/json"
)

func SendWebhook(layout structs.DiscordLayout, webhook string) error {
	layoutString, err := json.Marshal(layout)
	if err != nil {
		return err
	}

	return requests.SendDiscordRequest(string(layoutString), webhook)
}

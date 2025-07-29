package requests

import (
	"NFT_Bot/src/api/request_handler/http_extra"
	"net/http"
	"os"
	"strings"
)

func SendDiscordRequest(payload string, webhook string) error {
	payloadReader := strings.NewReader(payload)

	req, err := http.NewRequest("POST", webhook, payloadReader)

	if err != nil {
		return err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	res, err := http_extra.HttpClient.Do(req)

	if err != nil {
		if os.IsTimeout(err) {
			return SendDiscordRequest(payload, webhook)
		}
		return err
	}

	defer res.Body.Close()

	return nil
}

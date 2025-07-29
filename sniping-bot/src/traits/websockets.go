package traits

import (
	"NFT_Bot/src/misc"
	"NFT_Bot/src/structs"
	"encoding/json"
	"strconv"

	"github.com/gorilla/websocket"
)

type Websockets struct {
	Sockets  []socket
	Response chan serverOutput
}

type socket struct {
	Request        chan tasks
	Timeout        int
	MaxIPFSTasks   int
	MaxCustomTasks int
}

type tasks struct {
	TaskLimit int      `json:"taskLimit"`
	Timeout   int      `json:"timeout"`
	BaseURI   string   `json:"baseURI"`
	Appender  string   `json:"appender"`
	IDs       []int    `json:"tasks"`
	Gateways  []string `json:"gateways"`
	IsIpfs    bool     `json:"isIpfs"`
}

type serverOutput struct {
	Index  int
	Tokens map[int]structs.Token
}

// Creates websocket connections to the worker servers
// Returns a struct holding the all socket's their individual input channel and request data
// and a single output channel for all
func InitWebsockets(servers []structs.ServerObject) Websockets {

	websockets := Websockets{
		Sockets:  []socket{},
		Response: make(chan serverOutput, 50),
	}
	index := 0
	for _, server := range servers {
		failed := make(chan bool)
		task := make(chan tasks)

		go createSocket(server, task, websockets.Response, index, failed)

		if !<-failed {
			websockets.Sockets = append(websockets.Sockets, socket{
				Request:        task,
				Timeout:        server.Timeout,
				MaxIPFSTasks:   server.MaxIpfsTasks,
				MaxCustomTasks: server.MaxCustomTasks,
			})
			index++
		}
	}

	return websockets
}

// Creates the socket connection and recieve handler. If the connection failed, the failed channel will be set to true
func createSocket(server structs.ServerObject, tasksChan chan tasks, output chan serverOutput, index int, failed chan bool) {

	socketURL := "ws://" + server.IP + ":" + strconv.Itoa(server.Port) + "/traits"
	conn, _, err := websocket.DefaultDialer.Dial(socketURL, nil)

	if err != nil {
		misc.PrintRed("Error connecting to "+server.IP+": "+err.Error(), true)
		failed <- true
		return
	}

	defer conn.Close()
	go receiveHandler(conn, output, index)

	close(failed)

	for {

		task := <-tasksChan
		if task.BaseURI == "" { // will only be "" if tasks channel got closed
			break
		}

		payload, err := json.Marshal(task)

		if err != nil {
			misc.PrintRed("Error marshalling json: "+err.Error(), true)
			continue
		}

		err = conn.WriteMessage(websocket.TextMessage, payload)
		if err != nil {
			misc.PrintRed("Error during writing to websocket: "+err.Error(), true)
			return
		}

	}

	close(output)

}

// Handler that handles the websocket responses
func receiveHandler(connection *websocket.Conn, output chan serverOutput, index int) {
	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			return
		}

		parseResponse(msg, output, index)
	}
}

// Parses the responses and converts them to a serverOutput, which is then sent to the output channel
func parseResponse(msg []byte, output chan serverOutput, index int) {
	var tokens = make(map[int]structs.Token)
	err := json.Unmarshal(msg, &tokens)

	if err != nil {
		misc.PrintRed("Error parsing server response: "+err.Error(), true)
	}

	output <- serverOutput{
		Index:  index,
		Tokens: tokens,
	}

}

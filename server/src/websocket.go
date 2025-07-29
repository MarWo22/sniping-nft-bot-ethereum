package src

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type InputTask struct {
	TaskLimit int      `json:"taskLimit"`
	Timeout   int      `json:"timeout"`
	BaseURI   string   `json:"baseURI"`
	Appender  string   `json:"appender"`
	Gateways  []string `json:"gateways"`
	IDs       []int    `json:"tasks"`
	IsIpfs    bool     `json:"isIpfs"`
}

type Output struct {
	ID       int
	Response Trait
}

// This function established a connection with the main node through the traits endpoint
func TraitsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log(err.Error())
	}

	log("Established connection")
	reader(ws)
}

func reader(conn *websocket.Conn) {
	first := true
	index := 0
	terminate := make(chan bool)
	tasks := make(chan Task)
	output := make(chan Output, 1000)

	// Starts the task manager goroutine
	go taskManager(tasks, output, terminate)
	// Starts the responder goroutine
	go responder(conn, output, terminate)

	for {
		// read in a message
		_, p, err := conn.ReadMessage()

		if err != nil {
			log("Connection closed")
			close(terminate)
			close(tasks)
			close(output)
			return
		}

		var inputTask InputTask

		err = json.Unmarshal(p, &inputTask)
		if err != nil {
			log("Error unmarshling task: " + err.Error())
			continue
		}

		if first {
			os.Mkdir("logs", os.ModePerm)
			writeFile, _ = os.Create("logs/" + time.Now().Format(time.RFC3339))
			log(createInitLog(inputTask))
			first = false
		}

		for _, ID := range inputTask.IDs {
			var baseURI string
			if inputTask.IsIpfs {
				baseURI = inputTask.Gateways[index%len(inputTask.Gateways)] + inputTask.BaseURI
			} else {
				baseURI = inputTask.BaseURI
			}

			tasks <- Task{
				BaseURI:  baseURI,
				Token:    ID,
				Appender: inputTask.Appender,
			}
		}
	}
}

// Function that sends the API responses through the websocket connection to the main node
func responder(conn *websocket.Conn, output chan Output, terminate chan bool) {
	count := 0
	for {
		select {
		case <-terminate:
			return
		default:
			if len(output) > 0 {
				traits := make(map[int]Trait)
				for len(output) > 0 { // Converts the results channel to a json format
					result := <-output
					traits[result.ID] = result.Response
					count++
				}
				jsonString, err := json.Marshal(traits) // Converts the json object back to a json string

				if err != nil {
					log(err.Error())
				}
				conn.WriteMessage(1, jsonString) // Send the json string to the main node
				log("Succesfully delivered " + strconv.Itoa(count) + " tokens!")
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

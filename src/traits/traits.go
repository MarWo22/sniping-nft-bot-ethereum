package traits

import (
	"NFT_Bot/src/constants"
	"NFT_Bot/src/misc"
	"NFT_Bot/src/structs"
	"time"

	"github.com/schollz/progressbar/v3"
)

// Grabs the traits from a collection by using the websocket servers
func GrabTraits(websockets Websockets, tokenURI structs.TokenURI, name string, gateways []string) map[int]structs.Token {

	misc.PrintYellow("Grabbing rarities", true)
	start := time.Now()

	bar := progressbar.Default(int64(tokenURI.Supply))

	traits := make(map[int]structs.Token)
	serverTasks := initTasks(tokenURI.Supply, tokenURI.IncludesZero, len(websockets.Sockets))
	count := 0

	for idx, socket := range websockets.Sockets {
		var taskLimit int
		if tokenURI.IsIpfs {
			taskLimit = socket.MaxIPFSTasks
		} else {
			taskLimit = socket.MaxCustomTasks
		}

		socket.Request <- tasks{
			Timeout:   socket.Timeout,
			TaskLimit: taskLimit,
			BaseURI:   tokenURI.BaseURI,
			Appender:  tokenURI.Appender,
			IDs:       serverTasks[idx],
			IsIpfs:    tokenURI.IsIpfs,
			Gateways:  gateways,
		}
	}

	for count != tokenURI.Supply {
		output := <-websockets.Response
		tokensRead := readFromServer(&traits, output.Tokens, name)
		count += tokensRead
		bar.Add(tokensRead)
	}

	timeElapsed := time.Since(start)
	misc.PrintGreen("Succesfully grabbed traits in "+timeElapsed.String(), true)
	validateTraits(traits, tokenURI)
	return traits
}

func readFromServer(traits *map[int]structs.Token, tokens map[int]structs.Token, name string) int {
	count := 0
	for key, token := range tokens {
		token.Name = name
		(*traits)[key] = token
		count++
	}
	return count
}

// Returns a slice containing all tokens based on the supply and whether it includes zero
func initTasks(supply int, includesZero bool, nWebsockets int) [][]int {
	tasks := make([][]int, nWebsockets)

	increment := 0
	if !includesZero {
		increment = 1
	}

	for i := 0; i != supply; i++ {
		tasks[i%nWebsockets] = append(tasks[i%nWebsockets], i+increment)
	}

	return tasks
}

// Makes sure we don't continue on a collections that has too many unrevealed/broken tokens
// We exit the program if that is the case
func validateTraits(traits map[int]structs.Token, tokenURI structs.TokenURI) {
	if float32(tokenURI.Supply-len(traits)) > (float32(tokenURI.Supply) * (constants.MAX_MISSING_PERCENTAGE / 100.0)) {
		misc.PrintRed("Too many missing tokens, quitting", true)
		misc.ExitProgram()
	}
}

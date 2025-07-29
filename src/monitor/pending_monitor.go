package monitor

import (
	"NFT_Bot/src/abi_handler"
	"NFT_Bot/src/api"
	"NFT_Bot/src/api/api_structs"
	"NFT_Bot/src/misc"
	"NFT_Bot/src/structs"
	"encoding/hex"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

type pendingMonitorReturn struct {
	BaseURI string
	IsIpfs  bool
}

func monitorPending(monitorChan chan monitorResult, terminateChan chan bool, contract string, abi api_structs.Abi, pendingFunction string, tokenURI structs.TokenURI, alchemyKey string) {

	pendingChan := api.PendingWebsocket(contract, "", alchemyKey, terminateChan)

	keccak := crypto.Keccak256([]byte(pendingFunction))
	keccakString := "0x" + hex.EncodeToString(keccak)[:8]
	for {
		select {
		case <-terminateChan:
			return
		default:
			pendingTokenURI := parseTransaction((<-pendingChan).Params.Result.Input, abi, keccakString)

			if pendingTokenURI != (pendingMonitorReturn{}) {
				monitorChan <- monitorResult{
					TokenURI: pendingTokenURI.BaseURI,
					Appender: tokenURI.Appender,
					IsIPFS:   pendingTokenURI.IsIpfs,
					Mode:     "By Pending Transaction",
				}
			}

		}
	}
}

func parseTransaction(txInput string, abi api_structs.Abi, keccakFunction string) pendingMonitorReturn {
	if txInput != "" {
		if txInput[:10] != keccakFunction {
			return pendingMonitorReturn{}
		}

		abiMap, err := abi_handler.DecodeBaseUriUpdate(txInput, abi)

		if err != nil {
			misc.PrintRed("Error parsing transaction: "+err.Error(), true)
			return pendingMonitorReturn{}
		}

		// We can only parse 3 different cases
		// First case: There is only 1 entry and it is a string
		// Second case: It uses fair.xyz, in which case the elements are known
		// Third case: It has multiple entries, but only one is a string
		// Otherwise, it is skipped
		if len(abiMap) == 1 {
			for key := range abiMap {
				if reflect.TypeOf(abiMap[key]).Kind() == reflect.String {
					return parsePendingTokenURI(abiMap[key].(string))
				}
			}
		} else {
			if txInput[:10] == "0xd7818e28" || txInput[:10] == "0x6724f4b7" { // fair.xyz
				return parsePendingTokenURI(abiMap["newPathURI"].(string) + abiMap["newURI"].(string))
			}
			stringCount := 0
			var stringKey string
			for key := range abiMap {
				if reflect.TypeOf(abiMap[key]).Kind() == reflect.String {
					stringCount++
					stringKey = key
				}
			}
			if stringCount == 1 {
				return parsePendingTokenURI(abiMap[stringKey].(string))
			}
		}
	}

	return pendingMonitorReturn{}
}

func parsePendingTokenURI(uri string) pendingMonitorReturn {
	ipfs := false
	if strings.Contains(uri, "ipfs") {
		ipfs = true
	}

	tokenURI := pendingMonitorReturn{
		BaseURI: uri,
		IsIpfs:  ipfs,
	}

	if !ipfs {

		if tokenURI.BaseURI[len(tokenURI.BaseURI)-1] != '/' {
			tokenURI.BaseURI += "/"
		}

		return tokenURI
	}

	splits := strings.Split(tokenURI.BaseURI, "/")

	for _, split := range splits {
		if tokenURI.IsIpfs && len(split) >= 46 {
			tokenURI.BaseURI = split
		}
	}

	tokenURI.BaseURI += "/"

	return tokenURI
}

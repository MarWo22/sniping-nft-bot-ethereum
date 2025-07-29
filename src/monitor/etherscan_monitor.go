package monitor

import (
	"NFT_Bot/src/api"
	"NFT_Bot/src/misc"
	"NFT_Bot/src/structs"
	"time"
)

func monitorEtherscan(monitorChan chan monitorResult, terminateChan chan bool, contract string, uriFunction string, tokenURI structs.TokenURI, etherscanKey string) {
	count := 0
	for {
		select {
		case <-terminateChan:
			return
		default:
			misc.PrintMonitorIteration(count)
			newURI, err := api.GetTokenUri(contract, uriFunction, etherscanKey)
			if err != nil {
				misc.PrintRed("Error monitoring Etherscan: "+err.Error(), true)
				continue
			}

			if newURI.BaseURI != tokenURI.BaseURI {
				monitorChan <- monitorResult{
					TokenURI: newURI.BaseURI,
					Appender: newURI.Appender,
					IsIPFS:   newURI.IsIPFS,
					Mode:     "By Contract",
				}
			}
			time.Sleep(100 * time.Millisecond)
			count++
		}
	}
}

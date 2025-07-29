package monitor

import (
	"NFT_Bot/src/misc"
	"NFT_Bot/src/structs"
	"NFT_Bot/src/webhooks"
	"fmt"
)

type monitorResult struct {
	TokenURI string
	Appender string
	IsIPFS   bool
	Mode     string
}

// Monitors for a reveal. It monitors for changes on Etherscan, for updateBaseURI pending transactions sent by the owner, and for changes on the original API
// If a reveal is detected, the tokenURI struct is updated with the new uri, and sends a webhook
func MonitorForReveal(contract string, tokenURI *structs.TokenURI, collection structs.Collection, contractFunctions structs.ContractFunctions, apiKeys structs.ApiKeys) {

	monitorChan := make(chan monitorResult)
	terminateChan := make(chan bool)

	if !tokenURI.IsIpfs {
		go monitorOriginalAPI(monitorChan, terminateChan, *tokenURI)
	}

	go monitorEtherscan(monitorChan, terminateChan, contract, contractFunctions.TokenURIFunction, *tokenURI, apiKeys.EtherscanKey)
	go monitorPending(monitorChan, terminateChan, contract, contractFunctions.SetBaseUriAbi, contractFunctions.SetBaseUriFunction, *tokenURI, apiKeys.AlchemyKey)

	result := <-monitorChan

	tokenURI.BaseURI = result.TokenURI
	tokenURI.Appender = result.Appender
	tokenURI.IsIpfs = result.IsIPFS

	err := webhooks.SendRevealDetectedWebhook(collection, *tokenURI, result.Mode, apiKeys.DiscordWebhook)

	fmt.Println()

	if err != nil {
		misc.PrintRed("Error sending webhook: "+err.Error(), true)
	}
	misc.PrintYellow("Detected reveal "+result.Mode, true)
	close(terminateChan)
}

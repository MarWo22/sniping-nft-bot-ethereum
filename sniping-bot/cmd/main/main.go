package main

import (
	"NFT_Bot/src/api"
	"NFT_Bot/src/constants"
	"NFT_Bot/src/file_handler"
	"NFT_Bot/src/misc"
	"NFT_Bot/src/monitor"
	"NFT_Bot/src/offers"
	"NFT_Bot/src/rarities"
	"NFT_Bot/src/structs"
	init_struct "NFT_Bot/src/structs/init"
	"NFT_Bot/src/traits"
	"NFT_Bot/src/transactions"
	"bufio"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/manifoldco/promptui"
)

const (
	WITH_MONITOR     = 0
	WITHOUT_MONITOR  = 1
	PLACE_OFFERS     = 2
	WRITE_RARITES    = 3
	CHECK_AUTHTOKENS = 4
)

func main() {
	// Initializes all terminal settings
	misc.ResetTerminal()
	// Grabs the initial data from OpenSea and Etherscan
	taskData := init_struct.InitTaskData()
	// Starts the listing node
	websockets := traits.InitWebsockets(taskData.Servers)

	// Print main menu info such as collection, uri, wallets, etc.
	printMainMenuData(taskData, len(websockets.Sockets))
	// Prompt the user for the mode
	mode := mainMenuPrompt()
	offers := false
	switch mode {
	case WITH_MONITOR:
		offers = optionalOffersPrompt()
		misc.ResetTerminal()
		if offers {
			withMonitor(taskData, websockets, true)
		} else {
			withMonitor(taskData, websockets, false)
		}

	case WITHOUT_MONITOR:
		offers = optionalOffersPrompt()
		misc.ResetTerminal()
		if offers {
			withoutMonitor(taskData, websockets, true)
		} else {
			withoutMonitor(taskData, websockets, false)
		}

	case PLACE_OFFERS:
		offers = true
		misc.ResetTerminal()
		placeOffers(taskData, websockets)

	case WRITE_RARITES:
		misc.ResetTerminal()
		writeRarities(taskData, websockets)

	case CHECK_AUTHTOKENS:
		misc.ResetTerminal()
		checkAuthTokens(taskData.Wallets, taskData.ApiKeys.BlurKey)

	}
	if offers {
		// Wait for offers to expire
		time.Sleep(time.Duration(taskData.BuySettings.OfferDuration) * time.Second)
	}
	// Terminate all websockets
	fmt.Println()
	// Enable the cursor again
	misc.ShowCursor()
	// Prompt the user to exit the program
	misc.ExitProgram()
}

func withoutMonitor(taskData structs.TaskData, websockets traits.Websockets, placeOffers bool) {
	// Grabs the trait data from the tokenURI using websockets
	traits := traits.GrabTraits(websockets, taskData.TokenURI, taskData.Collection.Name, taskData.TokenURI.IPFSGateways)
	// Calculates the rarities and ranks of the tokens
	rarities := rarities.GetRarities(traits, taskData.TokenURI.Supply)
	// Starts the transaction node that handles transactions
	transactions.FulfillAvailableListings(taskData.Contract, rarities, taskData.BuySettings, taskData.Wallets.Wallets, taskData.ApiKeys)
	// Starts the offer node that handles offers
	// Should also listen for accepted offers
	if placeOffers {
		offers.Offers(taskData.Wallets.OfferWallet, taskData.BuySettings, taskData.Collection, rarities, taskData.Contract, taskData.ApiKeys)
	}
	// Writes the rarities to a JSON file
	file_handler.WriteRarities(rarities, taskData.Collection.Slug)
}

func withMonitor(taskData structs.TaskData, websockets traits.Websockets, placeOffers bool) {
	// monitor for reveal
	monitor.MonitorForReveal(taskData.Contract, &taskData.TokenURI, taskData.Collection, taskData.ContractFunctions, taskData.ApiKeys)
	// Grabs the trait data from the tokenURI using websockets
	traits := traits.GrabTraits(websockets, taskData.TokenURI, taskData.Collection.Name, taskData.TokenURI.IPFSGateways)
	// Calculates the rarities and ranks of the tokens
	rarities := rarities.GetRarities(traits, taskData.TokenURI.Supply)
	// Starts the transaction node that handles transactions
	transactions.FulfillAvailableListings(taskData.Contract, rarities, taskData.BuySettings, taskData.Wallets.Wallets, taskData.ApiKeys)
	// Starts the offer node that handles offers
	// Should also listen for accepted offers
	if placeOffers {
		offers.Offers(taskData.Wallets.OfferWallet, taskData.BuySettings, taskData.Collection, rarities, taskData.Contract, taskData.ApiKeys)
	}
	// Writes the rarities to a JSON file
	file_handler.WriteRarities(rarities, taskData.Collection.Slug)
}

func placeOffers(taskData structs.TaskData, websockets traits.Websockets) {
	// Grabs the trait data from the tokenURI using websockets
	traits := traits.GrabTraits(websockets, taskData.TokenURI, taskData.Collection.Name, taskData.TokenURI.IPFSGateways)
	// Calculates the rarities and ranks of the tokens
	rarities := rarities.GetRarities(traits, taskData.TokenURI.Supply)
	// Starts the offer node that handles offers
	// Should also listen for accepted offers
	offers.Offers(taskData.Wallets.OfferWallet, taskData.BuySettings, taskData.Collection, rarities, taskData.Contract, taskData.ApiKeys)
}

func writeRarities(taskData structs.TaskData, websockets traits.Websockets) {
	// Grabs the trait data from the tokenURI using websockets
	traits := traits.GrabTraits(websockets, taskData.TokenURI, taskData.Collection.Name, taskData.TokenURI.IPFSGateways)
	// Calculates the rarities and ranks of the tokens
	rarities := rarities.GetRarities(traits, taskData.TokenURI.Supply)
	// Writes the rarities to a JSON file
	file_handler.WriteRarities(rarities, taskData.Collection.Slug)
}

func checkAuthTokens(wallets structs.Wallets, blurKey string) {
	misc.PrintYellow("Checking authTokens", true)
	wallets, invalidTokens := api.CheckAuthTokens(wallets, blurKey)

	if invalidTokens == 0 {
		misc.PrintGreen("\nAll authTokens are valid", false)
	} else {
		misc.PrintYellow("\n"+strconv.Itoa(invalidTokens)+" authToken(s) are invalid", false)
		amountToUpdate := updateAuthTokenPrompt()
		if amountToUpdate != 0 {
			prioritizeOfferWallet := prioritizeOfferWalletPrompt()
			customBlurKey := customBlurKeyPrompt()
			if customBlurKey == "" {
				customBlurKey = blurKey
			}
			// Needed for some reason - Otherwise next prints wont work well
			fmt.Print()
			fmt.Println()

			misc.PrintGreen("Updating authTokens", true)

			updatedWallet := api.UpdateAuthKeys(wallets, amountToUpdate, prioritizeOfferWallet, customBlurKey)
			err := init_struct.WriteWalletsJson(updatedWallet)

			if err != nil {
				misc.PrintRed("Failed to update wallet json: "+err.Error(), true)
			} else {
				misc.PrintGreen("Succesfully updated wallet json", true)
			}
		}
	}

	fmt.Println("\nPress any key to return to main menu")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadRune()

	misc.ResetTerminal()
	main()

}

func printMainMenuData(taskData structs.TaskData, websocketCount int) {
	misc.ResetTerminal()

	logo :=
		`
		  ____        _       _               _   _ _____ _____   ____        _   
		 / ___| _ __ (_)_ __ (_)_ __   __ _  | \ | |  ___|_   _| | __ )  ___ | |_ 
		 \___ \| '_ \| | '_ \| | '_ \ / _' | |  \| | |_    | |   |  _ \ / _ \| __|
		  ___) | | | | | |_) | | | | | (_| | | |\  |  _|   | |   | |_) | (_) | |_ 
		 |____/|_| |_|_| .__/|_|_| |_|\__, | |_| \_|_|     |_|   |____/ \___/ \__|
		               |_|            |___/                                       `

	fmt.Println(logo)
	fmt.Println("")
	fmt.Println("Version", constants.VERSION)
	fmt.Println("")
	fmt.Println("Loaded Collection:", taskData.Collection.Name)
	fmt.Println("Supply:", taskData.TokenURI.Supply)
	fmt.Println("")
	fmt.Println("Connected servers:", websocketCount)
	if taskData.TokenURI.IsIpfs {
		fmt.Println("TokenURI:", constants.PUBLIC_IPFS_GATEWYA+taskData.TokenURI.BaseURI+"####"+taskData.TokenURI.Appender)
	} else {
		fmt.Println("TokenURI:", taskData.TokenURI.BaseURI+"####"+taskData.TokenURI.Appender)
	}
	fmt.Println("")
	fmt.Println("Loaded Wallets For Sniping:")

	for _, wallet := range taskData.Wallets.Wallets {
		fmt.Printf("%s: %.4f ETH\n", wallet.Address, new(big.Float).Quo(new(big.Float).SetInt(wallet.Balance), big.NewFloat(constants.WEI_TO_ETH)))
	}

	fmt.Println("Loaded Wallet For Offers")
	fmt.Printf("%s: %.4f wETH\n", taskData.Wallets.OfferWallet.Address, new(big.Float).Quo(new(big.Float).SetInt(taskData.Wallets.OfferWallet.Balance), big.NewFloat(constants.WEI_TO_ETH)))
	fmt.Println("")
	fmt.Println(taskData.ContractFunctions)
}

// Prompts the user to select a mode
func mainMenuPrompt() int {
	prompt := promptui.Select{
		HideHelp: true,
		Label:    "Select Mode",
		Items:    []string{"Run with monitor", "Run without monitor", "Place offers", "Write rarities", "Check Blur Auth Tokens"},
	}
	pos, _, _ := prompt.Run()
	return pos
}

func updateAuthTokenPrompt() int {

	fmt.Print("How many authTokens to update (2 post request per authToken): ")
	var input int

	// Read a single line of input
	_, err := fmt.Scan(&input)
	if err != nil {
		fmt.Println("Please enter a valid number")
		return updateAuthTokenPrompt()
	}
	return input
}

func optionalOffersPrompt() bool {
	prompt := promptui.Select{
		HideHelp:     true,
		HideSelected: true,
		Label:        "Also Create Offers?",
		Items:        []string{"Yes", "No"},
	}
	pos, _, _ := prompt.Run()

	if pos == 0 {
		return true
	} else {
		return false
	}
}

func prioritizeOfferWalletPrompt() bool {
	prompt := promptui.Select{
		HideHelp:     true,
		HideSelected: true,
		Label:        "Prioritize Offer Wallet?",
		Items:        []string{"Yes", "No"},
	}

	pos, _, _ := prompt.Run()

	if pos == 0 {
		return true
	} else {
		return false
	}

}

func useCustomBlurKeyPrompt() bool {
	prompt := promptui.Select{
		HideHelp:     true,
		HideSelected: true,
		Label:        "Use different blur key ('No' will select the key in settings json)",
		Items:        []string{"Yes", "No"},
	}

	pos, _, _ := prompt.Run()

	if pos == 0 {
		return true
	} else {
		return false
	}
}

func customBlurKeyPrompt() string {

	if useCustomBlurKeyPrompt() {
		fmt.Print("Enter blurKey (Leave empty to use the key in settings json): ")
		var input string
		// Read a single line of input
		fmt.Scan(&input)
		return input
	}
	return ""
}

/*	----- TODO -----

EXPERIMENT WITH BLUR.io
See if we can get all marketplace orders in reasonable time
https://core-api.prod.blur.io/v1/collections/bored-ape-kennel-club/prices?filters=%7B%22traits%22%3A%5B%5D%2C%22hasAsks%22%3Atrue%7D




#1 Fix/improve/test monitor
#2 Improve/fix websocket server
#3 Improve/fix listings node
#5 Offset and Offset monitoring
#6 Add blur.io support
#7 Thorougly test
#9 Fix paying too much for collections that have few traits (look at colars)

Refactoring and rewriting code as needed
*/

/*
	Contracts with bugs:
	0xa30cf1135be5af62e412f22bd01069e2ceba8706
	0x32edd2f7437665af088347791521f454831aaa29
	0xe1d7a7c25d6bacd2af454a7e863e7b611248c3e5
*/

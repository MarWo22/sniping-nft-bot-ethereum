package init

import (
	"NFT_Bot/src/api"
	"NFT_Bot/src/constants"
	"NFT_Bot/src/misc"
	"NFT_Bot/src/structs"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

func InitTaskData() structs.TaskData {
	settings, err := readSettingsJson()
	if err != nil {
		misc.PrintRed("Error reading settings: "+err.Error(), true)
		misc.ExitProgram()
	}

	wallets, err := readWalletsJson()
	if err != nil {
		misc.PrintRed("Error reading wallets: "+err.Error(), true)
		misc.ExitProgram()
	}

	keywords, err := getKeywords(settings)
	if err != nil {
		misc.PrintRed("Error getting keywords: "+err.Error(), true)
		misc.ExitProgram()
	}
	collection := InitCollection(settings, keywords)

	return organiseStructs(settings, wallets, collection, keywords, settings.AlchemyKey)
}

func convertRanges(ranges []rangeStruct) []structs.Range {
	// This may seem overly complicated, but it is needed. Otherwise we will encounter floating point errors

	var newRanges []structs.Range
	for _, oldRange := range ranges {
		// Convert float to string
		floatString := fmt.Sprintf("%f", oldRange.Value)
		// Trim trailing zeroes and "."
		floatString = strings.TrimRight(strings.TrimRight(floatString, "0"), ".")
		// Check for dot
		dotIdx := strings.Index(floatString, ".")

		if dotIdx == -1 {
			// If there is no dot, just add the required zeroes
			floatString += strings.Repeat("0", len(strconv.Itoa(constants.WEI_TO_ETH))-1)
		} else {
			// If there is, add the required zeroes and remove any leading zeroes
			floatString += strings.Repeat("0", len(strconv.Itoa(constants.WEI_TO_ETH))-(len(floatString[dotIdx:])))
			floatString = strings.Replace(floatString, ".", "", -1)
			floatString = strings.TrimLeft(floatString, "0")
		}

		valueBig := new(big.Int)
		// Convert string to big.Int
		valueBig.SetString(floatString, 10)
		// Convert to new range
		newRanges = append(newRanges, structs.Range{
			Low:         oldRange.Low,
			High:        oldRange.High,
			Value:       valueBig,
			PriorityFee: oldRange.PriorityFee,
		})
	}
	return newRanges
}

// Returns the wallet address of the given privateKey
// Use true for wETH, false for ETH
func intializeWallet(wallet walletStruct, wETH bool, apiKey string) structs.Wallet {

	// Convert the private key to ECDSA
	privateKeyECDSA, err := crypto.HexToECDSA(wallet.PrivateKey)
	if err != nil {
		misc.PrintRed("Error converting private key to ECSDA: "+err.Error(), true)
		misc.ExitProgram()
	}

	// Get the public key from the private key
	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		misc.PrintRed("Error converting private key to public key: "+err.Error(), true)
		misc.ExitProgram()
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	var balance *big.Int
	if wETH {
		balance, err = api.GetWethBalance(address, apiKey)
	} else {
		balance, err = api.GetBalance(address, apiKey)
	}
	if err != nil {
		misc.PrintRed("Error getting balance: "+err.Error(), true)
	}

	// Convert the public key to address in hex format
	return structs.Wallet{
		PrivateKey:    wallet.PrivateKey,
		BlurAuthToken: wallet.BlurAuthToken,
		Address:       crypto.PubkeyToAddress(*publicKeyECDSA).Hex(),
		Balance:       balance,
	}
}

func organiseStructs(settings settings, wallets wallets, collection collection, keywords keywords, apiKey string) structs.TaskData {
	walletStruct := make([]structs.Wallet, len(wallets.WalletPrivateKeys))

	for idx := 0; idx != len(wallets.WalletPrivateKeys); idx++ {
		walletStruct[idx] = intializeWallet(wallets.WalletPrivateKeys[idx], false, apiKey)
	}

	contractFunctions := structs.ContractFunctions{
		SetBaseUriFunction:  keywords.UpdateBaseUri,
		TotalSupplyFunction: keywords.Supply,
		TokenURIFunction:    keywords.TokenURI,
		SetBaseUriAbi:       keywords.UpdateBaseUriAbi,
	}

	tokenURI := structs.TokenURI{
		BaseURI:      collection.TokenURI,
		Appender:     collection.Appender,
		IsIpfs:       collection.IsIpfs,
		Offset:       collection.Offset,
		Increment:    settings.Increment,
		IncludesZero: collection.IncludesZero,
		IPFSGateways: settings.IPFSGateways,
		UsesOffset:   settings.UsesOffset,
		Supply:       collection.Supply,
	}

	collectionNew := structs.Collection{
		Slug:       collection.Slug,
		Name:       collection.Name,
		Owner:      collection.Owner,
		OpenSeaFee: collection.OpenSeaFee,
		ImageURL:   collection.ImageURL,
		Fees:       collection.Fees,
	}

	walletsNew := structs.Wallets{
		Wallets:     walletStruct,
		OfferWallet: intializeWallet(wallets.OfferWalletPrivateKey, true, apiKey),
	}

	apiKeys := structs.ApiKeys{
		OpenSeaKey:     settings.OpenSeaKey,
		AlchemyKey:     settings.AlchemyKey,
		EtherscanKey:   settings.EtherscanKey,
		DiscordWebhook: settings.DiscordWebhook,
		BlurKey:        settings.BlurKey,
	}

	buySettings := structs.BuySettings{
		MaxGas:            settings.MaxGas,
		BuyingRange:       settings.BuyingRange,
		EncryptedContract: settings.EncryptedContract,
		EncryptionKey:     settings.EncryptionKey,
		Ranges:            convertRanges(settings.Ranges),
		OfferDuration:     settings.OfferDuration,
	}

	return structs.TaskData{
		ContractFunctions: contractFunctions,
		TokenURI:          tokenURI,
		Collection:        collectionNew,
		Wallets:           walletsNew,
		ApiKeys:           apiKeys,
		BuySettings:       buySettings,
		Contract:          settings.Contract,
		MonitorOffset:     settings.MonitorOffset,
		Servers:           settings.Servers,
	}
}

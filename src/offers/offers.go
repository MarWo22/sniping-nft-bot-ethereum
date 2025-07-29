package offers

import (
	"NFT_Bot/src/abi_handler"
	"NFT_Bot/src/api"
	"NFT_Bot/src/constants"
	"NFT_Bot/src/misc"
	"NFT_Bot/src/webhooks"
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"

	"NFT_Bot/src/structs"
	"math/big"
	"math/rand"
	"strconv"
)

// Creates offers for all rare listing within the buying range.
// The offer value will be the buying range values
func Offers(offerWallet structs.Wallet, buySettings structs.BuySettings, collection structs.Collection, rarities structs.Rarities, contract string, apiKeys structs.ApiKeys) {
	// Loop over all tokens withing the buying range
	terminate := make(chan bool)
	go listenForAcceptedOffers(rarities, contract, offerWallet.Address, apiKeys, terminate)
	placeOffersOpensea(offerWallet, buySettings, collection, rarities, contract, apiKeys)
}

// Calculates the amount the offer should be based on the rank and balence of the wallet
func getOfferAmount(rank int, balance *big.Int, ranges []structs.Range) *big.Int {
	// Loop over all ranges
	for _, buyRange := range ranges {
		// Check for the range it is part of
		if rank >= buyRange.Low && rank <= buyRange.High {
			// Return the value if it is lower than balance, otherwise return balance
			if buyRange.Value.Cmp(balance) <= 0 { // If listing price is smaller or equal to the buying range
				return buyRange.Value
			} else {
				return balance
			}
		}
	}

	return big.NewInt(0)
}

func placeOffersOpensea(offerWallet structs.Wallet, buySettings structs.BuySettings, collection structs.Collection, rarities structs.Rarities, contract string, apiKeys structs.ApiKeys) {
	for idx := 0; idx != buySettings.BuyingRange; idx++ {
		tokenID := rarities.Ranks[idx]
		rank := rarities.Tokens[tokenID].Rank
		// Get the wETH balence of the offer wallet
		balance, err := api.GetWethBalance(offerWallet.Address, apiKeys.AlchemyKey)

		if err != nil {
			misc.PrintRed("Error getting wETH balance: "+err.Error(), true)
			continue
		}

		// Get the offer amount bases on rank and balance
		offerAmount := getOfferAmount(rank, balance, buySettings.Ranges)
		// Place the offer
		err = placeOpenSeaOffer(offerWallet, collection, contract, tokenID, offerAmount, buySettings.OfferDuration, apiKeys.OpenSeaKey)
		// Log the outcome
		if err != nil {
			misc.PrintRed("Error creating OpenSea offer for token "+strconv.Itoa(tokenID)+" with rank "+strconv.Itoa(rank)+": "+err.Error(), true)
		} else {
			misc.PrintYellow("Successfully created OpenSea offer for token "+strconv.Itoa(tokenID)+" with rank "+strconv.Itoa(rank), true)
		}
		time.Sleep(250 * time.Millisecond)
	}
}

// Creates an offer for the given token and sends it to the OpenSea API.
func placeOpenSeaOffer(offerWallet structs.Wallet, collection structs.Collection, contract string, tokenID int, price *big.Int, offerDuration int, openSeaKey string) error {
	offer := []structs.Offer{
		{
			ItemType:             1,
			Token:                constants.WRAPPED_ETHER_ADDRESS,
			IdentifierOrCriteria: "0",
			StartAmount:          price.String(),
			EndAmount:            price.String(),
		},
	}

	// Initialize consideration slice
	consideration := []structs.Consideration{
		{
			ItemType:             2,
			Token:                contract,
			IdentifierOrCriteria: strconv.Itoa(tokenID),
			StartAmount:          "1",
			EndAmount:            "1",
			Recipient:            offerWallet.Address,
		},
	}

	for _, feeObject := range collection.Fees {
		if feeObject.Required {
			priceBig, _ := new(big.Float).SetString(price.String()[:len(price.String())-2])
			feeFloat := new(big.Float).Mul(priceBig, big.NewFloat(feeObject.Fee))
			feeInt, _ := feeFloat.Int(nil)

			consideration = append(consideration, structs.Consideration{
				ItemType:             1,
				Token:                constants.WRAPPED_ETHER_ADDRESS,
				IdentifierOrCriteria: "0",
				StartAmount:          feeInt.String(),
				EndAmount:            feeInt.String(),
				Recipient:            feeObject.Recipient,
			})
		}
	}

	// Init the OfferParameters struct
	parameters := structs.OfferParameters{
		Offerer:                         offerWallet.Address,
		Offer:                           offer,
		StartTime:                       time.Now().Unix(),
		EndTime:                         time.Now().Unix() + int64(offerDuration),
		OrderType:                       2,
		Consideration:                   consideration,
		Zone:                            constants.ZONE,
		ZoneHash:                        constants.ZONE_HASH,
		TotalOriginalConsiderationItems: strconv.Itoa(len(consideration)),
		Salt:                            generateSalt(),
		ConduitKey:                      constants.CONDUIT_KEY,
		Counter:                         0,
	}

	// Generate the signature of the OfferParameters
	signature, err := generateSignature(parameters, offerWallet.PrivateKey)

	if err != nil {
		return err
	}

	// Create the final offer object
	offerStruct := structs.OfferStruct{
		Parameters:      parameters,
		Signature:       hexutil.Encode(signature),
		ProtocolAddress: constants.OPENSEA_CONTRACT,
	}

	// Post the offer to OpenSea
	return api.CreateOffer(offerStruct, openSeaKey)
}

// Signs the OfferParameters struct using the private key
// It returns the signature as a slice of bytes, and returns an error if it failed
func generateSignature(parameters structs.OfferParameters, privateKey string) ([]byte, error) {
	// Convert the OfferParameters struct to a JSON string
	jsonStr, err := json.Marshal(parameters)

	if err != nil {
		return nil, err
	}

	// Convert the JSON string to a map
	var paramMap map[string]interface{}
	json.Unmarshal(jsonStr, &paramMap)

	// Create the typedData struct
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"OrderComponents": []apitypes.Type{
				{Name: "offerer", Type: "address"},
				{Name: "zone", Type: "address"},
				{Name: "offer", Type: "OfferItem[]"},
				{Name: "consideration", Type: "ConsiderationItem[]"},
				{Name: "orderType", Type: "uint8"},
				{Name: "startTime", Type: "uint256"},
				{Name: "endTime", Type: "uint256"},
				{Name: "zoneHash", Type: "bytes32"},
				{Name: "salt", Type: "uint256"},
				{Name: "conduitKey", Type: "bytes32"},
				{Name: "counter", Type: "uint256"},
			},
			"OfferItem": []apitypes.Type{
				{Name: "itemType", Type: "uint8"},
				{Name: "token", Type: "address"},
				{Name: "identifierOrCriteria", Type: "uint256"},
				{Name: "startAmount", Type: "uint256"},
				{Name: "endAmount", Type: "uint256"},
			},
			"ConsiderationItem": []apitypes.Type{
				{Name: "itemType", Type: "uint8"},
				{Name: "token", Type: "address"},
				{Name: "identifierOrCriteria", Type: "uint256"},
				{Name: "startAmount", Type: "uint256"},
				{Name: "endAmount", Type: "uint256"},
				{Name: "recipient", Type: "address"},
			},
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
		},
		PrimaryType: "OrderComponents",
		Domain: apitypes.TypedDataDomain{
			Name:              "Seaport",
			Version:           "1.5",
			ChainId:           math.NewHexOrDecimal256(1),
			VerifyingContract: constants.OPENSEA_CONTRACT,
		},
		Message: apitypes.TypedDataMessage{ // provide padding to 32 bytes
			"offerer":       paramMap["offerer"],
			"zone":          paramMap["zone"],
			"offer":         paramMap["offer"],
			"consideration": paramMap["consideration"],
			"orderType":     paramMap["orderType"],
			"startTime":     paramMap["startTime"],
			"endTime":       paramMap["endTime"],
			"zoneHash":      paramMap["zoneHash"],
			"salt":          paramMap["salt"],
			"conduitKey":    paramMap["conduitKey"],
			"counter":       paramMap["counter"],
		},
	}

	// Hash the TypedData struct
	hash, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, err
	}

	// Create a ECSDA private key
	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}

	// Sign the hash
	signature, err := crypto.Sign(hash, privateKeyECDSA)
	if err != nil {
		return nil, err
	}

	// Add 27 to the last byte to make the signature valid (It is needed, but I don't know why)
	signature[64] += 27

	return signature, nil
}

// Generates a random 76 digit Salt
func generateSalt() string {
	// First digit is in range [1-9]
	salt := strconv.Itoa(1 + rand.Intn(8))
	// Remaining 75 digits are in range [0-9]
	for i := 0; i != 76; i++ {
		salt += strconv.Itoa(rand.Intn(9))
	}
	return salt
}

func listenForAcceptedOffers(rarities structs.Rarities, contract string, offerAddress string, apiKeys structs.ApiKeys, terminate chan bool) {
	var duplicates []string
	block, err := api.GetBlock(apiKeys.AlchemyKey)
	if err != nil {
		misc.PrintRed("Error getting current block: "+err.Error(), true)
		return
	}
	for {
		select {
		case <-terminate:
			return
		default:
			transferEvents, err := api.GetTransferEvents(contract, offerAddress, int(block), apiKeys.EtherscanKey)
			if err != nil {
				if err.Error() != "No transactions found" {
					misc.PrintRed("Error getting transfer event: "+err.Error(), true)
				}
				time.Sleep(1500 * time.Millisecond)
				continue
			}
			for _, transferEvent := range transferEvents {
				if misc.ContainsString(duplicates, transferEvent.Hash) {
					continue
				}

				duplicates = append(duplicates, transferEvent.Hash)

				tx, err := api.GetTransactionByHash(transferEvent.Hash, apiKeys.AlchemyKey)

				if err != nil {
					misc.PrintRed("Error getting offer transaction by hash", true)
					continue
				}

				parameters, err := abi_handler.DecodeAbi(tx.Input)

				if err != nil {
					misc.PrintRed("Error decoding offer ABI: "+err.Error(), true)
					continue
				}

				// if parameters[0].BasicOrderType.Int64() != 18 {
				// 	continue
				// }

				ID, err := strconv.Atoi(transferEvent.TokenID)

				if err != nil {
					misc.PrintRed("Error converting tokenID to int: "+err.Error(), true)
				}

				misc.PrintGreen("Offer for token "+transferEvent.TokenID+" with rank "+strconv.Itoa(rarities.Tokens[ID].Rank)+" has been accepted", true)
				err = webhooks.SendOfferAcceptedWebhook(rarities.Tokens[ID], offerAddress, transferEvent.Hash, ID, contract, parameters[0].OfferAmount, apiKeys.DiscordWebhook)

				if err != nil {
					misc.PrintRed("Error sending webhook: "+err.Error(), true)
				}

			}
			time.Sleep(1500 * time.Millisecond)
		}
	}
}

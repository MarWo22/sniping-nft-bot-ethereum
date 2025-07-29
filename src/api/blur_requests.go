package api

import (
	"NFT_Bot/src/api/api_structs"
	"NFT_Bot/src/api/request_handler/requests"
	"NFT_Bot/src/constants"
	"NFT_Bot/src/misc"
	"NFT_Bot/src/structs"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func generateAuthToken(wallet structs.Wallet, blurKey string) (string, error) {

	payload := "{\"walletAddress\": \"" + wallet.Address + "\"}"

	response, err := requests.SendBlurPostRequest("auth/challenge", payload, blurKey, "", "")

	if err != nil {
		return "", err
	}

	// Create a ECSDA private key
	privateKeyECDSA, err := crypto.HexToECDSA(wallet.PrivateKey)
	if err != nil {
		return "", err
	}

	fullMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(response.Message), response.Message)
	hash := crypto.Keccak256Hash([]byte(fullMessage))

	// Sign the hash
	signature, err := crypto.Sign(hash.Bytes(), privateKeyECDSA)
	if err != nil {
		return "", err
	}

	// Add 27 to the last byte to make the signature valid (It is needed, but I don't know why)
	signature[64] += 27

	time := response.ExpiresOn.Format(time.RFC3339)
	time = time[:len(time)-1] + "." + strconv.Itoa(int(response.ExpiresOn.UnixMilli()%1000)) + "Z"
	payload = "{\"message\": \"" + response.Message + "\", \"walletAddress\": \"" + response.WalletAddress + "\", \"expiresOn\": \"" + time + "\", \"hmac\": \"" + response.Hmac + "\", \"signature\": \"" + hexutil.Encode(signature) + "\"}"
	response, err = requests.SendBlurPostRequest("auth/login", payload, blurKey, "", "")

	if err != nil {
		return "", err
	}

	return response.AccessToken, nil
}

func UpdateAuthKeys(wallets structs.Wallets, amountToUpdate int, prioritizeOfferWallet bool, blurKey string) structs.Wallets {
	if prioritizeOfferWallet {
		authToken, err := generateAuthToken(wallets.OfferWallet, blurKey)
		if err != nil {
			misc.PrintRed("Error generating authToken for "+wallets.OfferWallet.Address+": "+err.Error(), true)
		} else {
			wallets.OfferWallet.BlurAuthToken = authToken
			misc.PrintGreen("Succesfully updated authToken for "+wallets.OfferWallet.Address, true)
		}
		amountToUpdate--
	}

	for i := 0; i != len(wallets.Wallets); i++ {
		if wallets.Wallets[i].BlurAuthToken == "" {
			if wallets.Wallets[i].Address == wallets.OfferWallet.Address && wallets.OfferWallet.BlurAuthToken != "" {
				wallets.Wallets[i].BlurAuthToken = wallets.OfferWallet.BlurAuthToken
				continue
			}

			if amountToUpdate == 0 {
				break
			}

			authToken, err := generateAuthToken(wallets.Wallets[i], blurKey)
			if err != nil {
				misc.PrintRed("Error generating authToken for "+wallets.Wallets[i].Address+": "+err.Error(), true)
			} else {
				wallets.Wallets[i].BlurAuthToken = authToken
				misc.PrintGreen("Succesfully updated authToken for "+wallets.Wallets[i].Address, true)
			}
			amountToUpdate--

			if wallets.Wallets[i].Address == wallets.OfferWallet.Address {
				wallets.OfferWallet.BlurAuthToken = wallets.Wallets[i].BlurAuthToken
			}

		}
	}

	if wallets.OfferWallet.BlurAuthToken == "" && amountToUpdate > 0 {
		authToken, err := generateAuthToken(wallets.OfferWallet, blurKey)
		if err != nil {
			misc.PrintRed("Error generating authToken for "+wallets.OfferWallet.Address+": "+err.Error(), true)
		} else {
			wallets.OfferWallet.BlurAuthToken = authToken
			misc.PrintGreen("Succesfully updated authToken for "+wallets.OfferWallet.Address, true)
		}
	}

	return wallets
}

func CheckAuthTokens(wallets structs.Wallets, blurKey string) (structs.Wallets, int) {
	invalidAuths := 0
	offerWalletChecked := false
	for i := 0; i != len(wallets.Wallets); i++ {
		if wallets.Wallets[i].BlurAuthToken == "" {
			misc.PrintYellow(wallets.Wallets[i].Address+" has no Auth Token", true)
			invalidAuths++
		} else {
			valid, err := checkAuthToken(wallets.Wallets[i], blurKey)
			if err != nil {
				misc.PrintRed("Error checking auth: "+err.Error(), true)
			} else if !valid {
				misc.PrintYellow(wallets.Wallets[i].Address+" has no valid Auth Token", true)
				wallets.Wallets[i].BlurAuthToken = ""
				invalidAuths++
			} else {
				misc.PrintGreen(wallets.Wallets[i].Address+" has a valid Auth Token", true)
			}
		}

		if wallets.Wallets[i].Address == wallets.OfferWallet.Address {
			wallets.OfferWallet.BlurAuthToken = wallets.Wallets[i].BlurAuthToken
			offerWalletChecked = true
		}

	}

	if !offerWalletChecked {
		if wallets.OfferWallet.BlurAuthToken == "" {
			misc.PrintYellow(wallets.OfferWallet.Address+" has no Auth Token", true)
			invalidAuths++
		} else {
			valid, err := checkAuthToken(wallets.OfferWallet, blurKey)
			if err != nil {
				misc.PrintRed("Error checking auth: "+err.Error(), true)
			} else if !valid {
				misc.PrintYellow(wallets.OfferWallet.Address+" has no valid Auth Token", true)
				wallets.OfferWallet.BlurAuthToken = ""
				invalidAuths++
			} else {
				misc.PrintGreen(wallets.OfferWallet.Address+" has a valid Auth Token", true)
			}
		}
	}

	return wallets, invalidAuths
}

func checkAuthToken(wallet structs.Wallet, blurKey string) (bool, error) {
	_, err := requests.SendBlurGetRequest("v1/rewards", "leaderboard", blurKey, wallet.BlurAuthToken, wallet.Address)

	if err != nil {
		if err.Error() == "status code 401" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func GetBlurListingsSequential(IDs []int, contract string, blurKey string, listingChannel chan structs.Listing) {
	var cursor interface{} = nil

	for {

		listings, cursorNew, err := getBlurBatch(contract, blurKey, cursor, IDs)
		if err != nil {
			misc.PrintRed("Error getting Blur listings: "+err.Error(), true)
			break
		}
		// Add to channel
		for _, listing := range listings {
			listingChannel <- listing
		}

		cursor = cursorNew

		if cursor == nil {
			break
		}
	}

	listingChannel <- structs.Listing{Marketplace: "BLUR"}
}

func getBlurBatch(contract string, blurKey string, cursor interface{}, IDs []int) ([]structs.Listing, interface{}, error) {
	json, err := json.Marshal(cursor)
	if err != nil {
		return nil, nil, err
	}

	params := "{\"cursor\":" + string(json) + ",\"traits\":[],\"hasAsks\":true}"

	parsedBody, err := requests.SendBlurGetRequest("v1/collections", contract+"/tokens?filters="+url.QueryEscape(params), blurKey, "", "")

	if err != nil {
		return nil, cursor, err
	}

	var cursorNew interface{}

	if len(parsedBody.Tokens) == 100 {
		cursorNew = parsedBody.Tokens[99]
	}

	return parseBlurListings(parsedBody, IDs), cursorNew, nil
}

func parseBlurListings(blurResponse api_structs.BlurResponse, IDs []int) []structs.Listing {

	var listings []structs.Listing

	for i := 0; i < len(blurResponse.Tokens); i++ {
		if blurResponse.Tokens[i].Price.Unit == "ETH" && blurResponse.Tokens[i].Price.Marketplace == "BLUR" {
			token, err := strconv.Atoi(blurResponse.Tokens[i].TokenID)
			if err != nil {
				continue
			}
			if misc.ContainsInt(IDs, token) {

				priceEth, _ := new(big.Float).SetString(blurResponse.Tokens[i].Price.Amount)
				priceEth.Mul(priceEth, big.NewFloat(constants.WEI_TO_ETH))
				priceWei := new(big.Int)
				priceEth.Int(priceWei)

				listings = append(listings, structs.Listing{
					Price:       priceWei,
					Marketplace: "BLUR",
					Collection:  blurResponse.ContractAddress,
					Token:       blurResponse.Tokens[i].TokenID,
				})
			}
		}
	}
	return listings
}

func GetBlurParameters(listing structs.Listing, wallet structs.Wallet, blurKey string) (structs.BlurParameters, error) {
	domain := "v1/buy/" + listing.Collection
	valueFloat := new(big.Float).SetInt(listing.Price)
	value := new(big.Float).Quo(valueFloat, big.NewFloat(constants.WEI_TO_ETH))

	payload := "{\"tokenPrices\": [{\"tokenId\": \"" + listing.Token + "\",\"price\": {\"amount\": \"" + value.String() + "\",\"unit\": \"ETH\"}}],\"userAddress\": \"" + wallet.Address + "\"}"
	response, err := requests.SendBlurPostRequest(domain, payload, blurKey, wallet.BlurAuthToken, wallet.Address)

	if err != nil {
		return structs.BlurParameters{}, err
	}

	decryptedResponse, err := decryptBlurBuyResponse(response.Data)

	if decryptedResponse.CancelReasons != nil {
		return structs.BlurParameters{}, fmt.Errorf("%v", decryptedResponse.CancelReasons)
	}

	if err != nil {
		return structs.BlurParameters{}, err
	}

	return parseBlurParameters(decryptedResponse), nil
}

func decryptBlurBuyResponse(encryptedData string) (api_structs.BlurPostResponse, error) {
	b64Decoded, err := b64.StdEncoding.DecodeString(encryptedData)

	if err != nil {
		return api_structs.BlurPostResponse{}, err
	}

	key := "XTtnJ44LDXvZ1MSjdyK4pPT8kg5meJtHF44RdRBGrsaxS6MtG19ekKBxiXgp"

	var decrypted string

	for i := 0; i != len(b64Decoded); i++ {
		decrypted += string(b64Decoded[i] ^ key[i%len(key)])
	}

	var transaction api_structs.BlurPostResponse
	err = json.Unmarshal([]byte(decrypted), &transaction)
	return transaction, err
}

func parseBlurParameters(blurResponse api_structs.BlurPostResponse) structs.BlurParameters {
	return structs.BlurParameters{
		Data: blurResponse.Buys[0].TxnData.Data,
		To:   blurResponse.Buys[0].TxnData.To,
	}
}

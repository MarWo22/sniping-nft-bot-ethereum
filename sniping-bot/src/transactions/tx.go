package transactions

import (
	"NFT_Bot/src/abi_handler"
	"NFT_Bot/src/api"
	"NFT_Bot/src/api/api_structs"
	"NFT_Bot/src/constants"
	"NFT_Bot/src/misc"
	"NFT_Bot/src/structs"
	"NFT_Bot/src/webhooks"
	"math/big"
	"strconv"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// Gathers listings of rare tokens, and submits transactions to mainnet using the provided wallets
func FulfillAvailableListings(contract string, rarities structs.Rarities, buySettings structs.BuySettings, wallets []structs.Wallet, apiKeys structs.ApiKeys) {
	// Create listings channel for listings of the selected rarity
	listingsChan := initListingsChannels(contract, rarities.Ranks[:buySettings.BuyingRange], len(wallets), apiKeys)
	var claimedTokens []int

	var wg sync.WaitGroup

	// Start a worker for every wallet
	for i := 0; i != len(wallets); i++ {
		wg.Add(1)
		go startWalletWorker(&wg, listingsChan[i], wallets[i], rarities, buySettings, apiKeys, claimedTokens)
	}

	// Wait for every worker to finish
	wg.Wait()
	misc.PrintYellow("No available listings", true)
}

// Starts a worker for the wallet. It will listen for listings, check if it can afford the listing, claim the listing, and create and send a transaction for that listing
func startWalletWorker(wg *sync.WaitGroup, listingsChan chan structs.Listing, wallet structs.Wallet, rarities structs.Rarities, buySettings structs.BuySettings, apiKeys structs.ApiKeys, claimedTokens []int) {

	for {
		// Extract listing from the listings channel
		listing, closed := <-listingsChan

		// Breaks the loop if the listing channel is closed (All listings have been extracted)
		if listing == (structs.Listing{}) && !closed {
			break
		}
		// Extract ID
		ID, _ := strconv.Atoi(listing.Token)

		// Find the range object it belongs to
		buyRange := checkRange(listing, rarities, buySettings.Ranges)
		priceToPay := calculateFullPrice(listing, int64(buyRange.PriorityFee), int64(buySettings.MaxGas))
		// Check whether the token is within budget, there is enough funds,  and not claimed
		if isWithinBudget(listing, buyRange) && !misc.ContainsInt(claimedTokens, ID) && priceToPay.Cmp(wallet.Balance) <= 0 {
			// Make sure the wallet has an authToken if the listing is on blur
			if listing.Marketplace == "BLUR" && wallet.BlurAuthToken == "" {
				continue
			}
			// Claim the token to prevent double transactions over multiple wallets
			claimedTokens = append(claimedTokens, ID)
			// Fulfill the listing to mainnet
			fulfillListing(listing, wallet, apiKeys, rarities, buyRange, buySettings)
			// Update balance for next listing
			wallet.Balance, _ = api.GetBalance(wallet.Address, apiKeys.AlchemyKey)
		}

	}
	// Communicate the worker is finished
	defer wg.Done()
}

// Initializes the channel that receives listings from Blur and OpenSea
func initListingsChannels(contract string, IDs []int, nWallets int, apiKeys structs.ApiKeys) []chan structs.Listing {
	listingsChan := make(chan structs.Listing)
	// Send all Blur listings to the channel
	go api.GetBlurListingsSequential(IDs, contract, apiKeys.BlurKey, listingsChan)
	// Send all OpenSea listings to the channel
	go api.GetOpenSeaListingsSequential(IDs, contract, apiKeys.OpenSeaKey, listingsChan)

	// Duplicate the channels to the amount of wallets
	duplicatedChan := make([]chan structs.Listing, nWallets)
	for i := 0; i != nWallets; i++ {
		duplicatedChan[i] = make(chan structs.Listing, 50)
	}

	// Delegate all listings to the duplicated channels
	go delegateDuplicateListings(listingsChan, duplicatedChan)

	return duplicatedChan
}

// Listens for listings and delegates them to the duplicated channels. The duplicated channels should be buffered to prevent softlocks
func delegateDuplicateListings(listingsChan chan structs.Listing, duplicatedChan []chan structs.Listing) {
	emptyListings := 0
	for {
		// Extract listing from the main channel
		listing := <-listingsChan

		// Checks for empty structs (indicating no more available listings) and quits loop after both marketplaces are checked
		if listing == (structs.Listing{Marketplace: "OPENSEA"}) || listing == (structs.Listing{Marketplace: "BLUR"}) {
			emptyListings++
			if emptyListings == 2 {
				// Closes all channels
				for _, channel := range duplicatedChan {
					close(channel)
				}
				break
			}
			continue
		}
		// Sends listing to all duplicate channels
		for _, channel := range duplicatedChan {
			channel <- listing
		}
	}
}

// Creates a transaction from the given listing and submits it to mainnet. Will wait for the transaction to be mined, and send a webhook with the result.
// The balance is updated after the transaction
func fulfillListing(listing structs.Listing, wallet structs.Wallet, apiKeys structs.ApiKeys, rarities structs.Rarities, buyRange structs.Range, buySettings structs.BuySettings) {
	// Create the signed transaction
	signedTx, err := generateSignedTransaction(listing, buyRange.PriorityFee, buySettings.MaxGas, wallet, apiKeys)

	if err != nil {
		misc.PrintRed("Error generating signed transaction: "+err.Error(), true)
		return
	}
	// Ignore if no signedTx was created
	if signedTx == nil {
		return
	}
	// Submit the main transaction to mainnet
	txChan, err := submitTransaction(signedTx, wallet, apiKeys, listing, rarities, buyRange)
	if err != nil {
		misc.PrintRed("Error submitting signed transaction: "+err.Error(), true)
		return
	}

	// Gather receipt and result of transaction
	err = wrapUpTransaction(txChan, listing, rarities, apiKeys)
	if err != nil {
		misc.PrintRed("Error wrapping up transaction: "+err.Error(), true)
	}

	// Update balance
	wallet.Balance, err = api.GetBalance(wallet.Address, apiKeys.AlchemyKey)
	if err != nil {
		misc.PrintRed("Error updating wallet balance: "+err.Error(), true)
	}
}

// Wrap-up function for submitted transactions. Will wait for the receipt and extract whether it is successfully mined or failed. Sends discord webhooks with the results
func wrapUpTransaction(txChan chan api_structs.AlchemyWebsocketResponse, listing structs.Listing, rarities structs.Rarities, apiKeys structs.ApiKeys) error {
	// Wait for the transaction to be mined
	minedTx := <-txChan

	// Extract receipt from transaction
	receipt, err := api.GetReceipt(minedTx.Params.Result.Transaction.Hash, apiKeys.AlchemyKey)
	if err != nil {
		return err
	}
	ID, _ := strconv.Atoi(listing.Token)

	// Check whether transaction was succesfull
	if receipt.Status == "0x1" {
		misc.PrintGreen("Transaction for token "+strconv.Itoa(ID)+" with rank "+strconv.Itoa(rarities.Tokens[ID].Rank)+" was succesfully mined", true)
		err := webhooks.SendSuccesfullWebhook(rarities.Tokens[ID], receipt, ID, listing.Collection, listing.Price, apiKeys.DiscordWebhook)
		if err != nil {
			misc.PrintRed("Error sending webhook: "+err.Error(), true)
		}
	} else {
		misc.PrintRed("Transaction for token "+strconv.Itoa(ID)+" with rank "+strconv.Itoa(rarities.Tokens[ID].Rank)+" failed", true)
		err := webhooks.SendFailedWebhook(rarities.Tokens[ID], receipt, ID, listing.Collection, listing.Price, apiKeys.DiscordWebhook)
		if err != nil {
			misc.PrintRed("Error sending webhook: "+err.Error(), true)
		}
	}

	return nil
}

// Creates a signed transaction from the given listing. Supports both OpenSea and Blur
func generateSignedTransaction(listing structs.Listing, priorityFee int, maxBaseFee int, wallet structs.Wallet, apiKeys structs.ApiKeys) (*types.Transaction, error) {
	// Generate appropriate transaction for each marketplace
	if listing.Marketplace == "BLUR" {
		return fulfillBlurListing(listing, int64(priorityFee), int64(maxBaseFee), wallet, apiKeys)
	} else if listing.Marketplace == "OPENSEA" {
		return fulfillOpenSeaListing(listing, int64(priorityFee), int64(maxBaseFee), wallet, apiKeys)
	}
	return nil, nil
}

// Submits a transaction to the mainnet
func submitTransaction(signedTx *types.Transaction, wallet structs.Wallet, apiKeys structs.ApiKeys, listing structs.Listing, rarities structs.Rarities, buyRange structs.Range) (chan api_structs.AlchemyWebsocketResponse, error) {
	// Convert transaction to bytes
	data, err := signedTx.MarshalBinary()
	if err != nil {
		return nil, err
	}

	// Encode the transaction to hex string
	tx := hexutil.Encode(data)

	// Create channel to wait for the transaction to be mined
	txChan := api.ListenForTx(wallet.Address, signedTx.Hash().Hex(), apiKeys.AlchemyKey)
	// Send the transaction to the mainnet
	_, err = api.SendRawSignedTransaction(tx, apiKeys.AlchemyKey)
	if err != nil {
		return nil, err
	}

	ID, _ := strconv.Atoi(listing.Token)

	misc.PrintYellow("Submitted transaction for token "+strconv.Itoa(ID)+" with rank "+strconv.Itoa(rarities.Tokens[ID].Rank), true)

	err = webhooks.SendSentWebhook(rarities.Tokens[ID], wallet.Address, ID, listing.Collection, listing.Price, buyRange.PriorityFee, apiKeys.DiscordWebhook)
	if err != nil {
		misc.PrintRed("Error sending webhook: "+err.Error(), true)
	}

	return txChan, nil
}

// Creates the signed transaction required to fulfill an opensea listing
func fulfillOpenSeaListing(listing structs.Listing, priorityFee int64, maxBaseFee int64, wallet structs.Wallet, apiKeys structs.ApiKeys) (*types.Transaction, error) {
	// Get the fulfillment data from OpenSea
	fulfilmentData, err := api.GetListingFulfillmentData(listing.OrderHash, wallet.Address, apiKeys.OpenSeaKey)

	if err != nil {
		return nil, err
	}
	// Encode it to calldate using the OpenSea ABI
	calldata, err := abi_handler.EncodeOpenseaABI(fulfilmentData)

	if err != nil {
		return nil, err
	}

	// Create the signed transaction and return it
	return createSignedTransaction(wallet, listing.Price, priorityFee, maxBaseFee, apiKeys.AlchemyKey, calldata, constants.OPENSEA_CONTRACT)
}

// Creates the signed transaction required to fulfill a blur listing
func fulfillBlurListing(listing structs.Listing, priorityFee int64, maxBaseFee int64, wallet structs.Wallet, apiKeys structs.ApiKeys) (*types.Transaction, error) {
	// Get the parameters required to fulfill the listing from Blur
	parameters, err := api.GetBlurParameters(listing, wallet, apiKeys.BlurKey)

	if err != nil {
		return nil, err
	}
	// Decode the hex string
	parametersHex, err := hexutil.Decode(parameters.Data)

	if err != nil {
		return nil, err
	}

	// Create the signed transaction and return it
	return createSignedTransaction(wallet, listing.Price, priorityFee, maxBaseFee, apiKeys.AlchemyKey, parametersHex, parameters.To)
}

// Creates a transaction with the provided calldata, contract, and gas settings. The transaction is signed by the provided wallet.
func createSignedTransaction(wallet structs.Wallet, price *big.Int, priorityFee int64, maxBaseFee int64, alchemyKey string, calldata []byte, contractAddress string) (*types.Transaction, error) {
	// Gather the nonce
	nonce, err := api.GetNonce(wallet.Address, alchemyKey)
	if err != nil {
		return nil, err
	}

	// Parsing parameters to required format
	address := common.HexToAddress(contractAddress)

	// Create the DynamicFeeTx transaction struct
	txStruct := types.DynamicFeeTx{
		Nonce:     uint64(nonce),
		GasTipCap: big.NewInt(priorityFee * constants.WEI_TO_GWEI),
		GasFeeCap: big.NewInt((maxBaseFee + priorityFee) * constants.WEI_TO_GWEI),
		Gas:       constants.GAS,
		To:        &(address),
		Value:     price,
		Data:      calldata,
	}

	// Convert the private key to ECDSA
	privateKeyECDSA, err := crypto.HexToECDSA(wallet.PrivateKey)
	if err != nil {
		return nil, err
	}

	// Sign the transaction
	signedTx, err := types.SignNewTx(privateKeyECDSA, types.NewLondonSigner(big.NewInt(constants.CHAIN_ID)), &txStruct)

	if err != nil {
		return nil, err
	}

	// Convert the signed transaction to bytes
	return signedTx, nil
}

// Checks whether the token is within budget
func isWithinBudget(listing structs.Listing, buyRange structs.Range) bool {
	return listing.Price.Cmp(buyRange.Value) <= 0
}

// Finds the price range the token is part of
func checkRange(listing structs.Listing, rarities structs.Rarities, ranges []structs.Range) structs.Range {
	ID, _ := strconv.Atoi(listing.Token)
	rank := rarities.Tokens[ID].Rank
	for _, buyRange := range ranges {
		if rank >= buyRange.Low && rank <= buyRange.High {
			return buyRange
		}
	}
	return structs.Range{}
}

// Returns the balance needed to pay for the transaction including fees
func calculateFullPrice(listing structs.Listing, priorityFee int64, maxBaseFee int64) *big.Int {
	gasPrice := new(big.Int).Mul(big.NewInt(priorityFee+maxBaseFee), big.NewInt(constants.WEI_TO_GWEI))
	totalFee := new(big.Int).Mul(gasPrice, big.NewInt(constants.GAS))
	priceToPay := new(big.Int).Add(totalFee, listing.Price)
	return priceToPay
}

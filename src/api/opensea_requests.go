package api

import (
	"NFT_Bot/src/api/api_structs"
	"NFT_Bot/src/api/request_handler/requests"
	"NFT_Bot/src/constants"
	"NFT_Bot/src/misc"
	"NFT_Bot/src/structs"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
)

// Grabs a collection from opensea and returns it as a openseaCollection struct
func GetCollection(contract string, openSeaKey string) (api_structs.OpenSeaResponse, error) {

	parsedBody, err := requests.SendOpenSeaRequest("api/v2/chain/ethereum/contract", contract, openSeaKey)

	if err != nil {
		return api_structs.OpenSeaResponse{}, err
	}

	return requests.SendOpenSeaRequest("api/v2/collections/", parsedBody.Collection, openSeaKey)
}

func GetOpenSeaListingsSequential(IDs []int, contract string, openSeaKey string, listingChannel chan structs.Listing) {
	for idx := 0; idx != int(math.Ceil(float64(len(IDs))/30.0)); idx++ {
		var task []int
		if (idx+1)*30 > len(IDs) {
			task = IDs[idx*30:]
		} else {
			task = IDs[idx*30 : (idx+1)*30]
		}
		listings, err := getListingsBatch(task, contract, openSeaKey)
		if err != nil {
			misc.PrintRed(err.Error(), true)
			continue
		}

		for _, listing := range listings {
			listingChannel <- listing
		}
	}

	listingChannel <- structs.Listing{Marketplace: "OPENSEA"}
}

func CreateOffer(offer structs.OfferStruct, openSeaKey string) error {
	payload, err := json.Marshal(offer)
	if err != nil {
		return err
	}

	result, err := requests.SendOpenseaPostRequest("api/v2/orders/ethereum/seaport/offers", string(payload), openSeaKey)

	if err != nil {
		return err
	}

	if result.Error != nil {
		return fmt.Errorf("%v", result.Error)
	}

	// needs redoing
	return nil
}

// Executes a request of at most 30 ID's. Any more and the response will be a 400 status
func getListingsBatch(IDs []int, contract string, openSeaKey string) ([]structs.Listing, error) {
	params := "listings?asset_contract_address=" + contract + "&limit=50"

	for _, id := range IDs {
		params += "&token_ids=" + strconv.Itoa(id)
	}

	parsedBody, err := requests.SendOpenSeaRequest("v2/orders/ethereum/seaport", params, openSeaKey)

	if err != nil {
		misc.PrintRed("Error sending OpenSea request: "+err.Error(), true)
		return nil, err
	}

	return parseListings(parsedBody), nil
}

func GetListingFulfillmentData(hash string, address string, openSeaKey string) (api_structs.OpenseaPostResponse, error) {
	payload := "{\"listing\":{\"hash\":\"" + hash + "\",\"chain\":\"ethereum\",\"protocol_address\":\"" + constants.OPENSEA_CONTRACT + "\"},\"fulfiller\":{\"address\":\"" + address + "\"}}"

	return requests.SendOpenseaPostRequest("v2/listings/fulfillment_data", string(payload), openSeaKey)
}

func parseListings(openSeaResponse api_structs.OpenSeaResponse) []structs.Listing {
	var listings []structs.Listing
	var listedTokens []string

	for i := 0; i < len(openSeaResponse.Orders); i++ {
		token := openSeaResponse.Orders[i].ProtocolData.Parameters.Offer[0].IdentifierOrCriteria

		if !misc.ContainsString(listedTokens, token) {
			parameters := openSeaResponse.Orders[i].ProtocolData.Parameters
			price, _ := new(big.Int).SetString(parameters.Consideration[0].StartAmount, 10)

			for _, additionalRecipient := range parameters.Consideration[1:] {
				amount, _ := new(big.Int).SetString(additionalRecipient.StartAmount, 10)
				price.Add(price, amount)
			}

			listings = append(listings, structs.Listing{
				Price:       price,
				OrderHash:   openSeaResponse.Orders[i].OrderHash,
				Collection:  parameters.Offer[0].Token,
				Token:       parameters.Offer[0].IdentifierOrCriteria,
				Marketplace: "OPENSEA",
			})

			listedTokens = append(listedTokens, token)
		}
	}

	return listings
}

package abi_handler

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"strings"

	"NFT_Bot/src/api/api_structs"
	"NFT_Bot/src/structs"

	abi_eth "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func EncodeOpenseaABI(input api_structs.OpenseaPostResponse) ([]byte, error) {

	switch input.FulfillmentData.Transaction.Function {
	case "fulfillBasicOrder_efficient_6GL6yc((address,uint256,uint256,address,address,address,uint256,uint256,uint8,uint256,uint256,bytes32,uint256,bytes32,bytes32,uint256,(uint256,address)[],bytes))":
		params := initBasicOrderParameters(input)
		return abiReaderOpenseaPort.Pack("fulfillBasicOrder_efficient_6GL6yc", params)
	case "fulfillBasicOrder((address,uint256,uint256,address,address,address,uint256,uint256,uint8,uint256,uint256,bytes32,uint256,bytes32,bytes32,uint256,(uint256,address)[],bytes))":
		params := initBasicOrderParameters(input)
		return abiReaderOpenseaPort.Pack("fulfillBasicOrder", params)
	}
	return nil, errors.New("unsuported contract method")
}

func initBasicOrderParameters(input api_structs.OpenseaPostResponse) BasicOrderParameters {
	price := big.NewInt(input.FulfillmentData.Transaction.Value)

	parameters := input.FulfillmentData.Transaction.InputData.Parameters
	totalOriginalAdditionalRecipients, _ := new(big.Int).SetString(parameters.TotalOriginalAdditionalRecipients, 10)

	var additionalRecipients []AdditionalRecipients
	for _, additionalRecipient := range parameters.AdditionalRecipients {
		amount, _ := new(big.Int).SetString(additionalRecipient.Amount, 10)
		additionalRecipients = append(additionalRecipients, AdditionalRecipients{
			Recipient: common.HexToAddress(additionalRecipient.Recipient),
			Amount:    amount,
		})
		price.Add(price, amount)
	}

	var orderParameters BasicOrderParameters

	orderParameters.ConsiderationToken = common.HexToAddress(parameters.ConsiderationToken)
	orderParameters.ConsiderationIdentifier, _ = new(big.Int).SetString(parameters.ConsiderationIdentifier, 10)
	orderParameters.ConsiderationAmount, _ = new(big.Int).SetString(parameters.ConsiderationAmount, 10)
	orderParameters.Offerer = common.HexToAddress(parameters.Offerer)
	orderParameters.Zone = common.HexToAddress(parameters.Zone)
	orderParameters.OfferToken = common.HexToAddress(parameters.OfferToken)
	orderParameters.OfferIdentifier, _ = new(big.Int).SetString(parameters.OfferIdentifier, 10)
	orderParameters.OfferAmount, _ = new(big.Int).SetString(parameters.OfferAmount, 10)
	orderParameters.BasicOrderType = uint8(parameters.BasicOrderType)
	orderParameters.StartTime, _ = new(big.Int).SetString(parameters.StartTime, 10)
	orderParameters.EndTime, _ = new(big.Int).SetString(parameters.EndTime, 10)
	orderParameters.Salt, _ = new(big.Int).SetString(parameters.Salt, 10)
	orderParameters.TotalOriginalAdditionalRecipients = totalOriginalAdditionalRecipients
	orderParameters.AdditionalRecipients = additionalRecipients
	orderParameters.Signature, _ = hexutil.Decode(parameters.Signature)

	zoneHash, _ := hexutil.Decode(parameters.ZoneHash)
	offererConduitKey, _ := hexutil.Decode(parameters.OffererConduitKey)
	fulfillerConduitKey, _ := hexutil.Decode(parameters.FulfillerConduitKey)

	// Copy the values to ZoneHash, OffereConduitKey and FulfillerConduitKye since they use 32byte{} arrays
	copy(orderParameters.ZoneHash[:], zoneHash[:])
	copy(orderParameters.OffererConduitKey[:], offererConduitKey[:])
	copy(orderParameters.FulfillerConduitKey[:], fulfillerConduitKey[:])

	return orderParameters
}

// Decodes an input string into an array of TXParameters.
// Can decode the following:
// FulfillBasicOrder inputs (ID: 0xfb0f3ee1), returns 1 TXParameter
// fulfillAvailableOrders inputs (ID: 0xed98a574), returns multiple TXParameters
// Function will fail if presented with a different method ID
// On success, it returns an array of TXParameters
// On failure, it returns an error
func DecodeAbi(params string) ([]structs.Parameters, error) {

	bytes, err := hex.DecodeString(params[10 : len(params)-(len(params)-10)%32])
	if err != nil {
		return nil, err
	}
	unpackedTx := make(map[string]interface{})

	if params[:10] == "0xfb0f3ee1" {
		// Unpack fulfillBasicOrder method
		err = abiReaderOpenseaPort.UnpackIntoMap(unpackedTx, "fulfillBasicOrder", bytes)
		if err != nil {
			return nil, err
		}

		return basicOrderToParameters(unpackedTx)
	} else if params[:10] == "0xed98a574" {
		// Unpack fulfillAvailableOrders method
		err := abiReaderOpenseaPort.UnpackIntoMap(unpackedTx, "fulfillAvailableOrders", bytes)
		if err != nil {
			return nil, err
		}
		return availableOrdersToParameters(unpackedTx)
	}

	return nil, errors.New("transaction uses unsupported method")
}

func DecodeBaseUriUpdate(txData string, abi api_structs.Abi) (map[string]interface{}, error) {
	abiString, err := json.Marshal(abi)
	if err != nil {
		return nil, err
	}

	baseUriAbiReader, err := abi_eth.JSON(strings.NewReader(string(abiString)))
	if err != nil {
		return nil, err
	}

	m, ok := baseUriAbiReader.Methods[abi[0].Name]
	if !ok {
		return nil, errors.New("error finding method")
	}
	bytes, err := hex.DecodeString(txData[10 : len(txData)-(len(txData)-10)%32])
	unpacked := make(map[string]interface{})

	if err := m.Inputs.UnpackIntoMap(unpacked, bytes); err != nil {
		return nil, err
	}
	return unpacked, err
}

// Converts an ABI decoded interface map of the fulfillBasicOrder entry
// to an array of TXParameters
// On success, it returns a TXParamaters array with 1 entry
// On failure, it returns an error
func basicOrderToParameters(unpackedTx map[string]interface{}) ([]structs.Parameters, error) {

	newVar, ok := unpackedTx["parameters"].(struct {
		ConsiderationToken                common.Address "json:\"considerationToken\""
		ConsiderationIdentifier           *big.Int       "json:\"considerationIdentifier\""
		ConsiderationAmount               *big.Int       "json:\"considerationAmount\""
		Offerer                           common.Address "json:\"offerer\""
		Zone                              common.Address "json:\"zone\""
		OfferToken                        common.Address "json:\"offerToken\""
		OfferIdentifier                   *big.Int       "json:\"offerIdentifier\""
		OfferAmount                       *big.Int       "json:\"offerAmount\""
		BasicOrderType                    uint8          "json:\"basicOrderType\""
		StartTime                         *big.Int       "json:\"startTime\""
		EndTime                           *big.Int       "json:\"endTime\""
		ZoneHash                          [32]uint8      "json:\"zoneHash\""
		Salt                              *big.Int       "json:\"salt\""
		OffererConduitKey                 [32]uint8      "json:\"offererConduitKey\""
		FulfillerConduitKey               [32]uint8      "json:\"fulfillerConduitKey\""
		TotalOriginalAdditionalRecipients *big.Int       "json:\"totalOriginalAdditionalRecipients\""
		AdditionalRecipients              []struct {
			Amount    *big.Int       "json:\"amount\""
			Recipient common.Address "json:\"recipient\""
		} "json:\"additionalRecipients\""
		Signature []uint8 "json:\"signature\""
	})

	if !ok {
		return nil, errors.New("error asserting basicOrder's type")
	}

	listing := structs.Parameters{
		ConsiderationToken:                newVar.ConsiderationToken.Hex(),
		ConsiderationIdentifier:           newVar.ConsiderationIdentifier,
		ConsiderationAmount:               newVar.ConsiderationAmount,
		Offerer:                           newVar.Offerer.Hex(),
		Zone:                              newVar.Zone.Hex(),
		OfferToken:                        newVar.OfferToken.Hex(),
		OfferIdentifier:                   newVar.OfferIdentifier,
		OfferAmount:                       newVar.OfferAmount,
		BasicOrderType:                    uint8(newVar.BasicOrderType),
		StartTime:                         newVar.StartTime,
		EndTime:                           newVar.EndTime,
		ZoneHash:                          string(newVar.ZoneHash[:]),
		Salt:                              newVar.Salt,
		OffererConduitKey:                 string(newVar.OffererConduitKey[:]),
		FulfillerConduitKey:               string(newVar.FulfillerConduitKey[:]),
		TotalOriginalAdditionalRecipients: newVar.TotalOriginalAdditionalRecipients,
		Signature:                         hex.EncodeToString(newVar.Signature),
	}

	return []structs.Parameters{listing}, nil

}

// Converts an ABI decoded interface map of the fulfillAvailableOrders method
// to an array of TXParameters
// On success, it returns a TXParamaters array with multiple entry
// On failure, it returns an error
func availableOrdersToParameters(unpackedTx map[string]interface{}) ([]structs.Parameters, error) {
	orders, ok := unpackedTx["orders"].([]struct {
		Parameters struct {
			Offerer common.Address "json:\"offerer\""
			Zone    common.Address "json:\"zone\""
			Offer   []struct {
				ItemType             uint8          "json:\"itemType\""
				Token                common.Address "json:\"token\""
				IdentifierOrCriteria *big.Int       "json:\"identifierOrCriteria\""
				StartAmount          *big.Int       "json:\"startAmount\""
				EndAmount            *big.Int       "json:\"endAmount\""
			} "json:\"offer\""
			Consideration []struct {
				ItemType             uint8          "json:\"itemType\""
				Token                common.Address "json:\"token\""
				IdentifierOrCriteria *big.Int       "json:\"identifierOrCriteria\""
				StartAmount          *big.Int       "json:\"startAmount\""
				EndAmount            *big.Int       "json:\"endAmount\""
				Recipient            common.Address "json:\"recipient\""
			} "json:\"consideration\""
			OrderType                       uint8     "json:\"orderType\""
			StartTime                       *big.Int  "json:\"startTime\""
			EndTime                         *big.Int  "json:\"endTime\""
			ZoneHash                        [32]uint8 "json:\"zoneHash\""
			Salt                            *big.Int  "json:\"salt\""
			ConduitKey                      [32]uint8 "json:\"conduitKey\""
			TotalOriginalConsiderationItems *big.Int  "json:\"totalOriginalConsiderationItems\""
		} "json:\"parameters\""
		Signature []uint8 "json:\"signature\""
	})

	if !ok {
		return nil, errors.New("error asserting availableOrders's type")
	}

	returnParameters := make([]structs.Parameters, len(orders))

	for idx := 0; idx != len(orders); idx++ {
		parameters := orders[idx].Parameters

		returnParameters[idx] = structs.Parameters{
			ConsiderationToken:                parameters.Consideration[0].Token.Hex(),
			ConsiderationIdentifier:           parameters.Consideration[0].IdentifierOrCriteria,
			ConsiderationAmount:               parameters.Consideration[0].StartAmount,
			Offerer:                           parameters.Offerer.Hex(),
			Zone:                              parameters.Zone.Hex(),
			OfferToken:                        parameters.Offer[0].Token.Hex(),
			OfferIdentifier:                   parameters.Offer[0].IdentifierOrCriteria,
			OfferAmount:                       big.NewInt(int64(len(parameters.Offer))),
			BasicOrderType:                    uint8(parameters.OrderType),
			StartTime:                         parameters.StartTime,
			EndTime:                           parameters.EndTime,
			ZoneHash:                          hex.EncodeToString(parameters.ZoneHash[:]),
			Salt:                              parameters.Salt,
			OffererConduitKey:                 hex.EncodeToString(parameters.ConduitKey[:]),
			FulfillerConduitKey:               hex.EncodeToString(parameters.ConduitKey[:]),
			TotalOriginalAdditionalRecipients: big.NewInt(int64(len(parameters.Consideration) - 1)),

			Signature: hex.EncodeToString(orders[idx].Signature),
		}
	}

	return returnParameters, nil
}

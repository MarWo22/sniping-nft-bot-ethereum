package api

import (
	"NFT_Bot/src/api/api_structs"
	"NFT_Bot/src/api/request_handler/requests"
	"encoding/json"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Returns the balance of the given address
func GetBalance(address string, apiKey string) (*big.Int, error) {
	result, err := requests.SendAlchemyRequest("eth_getBalance", "[\""+address+"\", \"latest\"]", apiKey)

	if err != nil {
		return nil, err
	}

	output, err := hexutil.DecodeBig(result.Result.(string))

	if err != nil {
		return nil, err
	}

	return output, nil
}

// Returns the balance of the given address
func GetNonce(address string, apiKey string) (int64, error) {
	result, err := requests.SendAlchemyRequest("eth_getTransactionCount", "[\""+address+"\"]", apiKey)

	if err != nil {
		return 0, err
	}

	output, err := strconv.ParseInt(result.Result.(string)[2:], 16, 64)

	if err != nil {
		return 0, err
	}

	return output, nil
}

func SendRawSignedTransaction(transaction string, apiKey string) (string, error) {
	result, err := requests.SendAlchemyRequest("eth_sendRawTransaction", "[\""+transaction+"\"]", apiKey)

	if err != nil {
		return "", err
	}

	return result.Result.(string), nil
}

func GetBlock(apiKey string) (int64, error) {
	result, err := requests.SendAlchemyRequest("eth_blockNumber", "", apiKey)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(result.Result.(string)[2:], 16, 64)
}

func GetReceipt(hash string, apiKey string) (api_structs.Receipt, error) {
	result, err := requests.SendAlchemyRequest("eth_getTransactionReceipt", "[\""+hash+"\"]", apiKey)

	if err != nil {
		return api_structs.Receipt{}, err
	}

	jsonString, _ := json.Marshal(result.Result.(map[string]interface{}))
	receipt := api_structs.Receipt{}
	json.Unmarshal(jsonString, &receipt)

	return receipt, nil
}

func GetWethBalance(address string, apiKey string) (*big.Int, error) {
	result, err := requests.SendAlchemyRequest("alchemy_getTokenBalances", "[\""+address+"\",[\"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2\"]]", apiKey)

	if err != nil {
		return nil, err
	}
	// Convert AlchemyResponse struct to map[string]interface{}
	outputMap := result.Result.(map[string]interface{})
	// Grab tokenBalances from map
	tokenBalances := outputMap["tokenBalances"].([]interface{})
	// Grab the balance of the first occurence (wETH)
	balance := tokenBalances[0].(map[string]interface{})["tokenBalance"]
	// Convert to *big.Int. All leading zeroes are trimmed
	output, err := hexutil.DecodeBig("0x" + strings.TrimLeft(balance.(string)[2:], "0"))

	if err != nil {
		return nil, err
	}

	return output, nil
}

func GetTransactionByHash(hash string, apiKey string) (api_structs.Transaction, error) {
	result, err := requests.SendAlchemyRequest("eth_getTransactionByHash", "[\""+hash+"\"]", apiKey)

	if err != nil {
		return api_structs.Transaction{}, err
	}

	jsonString, _ := json.Marshal(result.Result.(map[string]interface{}))
	tx := api_structs.Transaction{}
	json.Unmarshal(jsonString, &tx)
	return tx, nil
}

func GetBeaconProxyAddress(contract string, apiKey string) (string, error) {
	result, err := requests.SendAlchemyRequest("eth_getStorageAt", "[\""+contract+"\",\"0xa3f0ad74e5423aebfd80d3ef4346578335a9a72aeaee59ff6cb3582b35133d50\"]", apiKey)

	if err != nil {
		return "", err
	}

	return "0x" + strings.TrimLeft(result.Result.(string)[2:], "0"), nil
}

func GetLogicProxyAddress(contract string, apiKey string) (string, error) {
	result, err := requests.SendAlchemyRequest("eth_getStorageAt", "[\""+contract+"\",\"0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc\"]", apiKey)

	if err != nil {
		return "", err
	}

	return "0x" + strings.TrimLeft(result.Result.(string)[2:], "0"), nil
}

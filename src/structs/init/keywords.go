package init

import (
	"NFT_Bot/src/api"
	"NFT_Bot/src/api/api_structs"
	"NFT_Bot/src/constants"
	"NFT_Bot/src/misc"
	"encoding/json"
	"errors"
	"strings"

	"github.com/antzucaro/matchr"
)

type keywords struct {
	UpdateBaseUri    string
	Supply           string
	TokenURI         string
	UpdateBaseUriAbi api_structs.Abi
}

func getKeywords(settings settings) (keywords, error) {
	newContract, err := getAddress(settings.Contract, settings.AlchemyKey, settings.EtherscanKey)

	if err != nil {
		return keywords{}, err
	}

	abi, err := api.GetABI(newContract, settings.EtherscanKey)

	if err != nil {
		misc.PrintRed("Error getting ABI, switching to ABI settings in JSON!", true)
		if settings.SupplyFunction == "" || settings.BaseURIUpdateABI == "" || settings.BaseURIUpdateFunction == "" || settings.TokenURIFunction == "" {
			return keywords{}, errors.New("Keywords/ABI not set in settings")
		}
		var abiEntry api_structs.Abi
		err := json.Unmarshal([]byte(settings.BaseURIUpdateABI), &abiEntry)
		if err != nil {
			return keywords{}, err
		}
		return keywords{
			UpdateBaseUri:    settings.BaseURIUpdateFunction,
			Supply:           settings.SupplyFunction,
			TokenURI:         settings.TokenURIFunction,
			UpdateBaseUriAbi: abiEntry,
		}, nil
	} else {
		updateKeyword, abiEntry, err := getKeyword(abi, constants.UPDATE_BASE_URI_KEYWORDS[:], "nonpayable", true)
		if err != nil {
			return keywords{}, err
		}

		supplyKeyword, _, err := getKeyword(abi, constants.SUPPLY_KEYWORDS[:], "view", false)
		if err != nil {
			return keywords{}, err
		}

		uriKeyword, _, err := getKeyword(abi, constants.TOKEN_URI_KEYWORDS[:], "view", true)
		if err != nil {
			return keywords{}, err
		}
		return keywords{
			UpdateBaseUri:    updateKeyword,
			Supply:           supplyKeyword,
			TokenURI:         uriKeyword,
			UpdateBaseUriAbi: abiEntry,
		}, nil
	}
}

func getAddress(contract string, alchemyKey string, etherscanKey string) (string, error) {
	beaconProxy, err := api.GetBeaconProxyAddress(contract, alchemyKey)
	if err != nil {
		return "", err
	}

	if beaconProxy != "0x" {
		return api.GetImplementationContract(beaconProxy, etherscanKey)
	}

	logicProxy, err := api.GetLogicProxyAddress(contract, alchemyKey)

	if err != nil {
		return "", err
	}

	if logicProxy != "0x" {
		return logicProxy, nil
	}

	return contract, nil
}

func getKeyword(abi api_structs.Abi, keywords []string, stateMutability string, needsInput bool) (string, api_structs.Abi, error) {
	var maxSimilarity = 0.0
	var functionName string
	currentAbi := make(api_structs.Abi, 1)
	for _, entry := range abi {
		if entry.StateMutability == stateMutability && (!needsInput || len(entry.Inputs) > 0) {
			for _, comparable := range keywords {
				matchVal := matchr.Jaro(strings.ToLower(entry.Name), strings.ToLower(comparable))
				if matchVal > maxSimilarity {
					maxSimilarity = matchVal
					functionName = entry.Name + "("
					for _, input := range entry.Inputs {
						functionName += (input.(map[string]interface{})["type"].(string) + ",")
					}

					if len(entry.Inputs) > 0 {
						functionName = functionName[:len(functionName)-1]
					}

					functionName += ")"
					currentAbi[0] = entry
				}
			}

		}
	}
	return functionName, currentAbi, nil
}

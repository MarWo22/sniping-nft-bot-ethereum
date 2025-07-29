package api

import (
	"NFT_Bot/src/api/api_structs"
	"NFT_Bot/src/api/request_handler/requests"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

// Extracts the baseURI, appender and whether the URI is from IPFS
func parseTokenURI(uri string) api_structs.TokenURI {
	baseURI := uri
	appender := ""

	split := strings.Split(uri, "/")
	lastOccurence := split[len(split)-1]
	re := regexp.MustCompile(`\d+`)

	// Filters out IPFS's CIDv1 and CIDv2, and strings without digits
	if (len(lastOccurence) != 46 && len(lastOccurence) != 59) || !re.MatchString(lastOccurence) {
		// Finds all series of digits
		digitOccurences := re.FindAllString(lastOccurence, -1)
		// Makes sure only 1 series exists
		if len(digitOccurences) == 1 {
			// Split the last occurence on the digit series
			lastOccurenceSplit := strings.Split(lastOccurence, digitOccurences[0])
			// Add everything left of the split to baseURI, everything right to appender
			baseURI = baseURI[:len(baseURI)-len(lastOccurence)] + lastOccurenceSplit[0]
			appender = lastOccurenceSplit[1]
		}
	}

	// Checks if the baseURI uses ipfs
	ipfs := false
	if strings.Contains(uri, "ipfs") {
		ipfs = true
	}

	// If the baseURI uses ipfs, it is converted to the ipfsGateway
	if ipfs {
		ipfsSplit := strings.Split(baseURI, "/")
		length := 0
		// Loop over all sections in the splitted url
		for _, section := range ipfsSplit {
			// If any of the section meets the CID length requirement, return the CID
			if ipfs && len(section) >= 46 {
				baseURI = baseURI[length:]
				break
			}
			length += len(section) + 1
		}
	}

	return api_structs.TokenURI{
		BaseURI:  baseURI,
		Appender: appender,
		IsIPFS:   ipfs,
	}
}

// Returns the tokenURI of a collection
func GetTokenUri(contract string, function string, apiKey string) (api_structs.TokenURI, error) {
	keccak := crypto.Keccak256([]byte(function))
	data := "0x" + hex.EncodeToString(keccak)[:8] + "0000000000000000000000000000000000000000000000000000000000000001"
	parameters := "&to=" + contract + "&data=" + data
	parsedBody, err := requests.SendEtherscanRequest(apiKey, "proxy", "eth_call", parameters)

	if err != nil {
		return api_structs.TokenURI{}, err
	}

	// Length of data
	length, err := strconv.ParseInt(parsedBody.Result.(string)[66:130], 16, 64)

	if err != nil {
		return api_structs.TokenURI{}, err
	}

	parsedTokenURI, err := hex.DecodeString(parsedBody.Result.(string)[130 : 130+length*2])

	if err != nil {
		return api_structs.TokenURI{}, err
	}
	return parseTokenURI(string(parsedTokenURI)), nil
}

// Returns the Total Supply of a collection
func GetSupply(contract string, function string, apiKey string) (int, error) {
	keccak := crypto.Keccak256([]byte(function))
	data := "0x" + hex.EncodeToString(keccak)[:8]
	parameters := "&to=" + contract + "&data=" + data
	parsedBody, err := requests.SendEtherscanRequest(apiKey, "proxy", "eth_call", parameters)

	if err != nil || parsedBody.Status == "0" {
		return 0, errors.New("Error requesting suppply: " + parsedBody.Result.(string))
	}

	supply, err := strconv.ParseInt(parsedBody.Result.(string)[2:], 16, 64)
	if err != nil {
		return 0, errors.New("Error requesting suppply: " + err.Error())
	}

	return int(supply), nil
}

// Returns the owner from a collection
func GetOwner(contract string, function string, apiKey string) (string, error) {
	keccak := crypto.Keccak256([]byte(function))
	data := "0x" + hex.EncodeToString(keccak)[:8]
	parameters := "&to=" + contract + "&data=" + data
	parsedBody, err := requests.SendEtherscanRequest(apiKey, "proxy", "eth_call", parameters)

	if err != nil {
		return "", errors.New("Error requesting owner: " + err.Error())
	}

	return "0x" + parsedBody.Result.(string)[26:], nil
}

// Gets the offset from a collection
func GetOffset(contract string, function string, apiKey string) (int, error) {
	keccak := crypto.Keccak256([]byte(function))
	data := "0x" + hex.EncodeToString(keccak)[:8]
	parameters := "&to=" + contract + "&data=" + data
	parsedBody, err := requests.SendEtherscanRequest(apiKey, "proxy", "eth_call", parameters)

	if err != nil {
		return 0, errors.New("Error requesting offset: " + err.Error())
	}

	offset, err := strconv.ParseInt(parsedBody.Result.(string)[2:], 16, 64)
	if err != nil {
		return 0, errors.New("Error parsing offset: " + err.Error())
	}
	return int(offset), nil
}

// Returns the ABI of a contract
func GetABI(contract string, apiKey string) (api_structs.Abi, error) {
	data := "&address=" + contract
	parsedBody, err := requests.SendEtherscanRequest(apiKey, "contract", "getabi", data)

	if err != nil {
		return nil, err
	}

	var abiVar api_structs.Abi
	if err := json.Unmarshal([]byte(parsedBody.Result.(string)), &abiVar); err != nil {
		return nil, err
	}
	return abiVar, nil
}

// Returns true if token with index 0 exists
func IncludesZero(contract string, function string, apiKey string) (bool, error) {
	keccak := crypto.Keccak256([]byte(function))
	data := "0x" + hex.EncodeToString(keccak)[:8] + "0000000000000000000000000000000000000000000000000000000000000000"
	parameters := "&to=" + contract + "&data=" + data
	_, err := requests.SendEtherscanRequest(apiKey, "proxy", "eth_call", parameters)

	if err != nil {
		if strings.Contains(err.Error(), "execution reverted") {
			return false, nil
		}
		return false, errors.New("Error requesting zero inclusion: " + err.Error())
	}

	return true, nil
}

func GetTransferEvents(contract string, address string, block int, apiKey string) ([]api_structs.TransferEvent, error) {
	data := "&contractaddress=" + contract + "&address=" + address + "&startblock=" + strconv.Itoa(block) + "&sort=asc&apikey=" + apiKey

	parsedBody, err := requests.SendEtherscanRequest(apiKey, "account", "tokennfttx", data)

	if err != nil {
		return nil, err
	}
	var transfers []api_structs.TransferEvent
	jsonString, err := json.Marshal(parsedBody.Result.([]interface{}))

	if err != nil {
		return nil, err
	}

	json.Unmarshal(jsonString, &transfers)
	return transfers, nil
}

func GetImplementationContract(contract string, apiKey string) (string, error) {
	keccak := crypto.Keccak256([]byte("implementation()"))
	data := "0x" + hex.EncodeToString(keccak)[:8]
	parameters := "&to=" + contract + "&data=" + data
	parsedBody, err := requests.SendEtherscanRequest(apiKey, "proxy", "eth_call", parameters)

	if err != nil {
		return "", errors.New("Error requesting implementation contract: " + err.Error())
	}

	return "0x" + parsedBody.Result.(string)[26:], nil
}

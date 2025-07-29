package init

import (
	"NFT_Bot/src/api"
	"NFT_Bot/src/misc"
)

type collection struct {
	Slug         string //
	Name         string //
	Owner        string //
	OpenSeaFee   int    //
	Supply       int
	TokenURI     string //
	Appender     string //
	IsIpfs       bool   //
	Offset       int    //
	IncludesZero bool   //
	ImageURL     string
	Fees         []struct {
		Fee       float64
		Recipient string
		Required  bool
	}
}

/*
TODO handling errors
*/
func InitCollection(settings settings, keywords keywords) collection {
	misc.PrintYellow("Pulling OpenSea collection data", true)
	openSeaResponse, err := api.GetCollection(settings.Contract, settings.OpenSeaKey)
	if err != nil {
		misc.PrintRed("Error pulling OpenSea collection data: "+err.Error(), true)
		misc.ExitProgram()
	} else {
		misc.PrintGreen("Successfully pulled OpenSea collection data", true)
	}
	/*
		misc.PrintYellow("Pulling owner", true)
		owner, err := api.GetOwner(settings.Contract, keyword, settings.EtherscanKey)
		if err != nil {
			misc.PrintRed("Error pulling owner: "+err.Error(), true)
			misc.ExitProgram()
		} else {
			misc.PrintGreen("Successfully pulled owner", true)
		}
	*/

	misc.PrintYellow("Pulling tokenURI", true)
	tokenURI, err := api.GetTokenUri(settings.Contract, keywords.TokenURI, settings.EtherscanKey)
	if err != nil {
		misc.PrintRed("Error pulling tokenURI: "+err.Error(), true)
		misc.ExitProgram()
	} else {
		misc.PrintGreen("Successfully pulled tokenURI", true)
	}

	misc.PrintYellow("Pulling supply", true)
	supply, err := api.GetSupply(settings.Contract, keywords.Supply, settings.EtherscanKey)
	if err != nil {
		misc.PrintRed("Error pulling supply: "+err.Error(), true)
		misc.ExitProgram()
	} else {
		misc.PrintGreen("Successfully pulled supply", true)
	}

	misc.PrintYellow("Pulling starting token", true)
	includesZero, err := api.IncludesZero(settings.Contract, keywords.TokenURI, settings.EtherscanKey)
	if err != nil {
		misc.PrintRed("Error pulling starting token: "+err.Error(), true)
		misc.ExitProgram()
	} else {
		misc.PrintGreen("Successfully pulled starting token", true)
	}

	return collection{
		Slug:         openSeaResponse.Collection,
		Name:         openSeaResponse.Name,
		Owner:        "",
		Supply:       supply,
		TokenURI:     tokenURI.BaseURI,
		Appender:     tokenURI.Appender,
		IsIpfs:       tokenURI.IsIPFS,
		Offset:       0,
		IncludesZero: includesZero,
		ImageURL:     openSeaResponse.ImageURL,
		Fees: []struct {
			Fee       float64
			Recipient string
			Required  bool
		}(openSeaResponse.Fees),
	}
}

package init

import (
	"NFT_Bot/src/structs"
	"encoding/json"
	"os"
)

type rangeStruct struct {
	Low         int     `json:"low"`
	High        int     `json:"high"`
	Value       float64 `json:"value"`
	PriorityFee int     `json:"priority_fee"`
}

type walletStruct struct {
	PrivateKey    string `json:"private_key"`
	BlurAuthToken string `json:"blur_auth_token"`
}

type wallets struct {
	WalletPrivateKeys     []walletStruct `json:"wallet_private_keys"`
	OfferWalletPrivateKey walletStruct   `json:"offer_wallet_private_key"`
}

type settings struct {
	Contract              string                 `json:"collection"`
	EncryptedContract     string                 `json:"encrypted_contract"`
	EncryptionKey         string                 `json:"encryption_key"`
	DiscordWebhook        string                 `json:"discord_webhook"`
	BlurKey               string                 `json:"blur_api"`
	MaxGas                int                    `json:"max_gas"`
	BuyingRange           int                    `json:"buying_range"`
	Ranges                []rangeStruct          `json:"ranges"`
	IPFSGateways          []string               `json:"ipfs_gateway"`
	BaseURIUpdateFunction string                 `json:"base_uri_update_keyword"`
	BaseURIUpdateABI      string                 `json:"base_uri_update_abi"`
	SupplyFunction        string                 `json:"supply_keyword"`
	TokenURIFunction      string                 `json:"tokenURI_keyword"`
	OffsetFunction        string                 `json:"offset"`
	UsesOffset            bool                   `json:"uses_offset"`
	MonitorOffset         bool                   `json:"monitor_offset"`
	Increment             int                    `json:"increment"`
	OpenSeaKey            string                 `json:"opensea_api"`
	AlchemyKey            string                 `json:"alchemy_mainnet_api"`
	EtherscanKey          string                 `json:"etherscan_api"`
	Servers               []structs.ServerObject `json:"servers"`
	OfferDuration         int                    `json:"offer_duration"`
}

func WriteWalletsJson(walletsStruct structs.Wallets) error {
	walletJson := wallets{WalletPrivateKeys: make([]walletStruct, len(walletsStruct.Wallets))}

	for i := 0; i != len(walletsStruct.Wallets); i++ {
		walletJson.WalletPrivateKeys[0].BlurAuthToken = walletsStruct.Wallets[0].BlurAuthToken
		walletJson.WalletPrivateKeys[0].PrivateKey = walletsStruct.Wallets[0].PrivateKey
	}

	walletJson.OfferWalletPrivateKey.BlurAuthToken = walletsStruct.OfferWallet.BlurAuthToken
	walletJson.OfferWalletPrivateKey.PrivateKey = walletsStruct.OfferWallet.PrivateKey

	file, err := json.MarshalIndent(walletJson, "", "\t")
	if err != nil {
		return err
	}

	err = os.WriteFile("wallets.json", file, 0644)

	if err != nil {
		return err
	}
	return nil
}

func readWalletsJson() (wallets, error) {
	file, err := os.ReadFile("wallets.json")

	if err != nil {
		return wallets{}, err
	}

	var walletsVar wallets
	err = json.Unmarshal(file, &walletsVar)

	if err != nil {
		return wallets{}, err
	}
	return walletsVar, nil
}

func readSettingsJson() (settings, error) {
	file, err := os.ReadFile("settings.json")

	if err != nil {
		return settings{}, err
	}

	var settingsVar settings
	err = json.Unmarshal(file, &settingsVar)

	if err != nil {
		return settings{}, err
	}
	return settingsVar, nil
}

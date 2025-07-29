package structs

import (
	"NFT_Bot/src/api/api_structs"
	"math/big"
)

// Main data struct

type TaskData struct {
	ContractFunctions ContractFunctions
	TokenURI          TokenURI
	Collection        Collection
	Wallets           Wallets
	ApiKeys           ApiKeys
	BuySettings       BuySettings
	Contract          string
	MonitorOffset     bool
	Servers           []ServerObject
}

// Structs used by main struct

type ContractFunctions struct {
	SetBaseUriFunction  string
	TotalSupplyFunction string
	TokenURIFunction    string
	OffsetFunction      string
	SetBaseUriAbi       api_structs.Abi
}

type TokenURI struct {
	BaseURI      string
	Appender     string
	IsIpfs       bool
	Offset       int
	Increment    int
	IncludesZero bool
	IPFSGateways []string
	UsesOffset   bool
	Supply       int
}

type Collection struct {
	Slug       string
	Name       string
	Owner      string
	OpenSeaFee int
	ImageURL   string
	Fees       []struct {
		Fee       float64
		Recipient string
		Required  bool
	}
}

type Wallets struct {
	Wallets     []Wallet
	OfferWallet Wallet
}

type ApiKeys struct {
	OpenSeaKey     string `json:"opensea_api"`
	AlchemyKey     string `json:"alchemy_mainnet_api"`
	EtherscanKey   string `json:"etherscan_api"`
	DiscordWebhook string
	BlurKey        string
}

type BuySettings struct {
	MaxGas            int
	BuyingRange       int
	OfferDuration     int
	EncryptedContract string
	EncryptionKey     string
	Ranges            []Range
}

// Sub structs

type Wallet struct {
	PrivateKey    string
	BlurAuthToken string
	Address       string
	Balance       *big.Int
}

type Range struct {
	Low         int      `json:"low"`
	High        int      `json:"high"`
	Value       *big.Int `json:"value"`
	PriorityFee int      `json:"priority_fee"`
}

type ServerObject struct {
	IP             string `json:"ip"`
	Port           int    `json:"port"`
	Timeout        int    `json:"timeout"`
	MaxIpfsTasks   int    `json:"max_tasks_ipfs"`
	MaxCustomTasks int    `json:"max_tasks_custom"`
}

// Token/rarity structs

type Attributes struct {
	TraitType string      `json:"trait_type"`
	Value     interface{} `json:"value"`
}

type Token struct {
	Traits []Attributes `json:"attributes"`
	Image  string       `json:"image"`
	Name   string       `json:"name"`
	Rarity float64
	Rank   int
}

type Listing struct {
	Price       *big.Int
	PendingGwei int
	Token       string
	Collection  string
	OrderHash   string
	Marketplace string
}

type Rarities struct {
	Ranks  []int
	Tokens map[int]Token
}

// Listing node struct

type ListingNode struct {
	Request  chan []int
	Response chan []Listing
}

type WebsocketTask struct {
	TaskLimit int    `json:"taskLimit"`
	Timeout   int    `json:"timeout"`
	BaseURI   string `json:"baseURI"`
	Appender  string `json:"appender"`
	IDs       []int  `json:"tasks"`
}

// Transaction structs

// All parameters required to fulfill a OpenSea order
type Parameters struct {
	ConsiderationIdentifier           *big.Int
	OfferAmount                       *big.Int
	BasicOrderType                    uint8
	TotalOriginalAdditionalRecipients *big.Int
	OfferIdentifier                   *big.Int
	StartTime                         *big.Int
	EndTime                           *big.Int
	ConsiderationToken                string
	Offerer                           string
	Zone                              string
	OfferToken                        string
	ConsiderationAmount               *big.Int
	ZoneHash                          string
	Salt                              *big.Int
	OffererConduitKey                 string
	FulfillerConduitKey               string
	Signature                         string
	AdditionalRecipients              []AdditionalRecipient
}

type AdditionalRecipient struct {
	Recipient string
	Amount    *big.Int
}

// Discord webhook structs

type DiscordLayout struct {
	Content     interface{}   `json:"content"`
	Embeds      []Embed       `json:"embeds"`
	AvatarURL   string        `json:"avatar_url"`
	Attachments []interface{} `json:"attachments"`
}

type Embed struct {
	Title     string    `json:"title"`
	Color     int       `json:"color"`
	URL       string    `json:"url,omitempty"`
	Fields    []Field   `json:"fields"`
	Footer    Footer    `json:"footer"`
	Thumbnail Thumbnail `json:"thumbnail"`
	Timestamp string    `json:"timestamp"`
}

type Footer struct {
	Text string `json:"text"`
}

type Thumbnail struct {
	URL string `json:"url"`
}

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

// Offer structs

type OfferStruct struct {
	Parameters      OfferParameters `json:"parameters"`
	ProtocolAddress string          `json:"protocol_address"`
	Signature       string          `json:"signature"`
}
type Offer struct {
	ItemType             int    `json:"itemType"`
	Token                string `json:"token"`
	IdentifierOrCriteria string `json:"identifierOrCriteria"`
	StartAmount          string `json:"startAmount"`
	EndAmount            string `json:"endAmount"`
}
type Consideration struct {
	ItemType             int    `json:"itemType"`
	Token                string `json:"token"`
	IdentifierOrCriteria string `json:"identifierOrCriteria"`
	StartAmount          string `json:"startAmount"`
	EndAmount            string `json:"endAmount"`
	Recipient            string `json:"recipient"`
}
type OfferParameters struct {
	Offerer                         string          `json:"offerer"`
	Offer                           []Offer         `json:"offer"`
	Consideration                   []Consideration `json:"consideration"`
	StartTime                       int64           `json:"startTime"`
	EndTime                         int64           `json:"endTime"`
	OrderType                       int             `json:"orderType"`
	Zone                            string          `json:"zone"`
	ZoneHash                        string          `json:"zoneHash"`
	Salt                            string          `json:"salt"`
	ConduitKey                      string          `json:"conduitKey"`
	TotalOriginalConsiderationItems string          `json:"totalOriginalConsiderationItems"`
	Counter                         int             `json:"counter"`
}

type BlurParameters struct {
	Data string `json:"data"`
	To   string `json:"to"`
}

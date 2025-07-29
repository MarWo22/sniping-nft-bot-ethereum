package api_structs

import "time"

type OpenSeaResponse struct {
	Name       string  `json:"name"`
	ImageURL   string  `json:"image_url"`
	Collection string  `json:"collection"`
	Orders     []Order `json:"orders"`
	Success    bool    `json:"success"`
	Fees       []struct {
		Fee       float64 `json:"fee"`
		Recipient string  `json:"recipient"`
		Required  bool    `json:"required"`
	} `json:"fees"`
}

type Order struct {
	ProtocolData struct {
		Parameters struct {
			Offerer string `json:"offerer"`
			Offer   []struct {
				ItemType             int    `json:"itemType"`
				Token                string `json:"token"`
				IdentifierOrCriteria string `json:"identifierOrCriteria"`
				StartAmount          string `json:"startAmount"`
			} `json:"offer"`
			Consideration []struct {
				Token                string `json:"token"`
				IdentifierOrCriteria string `json:"identifierOrCriteria"`
				StartAmount          string `json:"startAmount"`
				Recipient            string `json:"recipient"`
			} `json:"consideration"`
			StartTime                       string `json:"startTime"`
			EndTime                         string `json:"endTime"`
			OrderType                       int    `json:"orderType"`
			Zone                            string `json:"zone"`
			ZoneHash                        string `json:"zoneHash"`
			Salt                            string `json:"salt"`
			ConduitKey                      string `json:"conduitKey"`
			TotalOriginalConsiderationItems int    `json:"totalOriginalConsiderationItems"`
		} `json:"parameters"`
		Signature string `json:"signature"`
	} `json:"protocol_data"`
	CurrentPrice string `json:"current_price"`
	OrderHash    string `json:"order_hash"`
}

type AlchemyResponse struct {
	JsonRPC string       `json:"jsonrpc"`
	Result  interface{}  `json:"result"`
	Error   ErrorAlchemy `json:"error"`
}

type Receipt struct {
	BlockNumber       string `json:"blockNumber"`
	EffectiveGasPrice string `json:"effectiveGasPrice"`
	From              string `json:"from"`
	GasUsed           string `json:"gasUsed"`
	TransactionHash   string `json:"transactionHash"`
	Status            string `json:"status"`
}

type Transaction struct {
	BlockNumber string `json:"blockNumber"`
	Hash        string `json:"hash"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
	Input       string `json:"input"`
}

type ErrorAlchemy struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type EtherscanResponse struct {
	Jsonrpc string         `json:"jsonrpc"`
	ID      int            `json:"id"`
	Result  interface{}    `json:"result"`
	Status  string         `json:"status"`
	Message string         `json:"Message"`
	Error   EtherscanError `json:"error"`
}

type Abi []struct {
	Inputs    []interface{} `json:"inputs"`
	Name      string        `json:"name"`
	Type      string        `json:"type"`
	Anonymous bool          `json:"anonymous,omitempty"`
	Outputs   []struct {
		InternalType string `json:"internalType"`
		Name         string `json:"name"`
		Type         string `json:"type"`
	} `json:"outputs"`
	StateMutability string `json:"stateMutability,omitempty"`
}

type TokenURI struct {
	BaseURI  string
	Appender string
	IsIPFS   bool
}

type EtherscanError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

type AlchemyWebsocketResponse struct {
	Params struct {
		Result struct {
			From                 string `json:"from"`
			MaxPriorityFeePerGas string `json:"maxPriorityFeePerGas"`
			Hash                 string `json:"hash"`
			Input                string `json:"input"`
			To                   string `json:"to"`
			Value                string `json:"value"`
			Transaction          struct {
				Hash        string `json:"hash"`
				Input       string `json:"input"`
				Value       string `json:"value"`
				Type        string `json:"type"`
				BlockNumber string `json:"blockNumber"`
			} `json:"transaction"`
		} `json:"result"`
	} `json:"params"`
}

type TransferEvent struct {
	Hash    string `json:"hash"`
	TokenID string `json:"tokenID"`
	Input   string `json:"input"`
}

type OpenseaPostResponse struct {
	FulfillmentData struct {
		Transaction struct {
			Function  string `json:"function"`
			Chain     int    `json:"chain"`
			To        string `json:"to"`
			Value     int64  `json:"value"`
			InputData struct {
				Parameters struct {
					ConsiderationToken                string `json:"considerationToken"`
					ConsiderationIdentifier           string `json:"considerationIdentifier"`
					ConsiderationAmount               string `json:"considerationAmount"`
					Offerer                           string `json:"offerer"`
					Zone                              string `json:"zone"`
					OfferToken                        string `json:"offerToken"`
					OfferIdentifier                   string `json:"offerIdentifier"`
					OfferAmount                       string `json:"offerAmount"`
					BasicOrderType                    int    `json:"basicOrderType"`
					StartTime                         string `json:"startTime"`
					EndTime                           string `json:"endTime"`
					ZoneHash                          string `json:"zoneHash"`
					Salt                              string `json:"salt"`
					OffererConduitKey                 string `json:"offererConduitKey"`
					FulfillerConduitKey               string `json:"fulfillerConduitKey"`
					TotalOriginalAdditionalRecipients string `json:"totalOriginalAdditionalRecipients"`
					AdditionalRecipients              []struct {
						Amount    string `json:"amount"`
						Recipient string `json:"recipient"`
					} `json:"additionalRecipients"`
					Signature string `json:"signature"`
				} `json:"parameters"`
			} `json:"input_data"`
		} `json:"transaction"`
	} `json:"fulfillment_data"`
	Error interface{}
}

type BlurResponse struct {
	Success         bool        `json:"success"`
	ContractAddress string      `json:"contractAddress"`
	TotalCount      int         `json:"totalCount"`
	Tokens          []BlurToken `json:"tokens"`
}

type BlurToken struct {
	TokenID  string `json:"tokenId"`
	Name     string `json:"name"`
	ImageURL string `json:"imageUrl"`
	Traits   struct {
	} `json:"traits"`
	RarityScore interface{} `json:"rarityScore"`
	RarityRank  interface{} `json:"rarityRank"`
	Price       struct {
		Amount      string    `json:"amount"`
		Unit        string    `json:"unit"`
		ListedAt    time.Time `json:"listedAt"`
		Marketplace string    `json:"marketplace"`
	} `json:"price"`
	HighestBid interface{} `json:"highestBid"`
	LastSale   struct {
		Amount   string    `json:"amount"`
		Unit     string    `json:"unit"`
		ListedAt time.Time `json:"listedAt"`
	} `json:"lastSale"`
	LastCostBasis struct {
		Amount   string    `json:"amount"`
		Unit     string    `json:"unit"`
		ListedAt time.Time `json:"listedAt"`
	} `json:"lastCostBasis"`
	Owner struct {
		Address  string      `json:"address"`
		Username interface{} `json:"username"`
	} `json:"owner"`
	NumberOwnedByOwner int  `json:"numberOwnedByOwner"`
	IsSuspicious       bool `json:"isSuspicious"`
}

type BlurPostResponse struct {
	Message       string    `json:"message"`
	WalletAddress string    `json:"walletAddress"`
	ExpiresOn     time.Time `json:"expiresOn"`
	Hmac          string    `json:"hmac"`
	AccessToken   string    `json:"accessToken"`
	Data          string    `json:"data"`
	Buys          []struct {
		TxnData struct {
			Data  string `json:"data"`
			To    string `json:"to"`
			Value struct {
				Type string `json:"type"`
				Hex  string `json:"hex"`
			} `json:"value"`
		} `json:"txnData"`
	} `json:"buys"`
	CancelReasons []interface{} `json:"cancelReasons"`
}

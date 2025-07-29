package abi_handler

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type BasicOrderParameters struct {
	ConsiderationToken                common.Address         `abi:"considerationToken"`
	ConsiderationIdentifier           *big.Int               `abi:"considerationIdentifier"`
	ConsiderationAmount               *big.Int               `abi:"considerationAmount"`
	Offerer                           common.Address         `abi:"offerer"`
	Zone                              common.Address         `abi:"zone"`
	OfferToken                        common.Address         `abi:"offerToken"`
	OfferIdentifier                   *big.Int               `abi:"offerIdentifier"`
	OfferAmount                       *big.Int               `abi:"offerAmount"`
	BasicOrderType                    uint8                  `abi:"basicOrderType"`
	StartTime                         *big.Int               `abi:"startTime"`
	EndTime                           *big.Int               `abi:"endTime"`
	ZoneHash                          [32]byte               `abi:"zoneHash"`
	Salt                              *big.Int               `abi:"salt"`
	OffererConduitKey                 [32]byte               `abi:"offererConduitKey"`
	FulfillerConduitKey               [32]byte               `abi:"fulfillerConduitKey"`
	TotalOriginalAdditionalRecipients *big.Int               `abi:"totalOriginalAdditionalRecipients"`
	AdditionalRecipients              []AdditionalRecipients `abi:"additionalRecipients"`
	Signature                         []byte                 `abi:"signature"`
}

type AdditionalRecipients struct {
	Amount    *big.Int       `abi:"amount"`
	Recipient common.Address `abi:"recipient"`
}

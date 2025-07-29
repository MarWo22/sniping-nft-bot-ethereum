package webhooks

import (
	"NFT_Bot/src/api"
	"NFT_Bot/src/api/api_structs"
	"NFT_Bot/src/constants"
	"NFT_Bot/src/structs"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
)

func SendFailedWebhook(token structs.Token, receipt api_structs.Receipt, ID int, contract string, basePrice *big.Int, webhook string) error {

	block, err := strconv.ParseInt(receipt.BlockNumber[2:], 16, 64)
	if err != nil {
		return err
	}
	gasUsed, err := strconv.ParseInt(receipt.GasUsed[2:], 16, 64)
	if err != nil {
		return err
	}
	gasPrice, err := strconv.ParseInt(receipt.EffectiveGasPrice[2:], 16, 64)
	if err != nil {
		return err
	}

	valueFloat := new(big.Float).SetInt(basePrice)
	value := new(big.Float).Quo(valueFloat, big.NewFloat(constants.WEI_TO_ETH))

	fields := []structs.Field{
		{
			Name:   "Token",
			Value:  "[" + token.Name + "#" + strconv.Itoa(ID) + "](https://opensea.io/assets/ethereum/" + contract + "/" + strconv.Itoa(ID) + ")",
			Inline: true,
		},
		{
			Name:   "Rank",
			Value:  strconv.Itoa(token.Rank),
			Inline: true,
		},
		{
			Name:   "\u200b",
			Value:  "\u200b",
			Inline: true,
		},
		{
			Name:   "Price",
			Value:  fmt.Sprintf("%.4fΞ", value),
			Inline: true,
		},
		{
			Name:   "Gas Used",
			Value:  fmt.Sprintf("%.4fΞ", (float64(gasUsed) * float64(gasPrice) / constants.WEI_TO_ETH)),
			Inline: true,
		},
		{
			Name:   "\u200b",
			Value:  "\u200b",
			Inline: true,
		},
		{
			Name:   "Wallet Used",
			Value:  "||[" + receipt.From + "](https://etherscan.io/address/" + receipt.From + ")||",
			Inline: true,
		},
	}

	embeds := []structs.Embed{
		{
			Title:  "Failed Transaction In Block " + strconv.Itoa(int(block)),
			Color:  16711680,
			URL:    "https://etherscan.io/tx/" + receipt.TransactionHash,
			Fields: fields,
			Footer: structs.Footer{
				Text: "Sniping NFT bot v" + constants.VERSION,
			},
			Thumbnail: structs.Thumbnail{
				URL: verifyImage(token.Image),
			},
			Timestamp: time.Now().UTC().Format(time.RFC3339)[:19] + ".000Z",
		},
	}

	layout := structs.DiscordLayout{
		Content:     nil,
		Embeds:      embeds,
		AvatarURL:   "https://play-lh.googleusercontent.com/T_vA5l9W1-XYTmgr3gCB2MBd7QmA-iG0wcm09_IFWNB-4gOpnS-tYNEmcalwdixSyw",
		Attachments: nil,
	}

	return api.SendWebhook(layout, webhook)

}

func SendSuccesfullWebhook(token structs.Token, receipt api_structs.Receipt, ID int, contract string, basePrice *big.Int, webhook string) error {

	block, err := strconv.ParseInt(receipt.BlockNumber[2:], 16, 64)
	if err != nil {
		return err
	}
	gasUsed, err := strconv.ParseInt(receipt.GasUsed[2:], 16, 64)
	if err != nil {
		return err
	}
	gasPrice, err := strconv.ParseInt(receipt.EffectiveGasPrice[2:], 16, 64)
	if err != nil {
		return err
	}

	valueFloat := new(big.Float).SetInt(basePrice)
	value := new(big.Float).Quo(valueFloat, big.NewFloat(constants.WEI_TO_ETH))
	totalGasUsed := float64(gasUsed) * float64(gasPrice) / constants.WEI_TO_ETH

	fields := []structs.Field{
		{
			Name:   "Token",
			Value:  "[" + token.Name + "#" + strconv.Itoa(ID) + "](https://opensea.io/assets/ethereum/" + contract + "/" + strconv.Itoa(ID) + ")",
			Inline: true,
		},
		{
			Name:   "Rank",
			Value:  strconv.Itoa(token.Rank),
			Inline: true,
		},
		{
			Name:   "\u200b",
			Value:  "\u200b",
			Inline: true,
		},
		{
			Name:   "Price",
			Value:  fmt.Sprintf("%.4fΞ", value),
			Inline: true,
		},
		{
			Name:   "Gas Used",
			Value:  fmt.Sprintf("%.4fΞ", totalGasUsed),
			Inline: true,
		},
		{
			Name:   "Total ETH Paid",
			Value:  fmt.Sprintf("%.4fΞ", value.Add(new(big.Float).Copy(value), big.NewFloat(totalGasUsed))),
			Inline: true,
		},
		{
			Name:   "Total ETH paid",
			Value:  "||[" + receipt.From + "](https://etherscan.io/address/" + receipt.From + ")||",
			Inline: true,
		},
	}

	embeds := []structs.Embed{
		{
			Title:  "OpenSea Transaction Included In Block " + strconv.Itoa(int(block)),
			Color:  4062976,
			URL:    "https://etherscan.io/tx/" + receipt.TransactionHash,
			Fields: fields,
			Footer: structs.Footer{
				Text: "Sniping NFT bot v" + constants.VERSION,
			},
			Thumbnail: structs.Thumbnail{
				URL: verifyImage(token.Image),
			},
			Timestamp: time.Now().UTC().Format(time.RFC3339)[:19] + ".000Z",
		},
	}

	layout := structs.DiscordLayout{
		Content:     nil,
		Embeds:      embeds,
		AvatarURL:   "https://play-lh.googleusercontent.com/T_vA5l9W1-XYTmgr3gCB2MBd7QmA-iG0wcm09_IFWNB-4gOpnS-tYNEmcalwdixSyw",
		Attachments: nil,
	}

	return api.SendWebhook(layout, webhook)

}

func SendSentWebhook(token structs.Token, sender string, ID int, contract string, basePrice *big.Int, priorityFee int, webhook string) error {

	valueFloat := new(big.Float).SetInt(basePrice)
	value := new(big.Float).Quo(valueFloat, big.NewFloat(constants.WEI_TO_ETH))

	fields := []structs.Field{
		{
			Name:   "Token",
			Value:  "[" + token.Name + "#" + strconv.Itoa(ID) + "](https://opensea.io/assets/ethereum/" + contract + "/" + strconv.Itoa(ID) + ")",
			Inline: true,
		},
		{
			Name:   "Rank",
			Value:  strconv.Itoa(token.Rank),
			Inline: true,
		},
		{
			Name:   "\u200b",
			Value:  "\u200b",
			Inline: true,
		},
		{
			Name:   "Price",
			Value:  fmt.Sprintf("%.4fΞ", value),
			Inline: true,
		},
		{
			Name:   "Priority Fee",
			Value:  strconv.Itoa(priorityFee),
			Inline: true,
		},
		{
			Name:   "\u200b",
			Value:  "\u200b",
			Inline: true,
		},
		{
			Name:   "Wallet used",
			Value:  "||[" + sender + "](https://etherscan.io/address/" + sender + ")||",
			Inline: true,
		},
	}

	embeds := []structs.Embed{
		{
			Title:  "OpenSea Transaction Sent",
			Color:  16121600,
			Fields: fields,
			Footer: structs.Footer{
				Text: "Sniping NFT bot v" + constants.VERSION,
			},
			Thumbnail: structs.Thumbnail{
				URL: verifyImage(token.Image),
			},
			Timestamp: time.Now().UTC().Format(time.RFC3339)[:19] + ".000Z",
		},
	}

	layout := structs.DiscordLayout{
		Content:     nil,
		Embeds:      embeds,
		AvatarURL:   "https://play-lh.googleusercontent.com/T_vA5l9W1-XYTmgr3gCB2MBd7QmA-iG0wcm09_IFWNB-4gOpnS-tYNEmcalwdixSyw",
		Attachments: nil,
	}

	return api.SendWebhook(layout, webhook)
}

func SendRevealDetectedWebhook(collection structs.Collection, tokenURI structs.TokenURI, mode string, webhook string) error {

	gateway := ""
	if tokenURI.IsIpfs {
		gateway = constants.PUBLIC_IPFS_GATEWYA
	}

	fields := []structs.Field{
		{
			Name:   "Collection",
			Value:  "[" + collection.Name + "](https://opensea.io/collection/" + collection.Slug + ")",
			Inline: true,
		},
		{
			Name:   "Mode",
			Value:  mode,
			Inline: true,
		},
		{
			Name:   "\u200b",
			Value:  "\u200b",
			Inline: true,
		},
		{
			Name:   "TokenURI",
			Value:  "||" + gateway + tokenURI.BaseURI + "#" + tokenURI.Appender + "||",
			Inline: true,
		},
		{
			Name:   "Supply",
			Value:  strconv.Itoa(tokenURI.Supply),
			Inline: true,
		},
	}

	embeds := []structs.Embed{
		{
			Title:  "Reveal Detected!",
			Color:  16121600,
			Fields: fields,
			Footer: structs.Footer{
				Text: "Sniping NFT bot v" + constants.VERSION,
			},
			Thumbnail: structs.Thumbnail{
				URL: verifyImage(collection.ImageURL),
			},
			Timestamp: time.Now().UTC().Format(time.RFC3339)[:19] + ".000Z",
		},
	}

	layout := structs.DiscordLayout{
		Content:     nil,
		Embeds:      embeds,
		AvatarURL:   "https://play-lh.googleusercontent.com/T_vA5l9W1-XYTmgr3gCB2MBd7QmA-iG0wcm09_IFWNB-4gOpnS-tYNEmcalwdixSyw",
		Attachments: nil,
	}

	return api.SendWebhook(layout, webhook)
}

func SendOfferAcceptedWebhook(token structs.Token, sender string, hash string, ID int, contract string, basePrice *big.Int, webhook string) error {

	valueFloat := new(big.Float).SetInt(basePrice)
	value := new(big.Float).Quo(valueFloat, big.NewFloat(constants.WEI_TO_ETH))
	fields := []structs.Field{
		{
			Name:   "Token",
			Value:  "[" + token.Name + "#" + strconv.Itoa(ID) + "](https://opensea.io/assets/ethereum/" + contract + "/" + strconv.Itoa(ID) + ")",
			Inline: true,
		},
		{
			Name:   "Rank",
			Value:  strconv.Itoa(token.Rank),
			Inline: true,
		},
		{
			Name:   "\u200b",
			Value:  "\u200b",
			Inline: true,
		},
		{
			Name:   "Price",
			Value:  fmt.Sprintf("%.4fΞ", value),
			Inline: true,
		},
		{
			Name:   "\u200b",
			Value:  "\u200b",
			Inline: true,
		},
		{
			Name:   "Wallet used",
			Value:  "||[" + sender + "](https://etherscan.io/address/" + sender + ")||",
			Inline: true,
		},
	}

	embeds := []structs.Embed{
		{
			Title:  "OpenSea Offer Accepted",
			Color:  4062976,
			Fields: fields,
			URL:    "https://etherscan.io/tx/" + hash,
			Footer: structs.Footer{
				Text: "Sniping NFT bot v" + constants.VERSION,
			},
			Thumbnail: structs.Thumbnail{
				URL: verifyImage(token.Image),
			},
			Timestamp: time.Now().UTC().Format(time.RFC3339)[:19] + ".000Z",
		},
	}

	layout := structs.DiscordLayout{
		Content:     nil,
		Embeds:      embeds,
		AvatarURL:   "https://play-lh.googleusercontent.com/T_vA5l9W1-XYTmgr3gCB2MBd7QmA-iG0wcm09_IFWNB-4gOpnS-tYNEmcalwdixSyw",
		Attachments: nil,
	}

	return api.SendWebhook(layout, webhook)
}

func verifyImage(imageURL string) string {
	if imageURL[:7] == "ipfs://" {
		return constants.PUBLIC_IPFS_GATEWYA + imageURL[7:]
	}

	if idx := strings.LastIndex(imageURL, "ipfs"); idx != -1 {
		return constants.PUBLIC_IPFS_GATEWYA + imageURL[idx+5:]
	}

	if idx := strings.Index(imageURL, "w=500"); idx != -1 {
		if idx+5 != len(imageURL) {
			return imageURL[:idx] + "w=256" + imageURL[idx+5:]
		} else {
			return imageURL[:idx] + "w=256"
		}
	}

	return imageURL
}

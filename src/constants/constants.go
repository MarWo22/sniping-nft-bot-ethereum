package constants

const VERSION = "3.0"

const RARITIES_DIRECTORY = "rarities/"

const CHAIN_ID = 1
const GAS = 350000
const OPENSEA_CONTRACT = "0x00000000000000adc04c56bf30ac9d3c0aaf14dc"
const WEI_TO_GWEI = 1000000000
const WEI_TO_ETH = 1000000000000000000
const PUBLIC_IPFS_GATEWYA = "https://gateway.ipfs.io/ipfs/"

// offer consts
const WRAPPED_ETHER_ADDRESS = "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2"
const OFFER_FEE_ADDRESS = "0x0000a26b00c1F0DF003000390027140000fAa719"
const ZONE = "0x000000e7ec00e7b300774b00001314b8610022b8"
const ZONE_HASH = "0x3000000000000000000000000000000000000000000000000000000000000000"
const CONDUIT_KEY = "0x0000007b02230091a7ed01230072f7006a004d60a8d4e71d599b8104250f0000"

const MAX_MISSING_PERCENTAGE = 1

const ALCHEMY_BASE_HTTP = "https://eth-mainnet.g.alchemy.com/v2/-"
const ALCHEMY_BASE_WEBSOCKET = "wss://eth-mainnet.g.alchemy.com/v2/-"

// Not actually constant, but are used as if 'constant'. Golang doesn't support const arrays

var (
	SUPPLY_KEYWORDS          = [...]string{"supply, totalSupply", "original_supply"}
	TOKEN_URI_KEYWORDS       = [...]string{"tokenURI"}
	UPDATE_BASE_URI_KEYWORDS = [...]string{"setBaseURI", "setUriPrefix", "changeURI", "overrideVariables"}
)

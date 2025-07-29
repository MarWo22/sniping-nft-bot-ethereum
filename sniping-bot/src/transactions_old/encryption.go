package transactionsold

// LEAVING IT FOR NOW
import (
	"NFT_Bot/src/structs"
	crypto_rand "crypto/rand"
	"encoding/binary"
	"math/big"
	"math/rand"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

type encryptedParameters struct {
	ConsiderationIdentifier           [1]byte
	OfferAmount                       [1]byte
	BasicOrderType                    [1]byte
	TotalOriginalAdditionalRecipients [1]byte
	OfferIdentifier                   [2]byte
	StartTime                         [4]byte
	EndTime                           [4]byte
	ConsiderationToken                [20]byte
	Offerer                           [20]byte
	Zone                              [20]byte
	OfferToken                        [20]byte
	ConsiderationAmount               [32]byte
	ZoneHash                          [32]byte
	Salt                              [32]byte
	OffererConduitKey                 [32]byte
	FulfillerConduitKey               [32]byte
	Signature                         []byte
	AdditionalRecipients              []encryptedAdditionalRecipient
}

type encryptedAdditionalRecipient struct {
	Recipient [20]byte
	Amount    [32]byte
}

func getBigBytes(bigInt *big.Int) []byte {
	bytes := bigInt.Bytes()
	if len(bytes) == 0 {
		return []byte{0}
	}
	return bytes
}

func encryptParameters(raw structs.Parameters, key []byte) (encryptedParameters, error) {

	encryptedParams := encryptedParameters{}

	considerationToken, _ := hexutil.Decode(raw.ConsiderationToken)
	offerer, _ := hexutil.Decode(raw.Offerer)
	zone, _ := hexutil.Decode(raw.Zone)
	offerToken, _ := hexutil.Decode(raw.OfferToken)
	zoneHash, _ := hexutil.Decode(raw.ZoneHash)
	offererConduitKey, _ := hexutil.Decode(raw.OffererConduitKey)
	fulfillerConduitKey, _ := hexutil.Decode(raw.FulfillerConduitKey)
	signature, _ := hexutil.Decode(raw.Signature)

	copy(encryptedParams.ConsiderationIdentifier[:], encryptBytes(key, getBigBytes(raw.ConsiderationIdentifier)))
	copy(encryptedParams.OfferAmount[:], encryptBytes(key, getBigBytes(raw.OfferAmount)))
	// copy(encryptedParams.BasicOrderType[:], encryptBytes(key, getBigBytes(raw.BasicOrderType)))
	copy(encryptedParams.TotalOriginalAdditionalRecipients[:], encryptBytes(key, getBigBytes(raw.TotalOriginalAdditionalRecipients)))
	copy(encryptedParams.OfferIdentifier[:], encryptBytes(key, getBigBytes(raw.OfferIdentifier)))
	copy(encryptedParams.StartTime[:], encryptBytes(key, getBigBytes(raw.StartTime)))
	copy(encryptedParams.EndTime[:], encryptBytes(key, getBigBytes(raw.EndTime)))

	copy(encryptedParams.ConsiderationToken[:], encryptBytes(key, considerationToken))
	copy(encryptedParams.Offerer[:], encryptBytes(key, offerer))
	copy(encryptedParams.Zone[:], encryptBytes(key, zone))
	copy(encryptedParams.OfferToken[:], encryptBytes(key, offerToken))
	copy(encryptedParams.ConsiderationAmount[:], encryptBytes(key, pad32Left(getBigBytes(raw.ConsiderationAmount))))
	copy(encryptedParams.ZoneHash[:], encryptBytes(key, zoneHash))
	copy(encryptedParams.Salt[:], encryptBytes(key, pad32Left(getBigBytes(raw.Salt))))
	copy(encryptedParams.OffererConduitKey[:], encryptBytes(key, offererConduitKey))
	copy(encryptedParams.FulfillerConduitKey[:], encryptBytes(key, fulfillerConduitKey))

	encryptedParams.Signature = encryptBytes(key, signature)

	for i, additionalRecipient := range raw.AdditionalRecipients {
		recipient, err := hexutil.Decode(additionalRecipient.Recipient)
		if err != nil {
			return encryptedParameters{}, err
		}
		encryptedParams.AdditionalRecipients = append(encryptedParams.AdditionalRecipients, encryptedAdditionalRecipient{})
		copy(encryptedParams.AdditionalRecipients[i].Recipient[:], encryptBytes(key, recipient))
		copy(encryptedParams.AdditionalRecipients[i].Amount[:], encryptBytes(key, pad32Left(getBigBytes(additionalRecipient.Amount))))
	}

	return encryptedParams, nil
}

func encryptBytes(key []byte, data []byte) []byte {
	encryptedData := make([]byte, len(data))
	keyLen := len(key)
	for i := 0; i != len(data); i++ {
		encryptedData[i] = data[i] ^ key[i%keyLen]
	}
	return encryptedData
}

// RandStringBytesMaskImprSrc returns a random hexadecimal string of length n.
func generateKey() ([]byte, error) {

	var seed [8]byte
	_, err := crypto_rand.Read(seed[:])
	if err != nil {
		return nil, err
	}

	var src = rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(seed[:]))))
	key := make([]byte, 32) // can be simplified to n/2 if n is always even

	if _, err := src.Read(key); err != nil {
		return nil, err
	}

	return key, nil
}

func pad32Left(data []byte) []byte {
	ret := make([]byte, 32)
	if len(data) > 32 {
		return nil
	}
	copy(ret[32-len(data):32], data)
	return ret
}

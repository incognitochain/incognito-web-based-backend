package pblur

import (
	"crypto/ecdsa"
	"encoding/base64"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func EncodeData(data, key string) (string, error) {
	x, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	y := ""

	for i := 0; i < len(x); i++ {
		y += string(x[i] ^ key[i%len(key)])
	}
	fmt.Println("y:", y)
	return y, nil
}

func Sign(message, privateKeyString string) (string, error) {
	// Parse private key
	privateKey, err := crypto.HexToECDSA(privateKeyString)
	if err != nil {
		return "", nil
	}
	// Sign message
	signature, err := PersonalSign(message, privateKey)
	if err != nil {
		return "", nil
	}
	return signature, nil
}

// Returns a signature string
func PersonalSign(message string, privateKey *ecdsa.PrivateKey) (string, error) {
	signatureBytes, err := crypto.Sign(signHash([]byte(message)), privateKey)
	if err != nil {
		return "", err
	}
	signatureBytes[64] += 27
	return hexutil.Encode(signatureBytes), nil
}

func VerifySig(from, sigHex string, msg []byte) (bool, error) {

	fromAddr := common.HexToAddress(from)

	sig, err := hexutil.Decode(sigHex)
	if err != nil {
		return false, err
	}
	// https://github.com/ethereum/go-ethereum/blob/55599ee95d4151a2502465e0afc7c47bd1acba77/internal/ethapi/api.go#L442
	if sig[64] != 27 && sig[64] != 28 {
		return false, nil
	}
	sig[64] -= 27

	pubKey, err := crypto.SigToPub(signHash(msg), sig)
	if err != nil {
		return false, err
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	return fromAddr == recoveredAddr, nil
}

func signHash(data []byte) []byte {
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(data), data)
	return crypto.Keccak256([]byte(msg))
}

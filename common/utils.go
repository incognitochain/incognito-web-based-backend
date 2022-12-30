package common

import (
	"crypto/ecdsa"
	"errors"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

func GetNativeNetworkCurrencyType(network string) int {
	switch network {
	case "inc":
		return NativeCurrencyTypePRV
	case "eth":
		return NativeCurrencyTypeETH
	case "bsc":
		return NativeCurrencyTypeBSC
	case "plg":
		return NativeCurrencyTypePLG
	case "ftm":
		return NativeCurrencyTypeFTM
	case "avax":
		return NativeCurrencyTypeAVAX
	case "aurora":
		return NativeCurrencyTypeAURORA
	}
	return -1
}

func IsNativeCurrency(currencyType int) bool {
	switch currencyType {
	case NativeCurrencyTypeETH:
		return true
	case NativeCurrencyTypeBSC:
		return true
	case NativeCurrencyTypePLG:
		return true
	case NativeCurrencyTypeFTM:
		return true
	case NativeCurrencyTypeAVAX:
		return true
	case NativeCurrencyTypeAURORA:
		return true
	}
	return false
}

func GetNetworkID(network string) int {
	switch network {
	case NETWORK_INC:
		return NETWORK_INC_ID
	case NETWORK_ETH:
		return NETWORK_ETH_ID
	case NETWORK_BSC:
		return NETWORK_BSC_ID
	case NETWORK_PLG:
		return NETWORK_PLG_ID
	case NETWORK_FTM:
		return NETWORK_FTM_ID
	case NETWORK_AURORA:
		return NETWORK_AURORA_ID
	case NETWORK_AVAX:
		return NETWORK_AVAX_ID
	}
	return -1
}

func GetNetworkName(network int) string {
	switch network {
	case NETWORK_INC_ID:
		return NETWORK_INC
	case NETWORK_ETH_ID:
		return NETWORK_ETH
	case NETWORK_BSC_ID:
		return NETWORK_BSC
	case NETWORK_PLG_ID:
		return NETWORK_PLG
	case NETWORK_FTM_ID:
		return NETWORK_FTM
	case NETWORK_AURORA_ID:
		return NETWORK_AURORA
	case NETWORK_AVAX_ID:
		return NETWORK_AVAX
	}
	return ""
}

func GetNetworkIDFromCurrencyType(currencyType int) (int, error) {
	netID, ok := NetworkCurrencyMap[currencyType]
	if !ok {
		return 0, errors.New("unsupported network")
	}
	return netID, nil
}

func CheckIsWrappedNativeToken(contractAddress string, network int) bool {
	list, exist := WrappedNativeMap[network]
	if exist {
		for _, v := range list {
			if strings.EqualFold(contractAddress, v) {
				return true
			}
		}
	}
	return false
}

func GetEVMAddress(privateKey string) (string, error) {
	privKey, _ := crypto.HexToECDSA(privateKey)

	publicKey := privKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", errors.New("error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	return address, nil
}

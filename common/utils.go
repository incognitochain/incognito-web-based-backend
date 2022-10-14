package common

import (
	"errors"
	"strings"
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
	case "inc":
		return NETWORK_INC_ID
	case "eth":
		return NETWORK_ETH_ID
	case "bsc":
		return NETWORK_BSC_ID
	case "plg":
		return NETWORK_PLG_ID
	case "ftm":
		return NETWORK_FTM_ID
	case "avax":
		return NETWORK_AVAX_ID
	case "aurora":
		return NETWORK_AURORA_ID
	}
	return -1
}

func GetNetworkName(network int) string {
	switch network {
	case 0:
		return NETWORK_INC
	case 1:
		return NETWORK_ETH
	case 2:
		return NETWORK_BSC
	case 3:
		return NETWORK_PLG
	case 4:
		return NETWORK_FTM
	case 5:
		return NETWORK_AVAX
	case 6:
		return NETWORK_AURORA
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
			if strings.ToLower(contractAddress) == v {
				return true
			}
		}
	}
	return false
}

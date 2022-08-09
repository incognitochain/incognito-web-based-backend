package common

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
	}
	return -1
}

func GetNetworkID(network string) int {
	switch network {
	case "inc":
		return 0
	case "eth":
		return 1
	case "bsc":
		return 2
	case "plg":
		return 3
	case "ftm":
		return 4
	}
	return 0
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
	}
	return ""
}

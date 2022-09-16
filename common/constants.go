package common

import "strings"

var DefaultConfig = Config{
	Port:           9898,
	CoinserviceURL: "http://51.161.117.193:8096",
	NetworkID:      "testnet-2",
	FullnodeURL:    "https://testnet.incognito.org/fullnode",
	ShieldService:  "https://staging-api-service.incognito.org",
}

const (
	MODE_TXSUBMITWATCHER = "submitwatcher"
	MODE_TXSUBMITWORKER  = "submitworker"
	MODE_API             = "api"
	MODE_FEEESTIMATOR    = "feeestimator"
	MODE_UNSHIELDWATCHER = "unshieldwatcher"
)

const (
	EVMGasLimit = 600000
)

const (
	StatusSubmitting   = "submitting"
	StatusSubmitFailed = "submit_failed"
	StatusPending      = "pending"
	StatusExecuting    = "executing"
	StatusRejected     = "rejected"
	StatusAccepted     = "accepted"
	// StatusSubmittingOutchain    = "outchain_submitting"
	// StatusPendingOutchain       = "outchain_pending"
	// StatusSubmitOutchainFailed  = "outchain_submit_failed"
	// StatusSubmitOutchainSuccess = "outchain_submit_success"
)

const (
	NETWORK_INC = "inc"
	NETWORK_ETH = "eth"
	NETWORK_BSC = "bsc"
	NETWORK_PLG = "plg"
	NETWORK_FTM = "ftm"
)

const (
	NativeCurrencyTypePRV = 0
	NativeCurrencyTypeETH = 1
	NativeCurrencyTypeBSC = 7
	NativeCurrencyTypePLG = 19
	NativeCurrencyTypeFTM = 21
	UnifiedCurrencyType   = 25
)

const (
	NETWORK_INC_ID = iota
	NETWORK_ETH_ID
	NETWORK_BSC_ID
	NETWORK_PLG_ID
	NETWORK_FTM_ID
)

const (
	Unknown  = iota
	ETH      //1
	BTC      //2
	ERC20    //3
	BNB      //4
	BNB_BEP2 //5
	USD      //6

	BNB_BSC   //7
	BNB_BEP20 //8

	TOMO //9
	ZIL  //10
	XMR  //11
	NEO  //12
	DASH //13
	LTC  //14
	DOGE //15
	ZEC  //16
	DOT  //17
	PDEX //18 0000000000000000000000000000000000000000000000000000000000000006

	// Polygon:
	MATIC     //19
	PLG_ERC20 //20

	FTM       //21
	FTM_ERC20 //22

	SOL     //23
	SOL_SPL //24

	// pUnifined token:
	UNIFINE_TOKEN //25
)

var (
	NetworkCurrencyMap = map[int]int{
		UNIFINE_TOKEN: NETWORK_INC_ID,
		ETH:           NETWORK_ETH_ID,
		ERC20:         NETWORK_ETH_ID,
		BNB_BSC:       NETWORK_BSC_ID,
		BNB_BEP20:     NETWORK_BSC_ID,
		MATIC:         NETWORK_PLG_ID,
		PLG_ERC20:     NETWORK_PLG_ID,
		FTM:           NETWORK_FTM_ID,
		FTM_ERC20:     NETWORK_FTM_ID,
	}
)

var (
	WrappedNativeMap = map[int][]string{
		NETWORK_PLG_ID: {strings.ToLower("0x9c3C9283D3e44854697Cd22D3Faa240Cfb032889"), strings.ToLower("0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270")},
		NETWORK_BSC_ID: {strings.ToLower("0xbb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c"), strings.ToLower("0xae13d989dac2f0debff460ac112a837c89baa7cd")},
		NETWORK_ETH_ID: {strings.ToLower("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), strings.ToLower("0xb4fbf271143f4fbf7b91a5ded31805e42b2208d6")},
	}
)

const (
	PappTypeUnknown = iota
	PappTypeSwap
)

// Default param mainnet
var MainnetBridgeNetworkData = []BridgeNetworkData{
	{
		Network:            "eth",
		ChainID:            "1",
		ConfirmationBlocks: 35,
		Endpoints:          []string{"https://mainnet.infura.io/v3"},
	},
	{
		Network:            "plg",
		ChainID:            "137",
		ConfirmationBlocks: 128,
		Endpoints:          []string{"https://rpc.ankr.com/polygon"},
	},
	{
		Network:            "bsc",
		ChainID:            "56",
		ConfirmationBlocks: 14,
		Endpoints:          []string{"https://bsc-dataseed1.ninicoin.io"},
	},
}

var MainnetPappsEndpointData = []PAppsEndpointData{
	{
		Network:      "plg",
		ExchangeApps: map[string]string{"uniswap": "uniswapep:3000"},
		AppContracts: map[string]string{"uniswap": ""},
	},
	{
		Network:      "bsc",
		ExchangeApps: map[string]string{"pancake": "pancakeswapep:3000"},
		AppContracts: map[string]string{"pancake": ""},
	},
}

// Default param testnet
var TestnetBridgeNetworkData = []BridgeNetworkData{
	{
		Network:            "eth",
		ChainID:            "42",
		ConfirmationBlocks: 35,
		Endpoints:          []string{"https://goerli.infura.io/v3"},
	},
	{
		Network:            "plg",
		ChainID:            "80001",
		ConfirmationBlocks: 128,
		Endpoints:          []string{"https://matic-mumbai.chainstacklabs.com"},
	},
	{
		Network:            "bsc",
		ChainID:            "97",
		ConfirmationBlocks: 14,
		Endpoints: []string{
			"https://data-seed-prebsc-1-s1.binance.org:8545",
			"https://data-seed-prebsc-2-s2.binance.org:8545",
			"https://data-seed-prebsc-1-s1.binance.org:8545",
		},
	},
}

var TestnetPappsEndpointData = []PAppsEndpointData{
	{
		Network:      "plg",
		ExchangeApps: map[string]string{"uniswap": "uniswapep:3000"},
		AppContracts: map[string]string{"uniswap": "0xAe85BB3D2ED209736E4d236DcE24624EA1A04249"},
	},
	{
		Network:      "bsc",
		ExchangeApps: map[string]string{"pancake": "pancakeswapep:3000"},
		AppContracts: map[string]string{"pancake": "0x0e2923c21E2C5A2BDD18aa460B3FdDDDaDb0aE18"},
	},
}

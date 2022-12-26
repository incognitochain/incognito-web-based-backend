package common

import "strings"

var DefaultConfig = Config{
	Port:           9898,
	CoinserviceURL: "http://51.161.117.193:8096",
	NetworkID:      "testnet",
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
	EVMGasLimit        = 1500000
	EVMGasLimitETH     = 800000
	EVMGasLimitPancake = 600000
	MinEVMTxs          = uint64(10)
)

const (
	PercentFeeDiff = float64(5)
)

const (
	StatusSubmitting   = "submitting"
	StatusSubmitFailed = "submit_failed"
	StatusPending      = "pending"
	StatusExecuting    = "executing"
	StatusRejected     = "rejected"
	StatusAccepted     = "accepted"
	StatusWaiting      = "waiting"
	// StatusSubmittingOutchain    = "outchain_submitting"
	// StatusPendingOutchain       = "outchain_pending"
	// StatusSubmitOutchainFailed  = "outchain_submit_failed"
	// StatusSubmitOutchainSuccess = "outchain_submit_success"
)

const (
	StatusPdaOutchainTxFailed     = "pdao_outchain_failed"
	StatusPdaOutchainTxSubmitting = "pdao_outchain_submitting"
	StatusPdaOutchainTxPending    = "pdao_outchain_pending"
	StatusPdaOutchainTxSuccess    = "pdao_outchain_success"
)

const (
	NETWORK_INC    = "inc"
	NETWORK_ETH    = "eth"
	NETWORK_BSC    = "bsc"
	NETWORK_PLG    = "plg"
	NETWORK_FTM    = "ftm"
	NETWORK_AVAX   = "avax"
	NETWORK_AURORA = "aurora"
	NETWORK_NEAR   = "near"
)

const (
	NativeCurrencyTypePRV    = 0
	NativeCurrencyTypeETH    = 1
	NativeCurrencyTypeBSC    = 7
	NativeCurrencyTypePLG    = 19
	NativeCurrencyTypeFTM    = 21
	NativeCurrencyTypeAVAX   = 28
	NativeCurrencyTypeAURORA = 30
	NativeCurrencyTypeNEAR   = 26
	UnifiedCurrencyType      = 25
)

const (
	NETWORK_INC_ID = iota
	NETWORK_ETH_ID
	NETWORK_BSC_ID
	NETWORK_PLG_ID
	NETWORK_FTM_ID
	NETWORK_AURORA_ID
	NETWORK_AVAX_ID
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

	NEAR       //26
	NEAR_TOKEN //27

	AVAX       //28
	AVAX_ERC20 //29

	AURORA_ETH   //30
	AURORA_ERC20 //31
)

var (
	NetworkCurrencyMap = map[int]int{
		Unknown:       NETWORK_INC_ID,
		UNIFINE_TOKEN: NETWORK_INC_ID,
		ETH:           NETWORK_ETH_ID,
		ERC20:         NETWORK_ETH_ID,
		BNB_BSC:       NETWORK_BSC_ID,
		BNB_BEP20:     NETWORK_BSC_ID,
		MATIC:         NETWORK_PLG_ID,
		PLG_ERC20:     NETWORK_PLG_ID,
		FTM:           NETWORK_FTM_ID,
		FTM_ERC20:     NETWORK_FTM_ID,
		AVAX:          NETWORK_AVAX_ID,
		AVAX_ERC20:    NETWORK_AVAX_ID,
		AURORA_ETH:    NETWORK_AURORA_ID,
		AURORA_ERC20:  NETWORK_AURORA_ID,
	}
)

var (
	WrappedNativeMap = map[int][]string{
		NETWORK_PLG_ID:    {strings.ToLower("0x9c3C9283D3e44854697Cd22D3Faa240Cfb032889"), strings.ToLower("0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270")},
		NETWORK_BSC_ID:    {strings.ToLower("0xbb4cdb9cbd36b01bd1cbaebf2de08d9173bc095c"), strings.ToLower("0xae13d989dac2f0debff460ac112a837c89baa7cd")},
		NETWORK_ETH_ID:    {strings.ToLower("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"), strings.ToLower("0xb4fbf271143f4fbf7b91a5ded31805e42b2208d6")},
		NETWORK_FTM_ID:    {strings.ToLower("0xf1277d1Ed8AD466beddF92ef448A132661956621"), strings.ToLower("0x21be370D5312f44cB42ce377BC9b8a0cEF1A4C83")},
		NETWORK_AVAX_ID:   {strings.ToLower("0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7"), strings.ToLower("0xd00ae08403b9bbb9124bb305c09058e32c39a48c")},
		NETWORK_AURORA_ID: {strings.ToLower("0xC9BdeEd33CD01541e1eeD10f90519d2C06Fe3feB"), strings.ToLower("0x1b6A3d5B5DCdF7a37CFE35CeBC0C4bD28eA7e946")},
	}
)

var (
	NativeTokenSymbol = map[int]string{
		NETWORK_PLG_ID:    "MATIC",
		NETWORK_BSC_ID:    "BNB",
		NETWORK_ETH_ID:    "ETH",
		NETWORK_FTM_ID:    "FTM",
		NETWORK_AVAX_ID:   "AVAX",
		NETWORK_AURORA_ID: "AURORA",
	}
)

const (
	TestnetPortalV4BTCID = "4584d5e9b2fc0337dfb17f4b5bb025e5b82c38cfa4f54e8a3d4fcdd03954ff82"
	MainnetPortalV4BTCID = "b832e5d3b1f01a4f0623f7fe91d6673461e1f5d37d91fe78c5c2e6183ff39696"
)
const (
	ExternalTxTypeUnknown = iota
	ExternalTxTypeSwap
	ExternalTxTypeUnshield
	ExternalTxTypePdaoProposal = 69
	ExternalTxTypePdaoVote     = 70
	ExternalTxTypePdaoCancel   = 71
)

// Default param mainnet
var MainnetBridgeNetworkData = []BridgeNetworkData{
	{
		Network:            "eth",
		ChainID:            "1",
		ConfirmationBlocks: 35,
		Endpoints:          []string{"https://eth-fullnode.incognito.org"},
	},
	{
		Network:            "plg",
		ChainID:            "137",
		ConfirmationBlocks: 128,
		Endpoints:          []string{"https://polygon-mainnet.infura.io/v3/9bc873177cf74a03a35739e45755a9ac", "https://rpc-mainnet.maticvigil.com", "https://rpc-mainnet.matic.quiknode.pro"},
	},
	{
		Network:            "bsc",
		ChainID:            "56",
		ConfirmationBlocks: 14,
		Endpoints:          []string{"https://bsc-dataseed1.binance.org", "https://bsc-dataseed1.ninicoin.io"},
	},
}

var MainnetPappsEndpointData = []PAppsEndpointData{
	{
		Network:      "plg",
		ExchangeApps: map[string]string{"uniswap": "uniswapep:3000", "curve": ""},
		AppContracts: map[string]string{"uniswap": "0xCC8c88e9Dae72fa07aC077933a2E73d146FECdf0", "curve": "0x55b08b7c1ecdc1931660b18fe2d46ce7b20613e2"},
	},
	{
		Network:      "bsc",
		ExchangeApps: map[string]string{"pancake": "pancakeswapep:3000"},
		AppContracts: map[string]string{"pancake": "0x95Cd8898917c7216Da0517aAB6A115d7A7b6CA90"},
	},
	{
		Network:      "eth",
		ExchangeApps: map[string]string{"uniswap": "uniswapep_eth:3000"},
		AppContracts: map[string]string{"uniswap": "0xe38e54B2d6B1FCdfaAe8B674bF36ca62429fdBDe"},
	},
}

var MainnetIncognitoVault = []PappVaultData{
	{
		Network:         "bsc",
		Type:            1,
		ContractAddress: "0x43D037A562099A4C2c95b1E2120cc43054450629",
	},
	{
		Network:         "eth",
		Type:            1,
		ContractAddress: "0x43D037A562099A4C2c95b1E2120cc43054450629",
	},
	{
		Network:         "plg",
		Type:            1,
		ContractAddress: "0x43D037A562099A4C2c95b1E2120cc43054450629",
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
	{
		Network:            "ftm",
		ChainID:            "4002",
		ConfirmationBlocks: 5,
		Endpoints:          []string{"https://rpc.testnet.fantom.network"},
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
	{
		Network:      "ftm",
		ExchangeApps: map[string]string{"spooky": "spookyswapep:3000"},
		AppContracts: map[string]string{"spooky": "0x14D0cf3bC307aA15DA40Aa4c8cc2A2a81eF96B3a"},
	},
}

var TestnetIncognitoVault = []PappVaultData{
	{
		Network:         "bsc",
		Type:            1,
		ContractAddress: "0x3534C0a523b3A862c06C8CAF61de230f9b408f51",
	},
	{
		Network:         "eth",
		Type:            1,
		ContractAddress: "0xc157CC3077ddfa425bae12d2F3002668971A4e3d",
	},
	{
		Network:         "plg",
		Type:            1,
		ContractAddress: "0x76318093c374e39B260120EBFCe6aBF7f75c8D28",
	},
	{
		Network:         "ftm",
		Type:            1,
		ContractAddress: "0x76318093c374e39B260120EBFCe6aBF7f75c8D28",
	},
}

const (
	PRV_TOKEN            = "0000000000000000000000000000000000000000000000000000000000000004"
	ETH_UT_TOKEN_MAINNET = "3ee31eba6376fc16cadb52c8765f20b6ebff92c0b1c5ab5fc78c8c25703bb19e"
	ETH_UT_TOKEN_TESTNET = "b366fa400c36e6bbcf24ac3e99c90406ddc64346ab0b7ba21e159b83d938812d"
)

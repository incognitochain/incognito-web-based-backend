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

// const (
// 	BurnForCallConfirmMeta      = 158
// 	BurnForCallRequestMeta      = 348
// 	BurnForCallResponseMeta     = 349
// 	IssuingReshieldResponseMeta = 350
// )

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
	}
)

const (
	PappTypeUnknown = iota
	PappTypeSwap
)

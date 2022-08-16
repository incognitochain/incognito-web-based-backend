package api

import (
	"time"

	"github.com/incognitochain/go-incognito-sdk-v2/coin"
	"github.com/incognitochain/go-incognito-sdk-v2/common"
	metadataCommon "github.com/incognitochain/go-incognito-sdk-v2/metadata/common"
)

type UnshieldRequest struct {
	UnifiedTokenID common.Hash           `json:"UnifiedTokenID"`
	Data           []UnshieldRequestData `json:"Data"`
	Receiver       coin.OTAReceiver      `json:"Receiver"`
	IsDepositToSC  bool                  `json:"IsDepositToSC"`
	metadataCommon.MetadataBase
}

type UnshieldRequestData struct {
	IncTokenID        common.Hash `json:"IncTokenID"`
	BurningAmount     uint64      `json:"BurningAmount"`
	MinExpectedAmount uint64      `json:"MinExpectedAmount"`
	RemoteAddress     string      `json:"RemoteAddress"`
}

type EstimateSwapResult struct {
	EstimateReceive float64
	Fees            map[string]FeeModel
	Rewards         map[string]RewardModel
}

type FeeModel struct {
	FeeType string
	Fee     float64
	TokenID string
}

type RewardModel struct {
	RewardType string
	Reward     float64
	TokenID    string
}

type EstimateSwapRequest struct {
	Network   string
	Amount    string
	FromToken string // IncTokenID
	ToToken   string // IncTokenID
}

type EstimateSwapRespond struct {
	Networks map[string]interface{}
}

type QuoteDataResp struct {
	AppName      string
	AmountIn     string
	AmountInRaw  string
	AmountOut    string
	AmountOutRaw string
	Fee          []PappNetworkFee
	Route        interface{}
}

type SubmitSwapTx struct {
	Network string
	Txhash  string
}

type EstimateRewardRequest struct {
	UnifiedTokenID string
	TokenID        string
	Amount         uint64
}

type EstimateRewardRespond struct {
	ReceiveAmount uint64
	Reward        uint64
}

type EstimateUnshieldRequest struct {
	UnifiedTokenID string
	TokenID        string
	ExpectedAmount uint64
	BurntAmount    uint64
}

type EstimateUnshieldRespond struct {
	BurntAmount       uint64
	Fee               uint64
	ReceivedAmount    uint64
	MaxFee            uint64
	MinReceivedAmount uint64
}

type HistoryAddressResp struct {
	ID      uint
	UserID  uint
	Address string

	ExpiredAt time.Time

	EstFeeAt *time.Time

	AddressType int
	Status      int

	StatusMessage string
	StatusDetail  string

	CurrencyType       int
	Network            string
	WalletAddress      string
	UserPaymentAddress string
	RequestedAmount    string
	ReceivedAmount     string
	IncognitoAmount    string

	EthereumTx  string
	IncognitoTx string

	Erc20TokenTx string

	PrivacyTokenAddress string
	Erc20TokenAddress   string

	CreatedAt time.Time
	UpdatedAt time.Time

	Decentralized int

	OutChainTx string
	InChainTx  string

	TokenFee   string
	PrivacyFee string

	OutChainPrivacyFee string
	OutChainTokenFee   string

	BurnTokenFee   string
	BurnPrivacyFee string

	IncognitoTxToPayOutsideChainFee string

	Note string
	Memo string

	TxReceive string

	UnifiedStatus *UnifiedStatus

	UnifiedReward *UnifiedReward

	Decimals  int64
	PDecimals uint64
}

type UnifiedStatus struct {
	Fee            uint64
	ReceivedAmount uint64
	Status         int
}

type UnifiedReward struct {
	Status int
	Amount uint64
	Reward uint64
}

type GenUnshieldAddressRequest struct {
	Network             string
	RequestedAmount     string
	AddressType         int
	IncognitoAmount     string
	PaymentAddress      string
	PrivacyTokenAddress string
	WalletAddress       string
	IncognitoTx         string
	UnifiedTokenID      string
	SignPublicKeyEncode string
}

type SubmitUnshieldTxRequest struct {
	Network             string
	IncognitoAmount     string
	PaymentAddress      string
	PrivacyTokenAddress string
	WalletAddress       string

	UserFeeLevel     int
	ID               int
	IncognitoTx      string
	UserFeeSelection int
}

type GenShieldAddressRequest struct {
	Network             string
	AddressType         int
	PrivacyTokenAddress string
	WalletAddress       string
}

type SubmitShieldTx struct {
	Txhash  string
	Network int
	TokenID string
	Captcha string
}

type APIRespond struct {
	Result interface{}
	Error  *string
}

type TransactionDetail struct {
	BlockHash   string `json:"BlockHash"`
	BlockHeight uint64 `json:"BlockHeight"`
	TxSize      uint64 `json:"TxSize"`
	Index       uint64 `json:"Index"`
	ShardID     byte   `json:"ShardID"`
	Hash        string `json:"Hash"`
	Version     int8   `json:"Version"`
	Type        string `json:"Type"` // Transaction type
	LockTime    string `json:"LockTime"`
	RawLockTime int64  `json:"RawLockTime,omitempty"`
	Fee         uint64 `json:"Fee"` // Fee applies: always consant
	Image       string `json:"Image"`

	IsPrivacy bool `json:"IsPrivacy"`
	// Proof           privacy.Proof `json:"Proof"`
	// ProofDetail     interface{}   `json:"ProofDetail"`
	InputCoinPubKey string `json:"InputCoinPubKey"`
	SigPubKey       string `json:"SigPubKey,omitempty"` // 64 bytes
	RawSigPubKey    []byte `json:"RawSigPubKey,omitempty"`
	Sig             string `json:"Sig,omitempty"` // 64 bytes

	Metadata                      string      `json:"Metadata"`
	CustomTokenData               string      `json:"CustomTokenData"`
	PrivacyCustomTokenID          string      `json:"PrivacyCustomTokenID"`
	PrivacyCustomTokenName        string      `json:"PrivacyCustomTokenName"`
	PrivacyCustomTokenSymbol      string      `json:"PrivacyCustomTokenSymbol"`
	PrivacyCustomTokenData        string      `json:"PrivacyCustomTokenData"`
	PrivacyCustomTokenProofDetail interface{} `json:"PrivacyCustomTokenProofDetail"`
	PrivacyCustomTokenIsPrivacy   bool        `json:"PrivacyCustomTokenIsPrivacy"`
	PrivacyCustomTokenFee         uint64      `json:"PrivacyCustomTokenFee"`

	IsInMempool bool `json:"IsInMempool"`
	IsInBlock   bool `json:"IsInBlock"`

	Info string `json:"Info"`
}

type SubmitTxListRequest struct {
	TxList []string
}

type SubmitSwapTxRequest struct {
	TxRaw string
}

type TxStatusRespond struct {
	TxHash string
	Status string
	Error  string
}

type PappSupportedTokenData struct {
	ID                string
	ContractID        string
	ContractIDGetRate string
	Name              string
	Symbol            string
	PricePrv          float64
	Decimals          int
	PDecimals         int
	Protocol          string
	Verify            bool
	IsPopular         bool
	Priority          int
	DappID            int
	CurrencyType      int
	NetworkID         int
	MovedUnifiedToken bool
	NetworkName       string
}

type UniswapQuote struct {
	Data struct {
		AmountIn         string      `json:"amountIn"`
		AmountOut        string      `json:"amountOut"`
		AmountOutRaw     string      `json:"amountOutRaw"`
		Route            interface{} `json:"route"`
		Impact           float64     `json:"impact"`
		EstimatedGasUsed string      `json:"estimatedGasUsed"`
	} `json:"data"`
	Message string `json:"message"`
}

type PancakeQuote struct {
	Data struct {
		Outputs []string    `json:"outputs"`
		Route   interface{} `json:"paths"`
		Impact  float64     `json:"impactAmount"`
	} `json:"data"`
	Message string `json:"message"`
}

type PappNetworkFee struct {
	FeeAddress string `json:"feeAddress"`
	Amount     uint64 `json:"amount"`
}

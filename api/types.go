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
	FeeAddress  string
	Network     string
	Amount      string
	Slippage    string
	MultiTrades bool
	MinSplit    int
	FromToken   string // IncTokenID
	ToToken     string // IncTokenID
	ShardID     string

	IsFromInterswap bool // dafault false
}

type EstimateSwapRespond struct {
	Networks      map[string]interface{}
	NetworksError map[string]interface{}
}

type QuoteDataResp struct {
	AppName              string
	CallContract         string
	AmountIn             string
	AmountInRaw          string
	AmountOut            string
	AmountOutRaw         string
	AmountOutPreSlippage string
	RedepositReward      string
	Rate                 string
	Fee                  []PappNetworkFee
	FeeAddress           string
	FeeAddressShardID    int
	Paths                interface{}
	PathsContract        interface{}
	PoolPairs            []string
	Calldata             string
	ImpactAmount         string
	RouteDebug           interface{}
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

	Network string
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
	CurrencyType        int
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

	//centralized
	PrivacyFee                      string
	TokenFee                        string
	Address                         string
	IncognitoTxToPayOutsideChainFee string
}

type GenShieldAddressRequest struct {
	Network             string
	AddressType         int
	CurrencyType        int
	PrivacyTokenAddress string
	WalletAddress       string
	RequestedAmount     string
	IncognitoAmount     string
	PaymentAddress      string

	BTCIncAddress string
}
type GenBTCShieldAddressRequest struct {
	ShieldAddress string `json:"btcaddress"`
	IncAddress    string `json:"incaddress"`
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
	TxRaw            string
	TxHash           string
	FeeRefundOTA     string
	FeeRefundAddress string
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
		AmountIn         string           `json:"amountIn"`
		AmountOut        string           `json:"amountOut"`
		AmountOutRaw     string           `json:"amountOutRaw"`
		Route            [][]UniswapRoute `json:"route"`
		Impact           float64          `json:"impact"`
		EstimatedGasUsed string           `json:"estimatedGasUsed"`
	} `json:"data"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

type UniswapRoute struct {
	AmountIn          string            `json:"amountIn"`
	AmountOut         string            `json:"amountOut"`
	Fee               int64             `json:"fee"`
	Liquidity         string            `json:"liquidity"`
	Percent           float64           `json:"percent"`
	Type              string            `json:"type"`
	PoolAddress       string            `json:"poolAddress"`
	RawQuote          string            `json:"rawQuote"`
	SqrtPriceX96After string            `json:"sqrtPriceX96After"`
	TokenIn           UniswapQuoteToken `json:"tokenIn"`
	TokenOut          UniswapQuoteToken `json:"tokenOut"`
}

type UniswapQuoteToken struct {
	Address  string `json:"address"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	IsNative bool   `json:"isNative"`
}

type PancakeQuote struct {
	Data struct {
		Outputs []string `json:"outputs"`
		Route   []string `json:"paths"`
		Impact  float64  `json:"impactAmount"`
	} `json:"data"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

type PappNetworkFee struct {
	TokenID          string  `json:"tokenid"`
	Amount           uint64  `json:"amount"`
	AmountInBuyToken string  `json:"amountInBuyToken"`
	PrivacyFee       uint64  `json:"privacyFee"`
	FeeInUSD         float64 `json:"feeInUSD"`
}

type UnshieldNetworkFee struct {
	FeeAddress        string  `json:"feeAddress"`
	FeeAddressShardID int     `json:"feeAddressShardID"`
	ExpectedReceive   uint64  `json:"expectedReceive"`
	BurntAmount       uint64  `json:"burntAmount"`
	TokenID           string  `json:"tokenid"`
	Amount            uint64  `json:"feeAmount"`
	PrivacyFee        uint64  `json:"privacyFee"`
	ProtocolFee       uint64  `json:"protocolFee"`
	FeeInUSD          float64 `json:"feeInUSD"`
}

type OpenSeaFee struct {
	FeeAddress        string `json:"feeAddress"`
	FeeAddressShardID int    `json:"feeAddressShardID"`
	TokenID           string `json:"tokenid"`
	Amount            uint64 `json:"feeAmount"`
	PrivacyFee        uint64 `json:"privacyFee"`
	// ProtocolFee       uint64  `json:"protocolFee"`
	FeeInUSD float64 `json:"feeInUSD"`
}

type PancakeTokenMapItem struct {
	Decimals int    `json:"decimals"`
	Symbol   string `json:"symbol"`
}

type CurvePoolIndex struct {
	CurveTokenIndex  int
	CurvePoolAddress string
	DappTokenAddress string
	DappTokenSymbol  string
}

type StatusSwapTxDetail struct {
	SellToken  string
	SellAmount uint64
	BuyToken   string
	Networks   []string
}

type PdexEstimateRespond struct {
	SellAmount    float64
	MaxGet        float64
	Fee           uint64
	Route         []string
	TokenRoute    []string
	IsSignificant bool
	ImpactAmount  float64
}

type RetrySwapTx struct {
	Txs []string
}
type APITokenInfoRequest struct {
	TokenIDs []string
	Nocache  bool
}

type ShieldStatusData struct {
	Amount uint64 `json:"Amount"`
	Reward uint64 `json:"Reward"`
}

type ShieldStatus struct {
	Status    byte               `json:"Status"`
	Data      []ShieldStatusData `json:"Data,omitempty"`
	ErrorCode int                `json:"ErrorCode,omitempty"`
}

type DexSwap struct {
	Txhash       string `json:"txhash"`
	TokenSell    string `json:"token_sell"`
	TokenBuy     string `json:"token_buy"`
	AmountIn     string `json:"amount_in"`
	MinAmountOut string `json:"amount_out"`
}

type TokenStruct struct {
	ID string `json:"id"`
}

// Pdao request
type CreatProposalReq struct {
	Txhash string `json:"Txhash" binding:"required"`
	TxRaw  string `json:"TxRaw" binding:"required"`

	ProposalID string `json:"ProposalID" binding:"required"`

	Targets    []string `json:"Targets" binding:"required"`
	Values     []string `json:"Values" binding:"required"`
	Signatures []string `json:"Signatures"`
	Calldatas  []string `json:"Calldatas"  binding:"required"`

	Description         string `json:"Description"  binding:"required"`
	Title               string `json:"Title"  binding:"required"`
	ReShieldSignature   string `json:"ReShieldSignature"  binding:"required"`
	CreatePropSignature string `json:"CreatePropSignature" binding:"required"`
	PropVoteSignature   string `json:"PropVoteSignature" binding:"required"`
}

type SubmitVoteReq struct {
	Txhash            string `json:"Txhash" binding:"required"`
	TxRaw             string
	ProposalID        string `json:"ProposalID" binding:"required"`
	Vote              uint8  `json:"Vote" binding:"required"`
	PropVoteSignature string `json:"PropVoteSignature" binding:"required"`
	ReShieldSignature string `json:"ReShieldSignature" binding:"required"`
}

type SubmitCancelReq struct {
	Txhash            string `json:"Txhash" binding:"required"`
	TxRaw             string `json:"TxRaw" binding:"required"`
	ProposalID        string `json:"ProposalID" binding:"required"`
	Signature         string `json:"Signature" binding:"required"`
	ReShieldSignature bool   `json:"ReShieldSignature" binding:"required"`
}

type ReShieldSignatureReq struct {
	Signature        string
	IncognitoAddress string
	Amount           string
	Timestamp        string
}

type PDaoNetworkFeeResp struct {
	FeeAddress        string `json:"feeAddress"`
	FeeAddressShardID int    `json:"feeAddressShardID"`
	TokenID           string `json:"tokenid"`
	FeeAmount         uint64 `json:"feeAmount"`
}

type PnftListingReq struct {
	Items []PnftListingItem `json:"items"`
}

type PnftListingItem struct {
	//TODO add more
	Collection string `json:"collection"`
	TokenID    string `json:"token_id"`
	Amount     string `json:"amount"`
	Signature  string `json:"signature"`
}
type PnftDelistingReq struct {
	Items []PnftDelistingItem `json:"items"`
}

type PnftDelistingItem struct {
	//TODO add more
	Collection string `json:"collection"`
	TokenID    string `json:"token_id"`
	Amount     string `json:"amount"`
	Signature  string `json:"signature"`
}

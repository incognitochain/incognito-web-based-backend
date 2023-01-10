package common

import "math/big"

type Config struct {
	Port            int
	Mode            string
	Mongo           string
	Mongodb         string
	CoinserviceURL  string
	FullnodeURL     string
	FullnodeAuthKey string
	ShieldService   string
	BTCShieldPortal string
	FaucetService   string
	NetworkID       string
	CaptchaSecret   string
	SlackMonitor    string
	// papps submit proof
	IncKey string
	EVMKey string

	CentralIncPaymentAddress string

	GGCProject string
	GGCAuth    string

	OpenSeaAPI    string
	OpenSeaAPIKey string

	// Interswap
	ISIncPrivKeys map[string]string
}

type TokenInfo struct {
	TokenID            string
	Name               string
	Symbol             string
	Image              string
	IsPrivacy          bool
	IsBridge           bool
	ExternalID         string
	PDecimals          int
	Decimals           int64
	ContractID         string
	Status             int
	Type               int
	CurrencyType       int
	Default            bool
	Verified           bool
	UserID             int
	ListChildToken     []TokenInfo
	ListUnifiedToken   []TokenInfo
	PSymbol            string
	OriginalSymbol     string
	LiquidityReward    float64
	ExternalPriceUSD   float64 `json:"ExternalPriceUSD"` // use to convert token
	PriceUsd           float64 `json:"PriceUsd"`
	PercentChange1h    string  `json:"PercentChange1h"`
	PercentChangePrv1h string  `json:"PercentChangePrv1h"`
	PercentChange24h   string  `json:"PercentChange24h"`
	CurrentPrvPool     uint64  `json:"CurrentPrvPool"`
	PricePrv           float64 `json:"PricePrv"`
	Volume24           uint64  `json:"volume24"`
	ParentID           int     `json:"ParentID"`
	Network            string
	DefaultPoolPair    string
	DefaultPairToken   string
	//additional p-unified token
	NetworkID         int
	MovedUnifiedToken bool
	ParentUnifiedID   int
	IsSwapable        bool
	ContractIDSwap    string
}

type ExternalTxSwapResult struct {
	LogResult   string
	IsRedeposit bool
	IsReverted  bool
	IsFailed    bool

	// token repsonse : failed/sucess
	TokenContract string
	Amount        *big.Int
}

type PappSwapInfo struct {
	DappName       string
	TokenIn        string
	TokenOut       string
	TokenInAmount  *big.Int
	MinOutAmount   *big.Int
	AdditionalData string
}

type TradeDataRespond struct {
	RequestTx           string
	RespondTxs          []string
	RespondTokens       []string
	RespondAmounts      []uint64
	WithdrawTxs         map[string]TradeWithdrawInfo
	SellTokenID         string
	BuyTokenID          string
	Status              string
	StatusCode          int
	PairID              string
	PoolID              string
	MinAccept           uint64
	Amount              uint64
	Matched             uint64
	Requestime          int64
	NFTID               string
	Receiver            string
	Fee                 uint64
	FeeToken            string
	IsCompleted         bool
	SellTokenBalance    uint64
	BuyTokenBalance     uint64
	SellTokenWithdrawed uint64
	BuyTokenWithdrawed  uint64
	TradingPath         []string
}

type TradeWithdrawInfo struct {
	TokenIDs   []string
	IsRejected bool
	Responds   map[string]struct {
		Amount    uint64
		Status    int
		RespondTx string
	}
}

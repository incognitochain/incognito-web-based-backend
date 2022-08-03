package common

import "github.com/kamva/mgm/v3"

type Config struct {
	Port           int
	Mode           string
	Mongo          string
	Mongodb        string
	CoinserviceURL string
	FullnodeURL    string
	ShieldService  string
	FaucetService  string
	NetworkID      string
	CaptchaSecret  string
	// papps submit proof
	IncKey string
	EVMKey string

	GGCProject string
	GGCAuth    string
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
	ExternalPriceUSD   float64 `json:"ExternalPriceUSD"`
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
}

type ShieldTxData struct {
	mgm.DefaultModel `bson:",inline"`
	Status           string `json:"status" bson:"status"`
	ExternalTx       string `json:"externaltx" bson:"externaltx"`
	NetworkID        int    `json:"networkid" bson:"networkid"`
	TokenID          string `json:"tokenid" bson:"tokenid"`
	UTokenID         string `json:"utokenid" bson:"utokenid"`
	PaymentAddress   string `json:"paymentaddress" bson:"paymentaddress"`
	IncTx            string `json:"inctx" bson:"inctx"`
	Error            string `json:"error" bson:"error"`
}

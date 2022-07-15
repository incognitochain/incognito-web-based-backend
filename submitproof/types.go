package submitproof

type TokenInfo struct {
	TokenID            string
	Name               string
	Symbol             string
	Image              string
	IsPrivacy          bool
	IsBridge           bool
	ExternalID         string
	PDecimals          int
	Decimals           uint64
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

type SubmitProofStatus struct {
	Status       int
	StatusString string
	Error        string
}

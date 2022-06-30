package api

import "time"

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
	SourceToken string
	// SourceNetwork string
	DestToken   string
	DestNetwork string
	Amount      float64
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

	Decimals  uint64
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
}

type SubmitUnshieldTxRequest struct {
	Network             string
	RequestedAmount     string
	AddressType         int
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

package common

import "github.com/kamva/mgm/v3"

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

type ExternalNetworksFeeData struct {
	mgm.DefaultModel `bson:",inline"`
	GasPrice         map[string]uint64 `json:"fees" bson:"fees"`
}

type PAppsEndpointData struct {
	mgm.DefaultModel `bson:",inline"`
	Network          string            `json:"network" bson:"network"`
	ExchangeApps     map[string]string `json:"excapps" bson:"excapps"`
	AppContracts     map[string]string `json:"appcontracts" bson:"appcontracts"`
}

type BridgeNetworkData struct {
	mgm.DefaultModel   `bson:",inline"`
	Network            string   `json:"network" bson:"network"`
	ChainID            string   `json:"chainid" bson:"chainid"`
	Endpoints          []string `json:"endpoints" bson:"endpoints"`
	ConfirmationBlocks int      `json:"confirmationblocks" bson:"confirmationblocks"`
}

type PappSupportedTokenData struct {
	mgm.DefaultModel `bson:",inline"`
	TokenID          string `json:"tokenid" bson:"tokenid"`
	ContractID       string `json:"contractid" bson:"contractid"`
	Verify           bool   `json:"verify" bson:"verify"`
}

type ExternalTxStatus struct {
	mgm.DefaultModel   `bson:",inline"`
	Txhash             string `json:"txhash" bson:"txhash"`
	IncRequestTx       string `json:"increquesttx" bson:"increquesttx"`
	Network            string `json:"network" bson:"network"`
	Status             string `json:"status" bson:"status"`
	Type               int    `json:"type" bson:"type"`
	Error              string `json:"error" bson:"error"`
	OtherInfo          string `json:"otherinfo" bson:"otherinfo"`
	Nonce              uint64 `json:"nonce" bson:"nonce"`
	WillRedeposit      bool   `json:"will_redeposit" bson:"will_redeposit"`
	RedepositSubmitted bool   `json:"redeposit_submitted" bson:"redeposit_submitted"`
}

type PappTxData struct {
	mgm.DefaultModel `bson:",inline"`
	IncTx            string   `json:"inctx" bson:"inctx"`
	Networks         []string `json:"networks" bson:"networks"`
	Type             int      `json:"type" bson:"type"`
	IncTxData        string   `json:"inctxdata" bson:"inctxdata"`
	FeeToken         string   `json:"feetoken" bson:"feetoken"`
	FeeAmount        uint64   `json:"feeamount" bson:"feeamount"`
	PFeeAmount       uint64   `json:"pfeeamount" bson:"pfeeamount"`
	BurntAmount      uint64   `json:"burntamount" bson:"burntamount"`
	BurntToken       string   `json:"burnttoken" bson:"burnttoken"`
	PappSwapInfo     string   `json:"pappswapinfo" bson:"pappswapinfo"`
	Status           string   `json:"status" bson:"status"`
	IsUnifiedToken   bool     `json:"isunifiedtoken" bson:"isunifiedtoken"`
	RefundSubmitted  bool     `json:"refundsubmitted" bson:"refundsubmitted"`
	RefundPrivacyFee bool     `json:"refundpfee" bson:"refundpfee"`
	FeeRefundOTA     string   `json:"fee_refundota" bson:"fee_refundota"`
	FeeRefundAddress string   `json:"fee_refundaddress" bson:"fee_refundaddress"`
	ShardID          int      `json:"shardid" bson:"shardid"`
	OutchainStatus   string   `json:"outchain_status" bson:"outchain_status"`
	UserAgent        string   `json:"useragent" bson:"useragent"`
	Error            string   `json:"error" bson:"error"`
}

type PappVaultData struct {
	mgm.DefaultModel `bson:",inline"`
	Network          string `json:"network" bson:"network"`
	Type             int    `json:"type" bson:"type"`
	ContractAddress  string `json:"contactaddress" bson:"contactaddress"`
}

type EVMProofRecordData struct {
	mgm.DefaultModel `bson:",inline"`
	BlockNumber      uint64 `json:"blocknumber" bson:"blocknumber"`
	BlockHash        string `json:"blockhash" bson:"blockhash"`
	TxIndex          uint   `json:"txidx" bson:"txidx"`
	Proof            string `json:"proof" bson:"proof"`
	ContractID       string `json:"contractid" bson:"contractid"`
	PaymentAddr      string `json:"paymentaddr" bson:"paymentaddr"`
	IsRedeposit      bool   `json:"isredeposit" bson:"isredeposit"`
	IsTxPass         bool   `json:"istxpass" bson:"istxpass"`
	OTAStr           string `json:"otaStr" bson:"otaStr"`
	Amount           uint64 `json:"amount" bson:"amount"`
	Network          string `json:"network" bson:"network"`
}

type RefundFeeData struct {
	mgm.DefaultModel `bson:",inline"`
	IncRequestTx     string `json:"increquesttx" bson:"increquesttx"`
	RefundAmount     uint64 `json:"refundamount" bson:"refundamount"`
	RefundToken      string `json:"refundtoken" bson:"refundtoken"`
	RefundOTA        string `json:"refundota" bson:"refundota"`
	RefundAddress    string `json:"refundaddress" bson:"refundaddress"`
	RefundTx         string `json:"refundtx" bson:"refundtx"`
	RefundStatus     string `json:"status" bson:"status"`
	RefundPrivacyFee bool   `json:"ispfee" bson:"ispfee"`
	Error            string `json:"error" bson:"error"`
}

type DexSwapTrackData struct {
	mgm.DefaultModel `bson:",inline"`
	IncTx            string `json:"inctx" bson:"inctx"`
	Status           string `json:"status" bson:"status"`
	TokenSell        string `json:"token_sell" bson:"token_sell"`
	TokenBuy         string `json:"token_buy" bson:"token_buy"`
	AmountIn         string `json:"amount_in" bson:"amount_in"`
	MinAmountOut     string `json:"amount_out" bson:"amount_out"`
	UserAgent        string `json:"useragent" bson:"useragent"`
}

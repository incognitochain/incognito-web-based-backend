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
	Fees             map[string]uint64 `json:"fees" bson:"fees"`
}

type PAppsEndpointData struct {
	mgm.DefaultModel `bson:",inline"`
	Network          string            `json:"network" bson:"network"`
	ExchangeApps     map[string]string `json:"excapps" bson:"excapps"`
}

type BridgeNetworkData struct {
	mgm.DefaultModel   `bson:",inline"`
	Network            string   `json:"network" bson:"network"`
	ChainID            string   `json:"chainid" bson:"chainid"`
	Endpoints          []string `json:"endpoints" bson:"endpoints"`
	ConfirmationBlocks int      `json:"confirmationblocks" bson:"confirmationblocks"`
}

// type PappSupportedTokenData struct {
// 	mgm.DefaultModel  `bson:",inline"`
// 	TokenID           string `json:"tokenid" bson:"tokenid"`
// 	ContractID        string `json:"contractid" bson:"contractid"`
// 	ContractIDGetRate string `json:"contractid_getrate" bson:"contractid_getrate"`
// 	Protocol          string `json:"protocol" bson:"protocol"`
// 	Verify            bool   `json:"verify" bson:"verify"`
// 	IsPopular         bool   `json:"ispopular" bson:"ispopular"`
// 	Priority          int    `json:"priority" bson:"priority"`
// 	DappID            int    `json:"dappid" bson:"dappid"`
// 	NetworkID         int    `json:"networkid" bson:"networkid"`
// }

type ExternalTxStatus struct {
	mgm.DefaultModel `bson:",inline"`
	Txhash           string `json:"txhash" bson:"txhash"`
	IncRequestTx     string `json:"increquesttx" bson:"increquesttx"`
	Network          string `json:"network" bson:"network"`
	Status           string `json:"status" bson:"status"`
	Type             int    `json:"type" bson:"type"`
	Error            string `json:"error" bson:"error"`
}

type PappTxData struct {
	mgm.DefaultModel `bson:",inline"`
	IncTx            string `json:"inctx" bson:"inctx"`
	ExternalTx       string `json:"externaltx" bson:"externaltx"`
	Network          string `json:"network" bson:"network"`
	Type             int    `json:"type" bson:"type"`
	IncTxData        string `json:"inctxdata" bson:"inctxdata"`
	ExternalTxData   string `json:"externaltxdata" bson:"externaltxdata"`
	FeeToken         string `json:"feetoken" bson:"feetoken"`
	FeeAmount        uint64 `json:"feeamount" bson:"feeamount"`
	Status           string `json:"status" bson:"status"`
	IsUnifiedToken   bool   `json:"isunifiedtoken" bson:"isunifiedtoken"`
}

type PappContractData struct {
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
	OTAStr           string `json:"otaStr" bson:"otaStr"`
	Amount           uint64 `json:"amount" bson:"amount"`
	Network          string `json:"network" bson:"network"`
}

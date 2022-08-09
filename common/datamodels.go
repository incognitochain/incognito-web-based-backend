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
	mgm.DefaultModel `bson:",inline"`
	Network          string   `json:"network" bson:"network"`
	ChainID          string   `json:"chainid" bson:"chainid"`
	Endpoints        []string `json:"endpoints" bson:"endpoints"`
}

type PappSupportedTokenData struct {
	mgm.DefaultModel  `bson:",inline"`
	TokenID           string `json:"tokenid" bson:"tokenid"`
	ContractID        string `json:"contractid" bson:"contractid"`
	ContractIDGetRate string `json:"contractid_getrate" bson:"contractid_getrate"`
	Protocol          string `json:"protocol" bson:"protocol"`
	Verify            bool   `json:"verify" bson:"verify"`
	IsPopular         bool   `json:"ispopular" bson:"ispopular"`
	Priority          int    `json:"priority" bson:"priority"`
	DappID            int    `json:"dappid" bson:"dappid"`
	NetworkID         int    `json:"networkid" bson:"networkid"`
}

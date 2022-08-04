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
	Fees             map[string]interface{} `json:"fees" bson:"fees"`
}

type PAppsEndpointData struct {
	mgm.DefaultModel `bson:",inline"`
	Network          string            `json:"network" bson:"network"`
	ExchangeApps     map[string]string `json:"excapps" bson:"excapps"`
}

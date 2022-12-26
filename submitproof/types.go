package submitproof

import (
	"time"

	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	"github.com/incognitochain/incognito-web-based-backend/common"
)

type SubmitProofShieldTask struct {
	TxHash    string
	NetworkID int
	TokenID   string
	Metatype  string
}

type SubmitPappTxTask struct {
	TxHash           string
	TxRawData        []byte
	IsPRVTx          bool
	IsUnifiedToken   bool
	FeeToken         string
	FeeAmount        uint64
	PFeeAmount       uint64
	FeeRefundOTA     string
	FeeRefundAddress string
	BurntToken       string
	BurntAmount      uint64
	PappSwapInfo     *common.PappSwapInfo
	Networks         []string
	Time             time.Time
	UserAgent        string
}

type SubmitUnshieldTxTask struct {
	TxHash           string
	TxRawData        []byte
	IsPRVTx          bool
	IsUnifiedToken   bool
	FeeToken         string
	FeeAmount        uint64
	PFeeAmount       uint64
	FeeRefundOTA     string
	FeeRefundAddress string
	Token            string
	UToken           string
	BurntAmount      uint64
	ExternalAddress  string
	Networks         []string
	Time             time.Time
	UserAgent        string
}

type SubmitRefundFeeTask struct {
	IncReqTx string
	Token    string
	OTA      string
	// OTASS          string
	IsPrivacyFeeRefund bool
	PaymentAddress     string
	Amount             uint64
	Time               time.Time
}

type SubmitProofOutChainTask struct {
	IncTxhash      string
	Network        string
	IsUnifiedToken bool
	IsRetry        bool
	Type           int
	Time           time.Time
}

// type WatchShieldProofTask struct {
// 	PaymentAddress string
// 	Txhash         string
// 	NetworkID      int
// 	TokenID        string
// 	IsPunified     bool
// 	IncTx          string
// 	Time           time.Time
// }

type SubmitProofConsumer struct {
	UseKey    string
	NetworkID int
}

type EVMProofResult struct {
	Proof *incclient.EVMDepositProof
}

package submitproof

import (
	"time"

	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
)

type SubmitProofShieldTask struct {
	TxHash    string
	NetworkID int
	TokenID   string
	Metatype  string
	Time      time.Time
}

type SubmitPappTxTask struct {
	TxHash         string
	TxRawData      []byte
	IsPRVTx        bool
	IsUnifiedToken bool
	FeeToken       string
	FeeAmount      uint64
	Networks       []string
	Time           time.Time
}

type SubmitPappProofOutChainTask struct {
	IncTxhash      string
	Network        string
	IsUnifiedToken bool
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

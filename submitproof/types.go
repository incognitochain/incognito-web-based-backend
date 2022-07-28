package submitproof

import "time"

type SubmitProofShieldTask struct {
	Txhash    string
	NetworkID int
	TokenID   string
	Metatype  string
	Time      time.Time
}

type WatchShieldProofTask struct {
	PaymentAddress string
	Txhash         string
	NetworkID      int
	TokenID        string
	IsPunified     bool
	IncTx          string
	Time           time.Time
}

type SubmitProofConsumer struct {
	UseKey    string
	NetworkID int
}

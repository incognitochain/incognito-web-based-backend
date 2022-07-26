package submitproof

import "time"

type SubmitProofShieldTask struct {
	Txhash    string
	NetworkID int
	TokenID   string
	Metatype  string
	Time      time.Time
}

type WatchProofTask struct {
	Txhash    string
	NetworkID int
	TokenID   string
	IncTx     string
	Time      time.Time
}

type SubmitProofConsumer struct {
	UseKey    string
	NetworkID int
}

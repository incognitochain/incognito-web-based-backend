package submitproof

type SubmitProofTask struct {
	Txhash    string
	NetworkID int
	TokenID   string
	Metatype  string
}

type WatchProofTask struct {
	Txhash string
	IncTx  string
}

type TaskConsumer struct {
	UseKey string
}

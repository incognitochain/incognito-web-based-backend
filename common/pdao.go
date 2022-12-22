package common

import "github.com/kamva/mgm/v3"

type Proposal struct {
	mgm.DefaultModel    `bson:",inline"`
	IncTx               string `bson:"inctx"`
	SubmitBurnTx        string `bson:"submit_burn_tx"`
	SubmitProposalTx    string `bson:"submit_proposal_tx"`
	Status              string `bson:"status"`
	ProposalID          string `bson:"token_sell"`
	Proposer            string `bson:"proposer"`
	Targets             string `bson:"targets"`
	Values              string `bson:"values"`
	Signatures          string `bson:"signatures"`
	Calldatas           string `bson:"calldatas"`
	CreatePropSignature string `bson:"create_prop_signature"`
	DescriptionLink     string `bson:"description_link"`
}

type Vote struct {
	mgm.DefaultModel `bson:",inline"`
	Status           string `bson:"status"`
	Txhash           string
	ProposalID       string
	Vote             uint8
	VoteSignature    string
	Reshield         string `bson:"reshield"`
	SubmitBurnTx     string `bson:"submit_burn_tx"`
	SubmitVoteTx     string `bson:"submit_vote_tx"`
}

type Cancel struct {
	mgm.DefaultModel `bson:",inline"`
	Status           string `bson:"status"`
	Txhash           string
	ProposalID       string
	CancelSignature  string
	Reshield         string `bson:"reshield"`
	SubmitCancelTx   string `bson:"submit_cancel_tx"`
}

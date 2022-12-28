package common

import "github.com/kamva/mgm/v3"

type Proposal struct {
	mgm.DefaultModel    `bson:",inline"`
	SubmitBurnTx        string `bson:"submit_burn_tx"`
	SubmitProposalTx    string `bson:"submit_proposal_tx"`
	Status              string `bson:"status"`
	ProposalID          string `bson:"proposal_id"`
	Proposer            string `bson:"proposer"`
	Targets             string `bson:"targets"`
	Values              string `bson:"values"`
	Signatures          string `bson:"signatures"`
	Calldatas           string `bson:"calldatas"`
	CreatePropSignature string `bson:"create_prop_signature"`
	PropVoteSignature   string
	Description         string `bson:"description"`
	Title               string `bson:"title"`
	ReShieldSignature   string
}

type Vote struct {
	mgm.DefaultModel  `bson:",inline"`
	Status            string `bson:"status"`
	ProposalID        string
	Vote              uint8
	PropVoteSignature string
	ReShieldSignature string
	SubmitBurnTx      string `bson:"submit_burn_tx"`
	SubmitVoteTx      string `bson:"submit_vote_tx"`
	SubmitReShieldTx  string `bson:"submit_re_shield_tx"`
	AutoVoted         bool
	IsReShield        bool `bson:"is_re_shield"`
}

type Cancel struct {
	mgm.DefaultModel  `bson:",inline"`
	Status            string `bson:"status"`
	ProposalID        string
	CancelSignature   string
	ReShieldSignature string
	SubmitCancelTx    string
}

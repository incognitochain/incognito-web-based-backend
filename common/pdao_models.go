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
	SubmitBurnTx      string `bson:"submit_burn_tx"`
	SubmitVoteTx      string `bson:"submit_vote_tx"`
	AutoVoted         bool

	ReShieldSignature string
	ReShieldStatus    string `bson:"reshield_status"`
	SubmitReShieldTx  string `bson:"submit_re_shield_tx"`

	SubmitReShieldMintTx int `bson:"submit_reshield_mint_tx"`
}

type Cancel struct {
	mgm.DefaultModel  `bson:",inline"`
	Status            string `bson:"status"`
	ProposalID        string
	CancelSignature   string
	ReShieldSignature string
	SubmitCancelTx    string
}

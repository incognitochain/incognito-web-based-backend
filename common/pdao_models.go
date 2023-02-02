package common

import "github.com/kamva/mgm/v3"

type Proposal struct {
	mgm.DefaultModel    `bson:",inline"`
	PID                 uint   `bson:"pid"`
	ProposalID          string `bson:"proposal_id"`
	Title               string `bson:"title"`
	Description         string `bson:"description"`
	Status              string `bson:"status"`
	SubmitBurnTx        string `bson:"submit_burn_tx"`
	SubmitProposalTx    string `bson:"submit_proposal_tx"`
	Proposer            string `bson:"proposer"`
	Targets             string `bson:"targets"`
	Values              string `bson:"values"`
	Signatures          string `bson:"signatures" json:"-"`
	Calldatas           string `bson:"calldatas" json:"-"`
	CreatePropSignature string `bson:"create_prop_signature" json:"-"`
	PropVoteSignature   string `json:"-"`
	ReShieldSignature   string `json:"-"`

	VoteFor     int `bson:"vote_for" json:"VoteFor"`
	VoteAgainst int `bson:"vote_against" json:"VoteAgainst"`
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

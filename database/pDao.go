package database

import (
	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
)

// proposal:
// insert a proposal:
func DBInsertProposalTable(data *common.Proposal) error {
	return mgm.Coll(data).Create(data)
}

// update a proposal:
func DBUpdateProposalTable(data *common.Proposal) error {
	return mgm.Coll(&common.Proposal{}).Update(data)
}

// load all proposal:
func DBListProposalTable() []common.Proposal {

	p := []common.Proposal{}
	mgm.Coll(&common.Proposal{}).SimpleFind(&p, bson.M{})
	return p
}

// create a voting:
func DBInsertVoteTable(data *common.Vote) error {
	return mgm.Coll(data).Create(data)
}

// update a voting:
func DBUpdateVoteTable(data *common.Vote) error {
	return mgm.Coll(&common.Vote{}).Update(data)
}

// list voting:
func DBListVoteTable() []common.Vote {

	l := []common.Vote{}
	mgm.Coll(&common.Vote{}).SimpleFind(&l, bson.M{})
	return l
}

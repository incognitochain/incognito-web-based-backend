package database

import (
	"context"
	"time"

	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/kamva/mgm/v3"
	"github.com/kamva/mgm/v3/operator"
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

// get by id
func DBgetProposalTable(id string) (*common.Proposal, error) {

	p := &common.Proposal{}
	err := mgm.Coll(p).FindByID(id, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func DBGetProposalByIncTx(incReqTx string) (*common.Proposal, error) {

	result := common.Proposal{}
	filter := bson.M{"submit_burn_tx": bson.M{operator.Eq: incReqTx}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	dbresult := mgm.Coll(&common.Proposal{}).FindOne(ctx, filter)
	if dbresult.Err() != nil {
		return nil, dbresult.Err()
	}

	if err := dbresult.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func DBGetPendingProposal() ([]common.Proposal, error) {
	result := []common.Proposal{}
	filter := bson.M{"status": bson.M{operator.In: []string{common.StatusSubmitting}}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.Proposal{}).SimpleFindWithCtx(ctx, &result, filter)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func DBGetSuccessProposalNoVoted() ([]common.Proposal, error) {
	result := []common.Proposal{}
	filter := bson.M{"status": bson.M{operator.In: []string{common.StatusPdaOutchainTxSuccess}}, "voted": false}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.Proposal{}).SimpleFindWithCtx(ctx, &result, filter)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func DBUpdatePdaoProposalStatus(incTx string, status string) error {
	filter := bson.M{"submit_burn_tx": bson.M{operator.Eq: incTx}}
	update := bson.M{"$set": bson.M{"status": status}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.Proposal{}).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func DBUpdatePdaoVoteStatus(incTx string, status string) error {
	filter := bson.M{"submit_burn_tx": bson.M{operator.Eq: incTx}}
	update := bson.M{"$set": bson.M{"status": status}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.Vote{}).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func DBUpdatePdaoCancelStatus(incTx string, status string) error {
	//TODO: @phuong => don't need yes for now.
	filter := bson.M{"submit_burn_tx": bson.M{operator.Eq: incTx}}
	update := bson.M{"$set": bson.M{"status": status}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.Cancel{}).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
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

func DBGetVoteByIncTx(incReqTx string) (*common.Vote, error) {

	result := common.Vote{}
	filter := bson.M{"submit_burn_tx": bson.M{operator.Eq: incReqTx}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	dbresult := mgm.Coll(&common.Vote{}).FindOne(ctx, filter)
	if dbresult.Err() != nil {
		return nil, dbresult.Err()
	}

	if err := dbresult.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

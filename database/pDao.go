package database

import (
	"context"
	"time"

	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/kamva/mgm/v3"
	"github.com/kamva/mgm/v3/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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
func DBListLimitProposalTable() []common.Proposal {
	var result []common.Proposal
	filter := bson.M{}
	// limit := int64(100)
	mgm.Coll(&common.Proposal{}).SimpleFind(&result, filter, &options.FindOptions{
		Sort: bson.D{{"created_at", -1}},
		//Limit: &limit,
	})
	return result
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

// get by id
func DBgetProposalTableByPID(pid uint) (*common.Proposal, error) {

	p := &common.Proposal{}
	// filter := bson.M{"pid": pid}
	filter := bson.M{"pid": bson.M{operator.Eq: pid}}
	err := mgm.Coll(p).First(filter, p)
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

func DBVoteForPdaoProposal(proposal_id string) error {
	filter := bson.M{"proposal_id": bson.M{operator.Eq: proposal_id}}
	update := bson.M{"$inc": bson.M{"vote_for": 1}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.Proposal{}).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}
func DBVoteAgainstPdaoProposal(proposal_id string) error {
	filter := bson.M{"proposal_id": bson.M{operator.Eq: proposal_id}}
	update := bson.M{"$inc": bson.M{"vote_against": 1}}
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

func DBUpdatePdaoVoteReshieldStatus(incTx string, status string) error {
	filter := bson.M{"submit_burn_tx": bson.M{operator.Eq: incTx}}
	update := bson.M{"$set": bson.M{"reshield_status": status}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.Vote{}).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func DBUpdatePdaoCancelStatus(incTx string, status string) error {
	//TODO: @phuong => don't need use for now.
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

// create a auto vote from proposal:
func DBCreateVoteFromProposalIncTxTable(tx string) error {

	p, err := DBGetProposalByIncTx(tx)
	if p != nil {
		// auto vote now (insert to vote):
		vote := &common.Vote{
			ProposalID:        p.ProposalID,
			Status:            common.StatusPdaoReadyForVote,
			ReShieldStatus:    common.StatusPending,
			Vote:              1,
			PropVoteSignature: p.PropVoteSignature,
			ReShieldSignature: p.ReShieldSignature,
			AutoVoted:         true,           // auto vote for owner of proposal.
			SubmitBurnTx:      p.SubmitBurnTx, // use proposal burn prv tx for tracking
		}
		// increase vor for:
		err = DBVoteAgainstPdaoProposal(vote.ProposalID)
		if err != nil {
			return err
		}
		return DBInsertVoteTable(vote)
	}
	return err

}

func DBGetPendingVote() ([]common.Vote, error) {
	result := []common.Vote{}
	filter := bson.M{"status": bson.M{operator.In: []string{common.StatusSubmitting}}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.Vote{}).SimpleFindWithCtx(ctx, &result, filter)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func DBGetReadyToVote() ([]common.Vote, error) {
	result := []common.Vote{}
	filter := bson.M{"status": bson.M{operator.In: []string{common.StatusPdaoReadyForVote}}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.Vote{}).SimpleFindWithCtx(ctx, &result, filter)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func DBGetVotingToReShield() ([]common.Vote, error) {
	result := []common.Vote{}
	filter := bson.M{"status": bson.M{operator.In: []string{common.StatusPdaOutchainTxSuccess, common.StatusPdaOutchainTxFailed}}, "reshield_status": bson.M{operator.Eq: common.StatusPending}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.Vote{}).SimpleFindWithCtx(ctx, &result, filter)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func DBUpdateVotingReshieldStatus(incTx string, status string) error {
	filter := bson.M{"submit_burn_tx": bson.M{operator.Eq: incTx}}
	update := bson.M{"$set": bson.M{"reshield_status": status}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.Vote{}).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

// create proposal with increase PID:
func DBCreateAProposalTable(data *common.Proposal) error {
	// get last ID:
	// try for 10 times:
	for i := 0; i < 10; i++ {
		var pid uint = 0
		listProd := DBListLimitProposalTable()
		if len(listProd) > 0 {
			pid = listProd[0].PID
		}
		pid += 1
		data.PID = pid
		err := DBInsertProposalTable(data)
		if err == nil {
			return nil
		}
	}
	return nil
}

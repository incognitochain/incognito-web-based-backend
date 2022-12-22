package database

import (
	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/kamva/mgm/v3"
)

// func DBInsertProposalTable(data common.Proposal) error {
// 	data.Creating()
// 	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
// 	_, err := mgm.Coll(&common.Proposal{}).InsertOne(ctx, data)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func DBUpdateProposalTable(data *common.Proposal) error {
	return mgm.Coll(&common.Proposal{}).Update(data)
}

func DBInsertProposalTable(data *common.Proposal) error {
	return mgm.Coll(data).Create(data)
}

// func DBGetProposalTable() (*common.Proposal, error) {
// 	// return mgm.Coll(data).Create(data)
// }

package database

import (
	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
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

func DBListProposalTable() []common.Proposal {

	p := []common.Proposal{}
	mgm.Coll(&common.Proposal{}).SimpleFind(&p, bson.M{})
	return p

}

// func DBGetProposalTable() (*common.Proposal, error) {
// 	// return mgm.Coll(data).Create(data)
// }

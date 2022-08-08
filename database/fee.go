package database

import (
	"context"
	"errors"
	"time"

	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBSaveFeetTable(data common.ExternalNetworksFeeData) error {
	data.Creating()
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.ExternalNetworksFeeData{}).InsertOne(ctx, data)
	if err != nil {
		return err
	}
	return nil
}

func DBRetrieveFeeTable() (*common.ExternalNetworksFeeData, error) {
	var result []common.ExternalNetworksFeeData
	filter := bson.M{}
	limit := int64(1)
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(10)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.ExternalNetworksFeeData{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
		Sort:  bson.D{{"created_at", 1}},
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, errors.New("fee table not found")
	}
	return &result[0], nil
}

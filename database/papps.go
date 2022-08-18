package database

import (
	"context"
	"log"
	"time"

	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/kamva/mgm/v3"
	"github.com/kamva/mgm/v3/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBRetrievePAppsByNetwork(network string) (*common.PAppsEndpointData, error) {
	var result common.PAppsEndpointData

	filter := bson.M{"network": bson.M{operator.Eq: network}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	dbresult := mgm.Coll(&common.PAppsEndpointData{}).FindOne(ctx, filter)
	if dbresult.Err() != nil {
		return nil, dbresult.Err()
	}

	if err := dbresult.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func DBRetrievePendingExternalTx(network string, offset, limit int64) ([]common.ExternalTxStatus, error) {
	startTime := time.Now()
	var result []common.ExternalTxStatus
	if limit == 0 {
		limit = int64(1000)
	}
	filter := bson.M{"status": bson.M{operator.Eq: common.StatusPending}, "network": bson.M{operator.Eq: network}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.ExternalTxStatus{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
		Skip:  &offset,
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("found %v ExternalTxStatus in %v", len(result), time.Since(startTime))
	return result, nil
}

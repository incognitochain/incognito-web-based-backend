package database

import (
	"context"
	"log"
	"time"

	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/kamva/mgm/v3"
	"github.com/kamva/mgm/v3/operator"
	"go.mongodb.org/mongo-driver/bson"
)

func DBGetBridgeNetworkInfos() ([]common.BridgeNetworkData, error) {
	startTime := time.Now()
	var result []common.BridgeNetworkData
	filter := bson.M{}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(10)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.BridgeNetworkData{}).SimpleFindWithCtx(ctx, &result, filter)
	if err != nil {
		return nil, err
	}
	log.Printf("found %v BridgeNetworkData in %v", len(result), time.Since(startTime))
	return result, nil
}

func DBGetBridgeNetworkInfo(network string) (*common.BridgeNetworkData, error) {
	startTime := time.Now()
	var result []common.BridgeNetworkData
	filter := bson.M{"network": bson.M{operator.Eq: network}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(10)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.BridgeNetworkData{}).SimpleFindWithCtx(ctx, &result, filter)
	if err != nil {
		return nil, err
	}
	log.Printf("found %v BridgeNetworkData in %v", len(result), time.Since(startTime))
	return &result[0], nil
}

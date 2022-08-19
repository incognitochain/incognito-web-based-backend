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

func DBUpdateExternalTxStatus(externalTx string, status string, errStr string) error {
	filter := bson.M{"txhash": bson.M{operator.Eq: externalTx}}
	update := bson.M{"$set": bson.M{"status": status, "error": errStr}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.ExternalTxStatus{}).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func DBUpdatePappTxStatus(incTx string, status string, errStr string) error {
	filter := bson.M{"inctxhash": bson.M{operator.Eq: incTx}}
	update := bson.M{"$set": bson.M{"status": status, "error": errStr}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.PappTxData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

func DBAddPappTxData(txdata common.PappTxData) error {
	_, err := mgm.Coll(&common.PappTxData{}).InsertOne(context.Background(), txdata)
	if err != nil {
		return err
	}
	return nil
}

func DBRetrievePendingPappTxs(offset, limit int64) ([]common.PappTxData, error) {
	startTime := time.Now()
	var result []common.PappTxData
	if limit == 0 {
		limit = int64(10000)
	}
	filter := bson.M{"status": bson.M{operator.Eq: common.StatusPending}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.PappTxData{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
		Skip:  &offset,
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("found %v PappTxData in %v", len(result), time.Since(startTime))
	return result, nil
}

func DBGetPappTxStatus(incTx string) (string, error) {
	var result common.PappTxData

	filter := bson.M{"inctxhash": bson.M{operator.Eq: incTx}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	dbresult := mgm.Coll(&common.PappTxData{}).FindOne(ctx, filter)
	if dbresult.Err() != nil {
		return "", dbresult.Err()
	}

	if err := dbresult.Decode(&result); err != nil {
		return "", err
	}

	return result.Status, nil
}

func DBGetPappContractData(network string, pappType int) (*common.PappContractData, error) {
	var result common.PappContractData

	filter := bson.M{"network": bson.M{operator.Eq: network}, "type": bson.M{operator.Eq: pappType}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	dbresult := mgm.Coll(&common.PappContractData{}).FindOne(ctx, filter)
	if dbresult.Err() != nil {
		return nil, dbresult.Err()
	}

	if err := dbresult.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil

}

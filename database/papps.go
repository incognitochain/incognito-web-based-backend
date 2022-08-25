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

func DBUpdateExternalTxStatusByIncTx(incTx string, status string, errStr string) error {
	filter := bson.M{"increquesttx": bson.M{operator.Eq: incTx}}
	update := bson.M{"$set": bson.M{"status": status, "error": errStr}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.ExternalTxStatus{}).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func DBSaveExternalTxStatus(txdata *common.ExternalTxStatus) error {
	filter := bson.M{"increquesttx": bson.M{operator.Eq: txdata.IncRequestTx}}
	update := bson.M{"$set": bson.M{"txhash": txdata.Txhash, "status": txdata.Status, "network": txdata.Network, "type": txdata.Type}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.ExternalTxStatus{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

func DBUpdatePappTxStatus(incTx string, status string, errStr string) error {
	filter := bson.M{"inctx": bson.M{operator.Eq: incTx}}
	update := bson.M{"$set": bson.M{"status": status, "error": errStr}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.PappTxData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

func DBSavePappTxData(txdata common.PappTxData) error {
	filter := bson.M{"inctx": bson.M{operator.Eq: txdata.IncTx}}
	update := bson.M{"$set": bson.M{"status": txdata.Status, "networks": txdata.Networks, "type": txdata.Type, "inctxdata": txdata.IncTxData, "feetoken": txdata.FeeToken, "feeamount": txdata.FeeAmount, "isunifiedtoken": txdata.IsUnifiedToken}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.PappTxData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

func DBRetrievePendingPappTxs(pappType int, offset, limit int64) ([]common.PappTxData, error) {
	startTime := time.Now()
	var result []common.PappTxData
	if limit == 0 {
		limit = int64(1000)
	}
	filter := bson.M{"status": bson.M{operator.In: []string{common.StatusPending, common.StatusExecuting}}, "type": bson.M{operator.Eq: pappType}}
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

func DBRetrieveAcceptedPappTxs(pappType int, offset, limit int64) ([]common.PappTxData, error) {
	startTime := time.Now()
	var result []common.PappTxData
	if limit == 0 {
		limit = int64(1000)
	}
	filter := bson.M{"status": bson.M{operator.In: []string{common.StatusAccepted}}, "type": bson.M{operator.Eq: pappType}}
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

	filter := bson.M{"inctx": bson.M{operator.Eq: incTx}}
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

func DBGetExternalTxStatusByIncTx(incTx string, network string) (string, error) {
	var result common.ExternalTxStatus

	filter := bson.M{"increquesttx": bson.M{operator.Eq: incTx}, "network": bson.M{operator.Eq: network}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	dbresult := mgm.Coll(&common.ExternalTxStatus{}).FindOne(ctx, filter)
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

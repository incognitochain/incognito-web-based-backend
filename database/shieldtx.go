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

func DBCreateShieldTxData(externalTx string, networkID int) error {
	var data common.ShieldTxData
	data.NetworkID = networkID
	data.ExternalTx = externalTx
	data.Creating()
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.ShieldTxData{}).InsertOne(ctx, data)
	if err != nil {
		return err
	}
	return nil
}

func DBRetrievePendingShieldTxs(offset, limit int64) ([]common.ShieldTxData, error) {
	startTime := time.Now()
	var result []common.ShieldTxData
	if limit == 0 {
		limit = int64(10000)
	}
	filter := bson.M{"status": bson.M{operator.Eq: common.StatusPending}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.ShieldTxData{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
		Skip:  &offset,
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("found %v ShieldTxData in %v", len(result), time.Since(startTime))
	return result, nil
}

func DBRetrieveFailedShieldTxs(offset, limit int64) ([]common.ShieldTxData, error) {
	startTime := time.Now()
	var result []common.ShieldTxData
	if limit == 0 {
		limit = int64(10000)
	}
	filter := bson.M{"status": bson.M{operator.Eq: common.StatusSubmitFailed}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.ShieldTxData{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
		Skip:  &offset,
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("found %v ShieldTxData in %v", len(result), time.Since(startTime))
	return result, nil
}

func DBRetrieveRejectedShieldTxs(offset, limit int64) ([]common.ShieldTxData, error) {
	startTime := time.Now()
	var result []common.ShieldTxData
	if limit == 0 {
		limit = int64(10000)
	}
	filter := bson.M{"status": bson.M{operator.Eq: common.StatusRejected}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.ShieldTxData{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
		Skip:  &offset,
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("found %v ShieldTxData in %v", len(result), time.Since(startTime))
	return result, nil
}

func DBGetShieldTxStatusByExternalTx(externalTx string, networkID int) (string, error) {
	var result common.ShieldTxData

	filter := bson.M{"externalTx": bson.M{operator.Eq: externalTx}, "networkid": bson.M{operator.Eq: networkID}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	dbresult := mgm.Coll(&common.ShieldTxData{}).FindOne(ctx, filter)
	if dbresult.Err() != nil {
		return "", dbresult.Err()
	}

	if err := dbresult.Decode(&result); err != nil {
		return "", err
	}

	return result.Status, nil
}

func DBUpdateShieldTxStatus(externalTx string, networkID int, status string, errStr string) error {
	filter := bson.M{"externalTx": bson.M{operator.Eq: externalTx}, "networkid": bson.M{operator.Eq: networkID}}
	update := bson.M{"$set": bson.M{"externalTx": externalTx, "networkid": networkID, "status": status, "error": errStr}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.ShieldTxData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

func DBUpdateShieldOnChainTxInfo(externalTx string, networkID int, paymentAddr string, incTx string, tokenID string, linkedTokenID string) error {
	filter := bson.M{"externalTx": bson.M{operator.Eq: externalTx}, "networkid": bson.M{operator.Eq: networkID}}
	update := bson.M{"$set": bson.M{"externalTx": externalTx, "networkid": networkID, "paymentaddress": paymentAddr, "inctx": incTx, "tokenid": tokenID, "utokenid": linkedTokenID}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.ShieldTxData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(false))
	if err != nil {
		return err
	}
	return nil
}

package database

import (
	"context"
	"log"
	"time"

	"github.com/incognitochain/coin-service/shared"
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
	return nil
}

func DBRetrievePendingShieldTxs(offset, limit int64) ([]common.ShieldTxData, error) {
	startTime := time.Now()
	var result []common.ShieldTxData
	if limit == 0 {
		limit = int64(10000)
	}
	filter := bson.M{"status": bson.M{operator.Eq: common.ShieldStatusPending}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*shared.DB_OPERATION_TIMEOUT)
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
	filter := bson.M{"status": bson.M{operator.Eq: common.ShieldStatusSubmitFailed}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*shared.DB_OPERATION_TIMEOUT)
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
	filter := bson.M{"status": bson.M{operator.Eq: common.ShieldStatusRejected}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*shared.DB_OPERATION_TIMEOUT)
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
	return result.Status, nil
}

func DBUpdateShieldTxStatus(externalTx string, networkID int, status string, err string) error {
	return nil
}

func DBUpdateShieldOnChainTxInfo(externalTx string, networkID int, paymentAddr string, incTx string, tokenID string, linkedTokenID string) error {
	return nil
}

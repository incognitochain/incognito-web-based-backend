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

func DBRetrievePendingRedepositExternalTx(offset, limit int64) ([]common.ExternalTxStatus, error) {
	startTime := time.Now()
	var result []common.ExternalTxStatus
	if limit == 0 {
		limit = int64(1000)
	}
	filter := bson.M{"will_redeposit": bson.M{operator.Eq: true}, "redeposit_submitted": bson.M{operator.Eq: false}}
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

func DBUpdateExternalTxWillRedeposit(externalTx string, willRedeposit bool) error {
	filter := bson.M{"txhash": bson.M{operator.Eq: externalTx}}
	update := bson.M{"$set": bson.M{"will_redeposit": willRedeposit}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.ExternalTxStatus{}).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func DBUpdateExternalTxSubmitedRedeposit(externalTx string, redepositSubmitted bool) error {
	filter := bson.M{"txhash": bson.M{operator.Eq: externalTx}}
	update := bson.M{"$set": bson.M{"redeposit_submitted": redepositSubmitted}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.ExternalTxStatus{}).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func DBUpdateExternalTxOtherInfo(externalTx string, otherInfo string) error {
	filter := bson.M{"txhash": bson.M{operator.Eq: externalTx}}
	update := bson.M{"$set": bson.M{"otherinfo": otherInfo}}
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
	update := bson.M{"$set": bson.M{"created_at": time.Now(), "txhash": txdata.Txhash, "status": txdata.Status, "network": txdata.Network, "type": txdata.Type}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.ExternalTxStatus{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

func DBRetrieveExternalTxByIncTxID(incTxID string) (*common.ExternalTxStatus, error) {
	var result common.ExternalTxStatus
	filter := bson.M{"increquesttx": bson.M{operator.Eq: incTxID}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.ExternalTxStatus{}).SimpleFindWithCtx(ctx, &result, filter)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

package database

import (
	"context"
	"log"
	"time"

	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/kamva/mgm/v3"
	"github.com/kamva/mgm/v3/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBSaveUnshieldTxData(txdata common.UnshieldTxData) (*primitive.ObjectID, error) {
	filter := bson.M{"inctx": bson.M{operator.Eq: txdata.IncTx}}
	update := bson.M{"$set": bson.M{"created_at": time.Now(), "status": txdata.Status, "networks": txdata.Networks, "inctxdata": txdata.IncTxData, "feetoken": txdata.FeeToken, "feeamount": txdata.FeeAmount, "tokenid": txdata.TokenID, "utokenid": txdata.UTokenID, "amount": txdata.Amount, "isunifiedtoken": txdata.IsUnifiedToken, "fee_refundota": txdata.FeeRefundOTA, "fee_refundaddress": txdata.FeeRefundAddress, "refundsubmitted": txdata.RefundSubmitted, "refundpfee": txdata.RefundPrivacyFee, "pfeeamount": txdata.PFeeAmount, "outchain_status": txdata.OutchainStatus, "useragent": txdata.UserAgent}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	result, err := mgm.Coll(&common.UnshieldTxData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return nil, err
	}
	docID := result.UpsertedID.(primitive.ObjectID)
	return &docID, nil
}
func DBRetrievePendingUnshieldTxs(offset, limit int64) ([]common.UnshieldTxData, error) {
	startTime := time.Now()
	var result []common.UnshieldTxData
	if limit == 0 {
		limit = int64(10000)
	}
	filter := bson.M{"status": bson.M{operator.Eq: common.StatusPending}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.UnshieldTxData{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
		Skip:  &offset,
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("found %v UnshieldTxData in %v", len(result), time.Since(startTime))
	return result, nil
}

func DBRetrieveFailedUnshieldTxs(offset, limit int64) ([]common.UnshieldTxData, error) {
	startTime := time.Now()
	var result []common.UnshieldTxData
	if limit == 0 {
		limit = int64(10000)
	}
	filter := bson.M{"status": bson.M{operator.Eq: common.StatusSubmitFailed}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.UnshieldTxData{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
		Skip:  &offset,
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("found %v UnshieldTxData in %v", len(result), time.Since(startTime))
	return result, nil
}

func DBRetrieveRejectedUnshieldTxs(offset, limit int64) ([]common.UnshieldTxData, error) {
	startTime := time.Now()
	var result []common.UnshieldTxData
	if limit == 0 {
		limit = int64(10000)
	}
	filter := bson.M{"status": bson.M{operator.Eq: common.StatusRejected}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.UnshieldTxData{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
		Skip:  &offset,
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("found %v UnshieldTxData in %v", len(result), time.Since(startTime))
	return result, nil
}

func DBGetUnshieldTxStatusByExternalTx(externalTx string, networkID int) (string, error) {
	result, err := DBGetUnshieldTxByExternalTx(externalTx, networkID)
	if err != nil {
		return "", err
	}
	return result.Status, nil
}
func DBGetUnshieldTxByExternalTx(externalTx string, networkID int) (*common.UnshieldTxData, error) {
	var result common.UnshieldTxData

	filter := bson.M{"externaltx": bson.M{operator.Eq: externalTx}, "networkid": bson.M{operator.Eq: networkID}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	dbresult := mgm.Coll(&common.UnshieldTxData{}).FindOne(ctx, filter)
	if dbresult.Err() != nil {
		return nil, dbresult.Err()
	}

	if err := dbresult.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func DBGetUnshieldTxStatusByIncTx(incTx string) (string, error) {
	result, err := DBGetUnshieldTxByIncTx(incTx)
	if err != nil {
		return "", err
	}
	return result.Status, nil
}
func DBGetUnshieldTxByIncTx(incTx string) (*common.UnshieldTxData, error) {
	var result common.UnshieldTxData

	filter := bson.M{"inctx": bson.M{operator.Eq: incTx}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	dbresult := mgm.Coll(&common.UnshieldTxData{}).FindOne(ctx, filter)
	if dbresult.Err() != nil {
		return nil, dbresult.Err()
	}

	if err := dbresult.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func DBUpdateUnshieldExternalTxStatus(externalTx string, networkID int, status string, errStr string) error {
	filter := bson.M{"externaltx": bson.M{operator.Eq: externalTx}, "networkid": bson.M{operator.Eq: networkID}}
	update := bson.M{"$set": bson.M{"externaltx": externalTx, "networkid": networkID, "status": status, "error": errStr}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.UnshieldTxData{}).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func DBUpdateUnshieldOnChainTxInfo(networkID int, paymentAddr string, incTx string, utokenID string, linkedTokenID string) error {
	filter := bson.M{"inctx": bson.M{operator.Eq: incTx}, "networkid": bson.M{operator.Eq: networkID}}
	update := bson.M{"$set": bson.M{"networkid": networkID, "paymentaddress": paymentAddr, "inctx": incTx, "tokenid": linkedTokenID, "utokenid": utokenID}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.UnshieldTxData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

func DBUpdateUnshieldTxStatus(incTx string, status string, errStr string) error {
	filter := bson.M{"inctx": bson.M{operator.Eq: incTx}}
	update := bson.M{"$set": bson.M{"status": status, "error": errStr}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.UnshieldTxData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(false))
	if err != nil {
		return err
	}
	return nil
}

func DBGetUnshieldTxPendingOutchainSubmit(offset, limit int64) ([]common.UnshieldTxData, error) {
	startTime := time.Now()
	var result []common.UnshieldTxData
	if limit == 0 {
		limit = int64(1000)
	}
	filter := bson.M{"outchain_status": bson.M{operator.In: []string{common.StatusSubmitting, common.StatusPending, common.StatusWaiting, common.StatusSubmitFailed}}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.UnshieldTxData{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
		Skip:  &offset,
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("found %v UnshieldTxData in %v", len(result), time.Since(startTime))
	return result, nil
}

func DBUpdateUnshieldTxSubmitOutchainStatus(incTx string, status string) error {
	filter := bson.M{"inctx": bson.M{operator.Eq: incTx}}
	update := bson.M{"$set": bson.M{"outchain_status": status}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.UnshieldTxData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(false))
	if err != nil {
		return err
	}
	return nil
}

func DBGetUnshieldTxDataByStatusAndShardID(status string, shardid int, offset, limit int64) ([]common.UnshieldTxData, error) {
	startTime := time.Now()
	var result []common.UnshieldTxData
	if limit == 0 {
		limit = int64(1000)
	}
	filter := bson.M{"status": bson.M{operator.Eq: status}, "shardid": bson.M{operator.Eq: shardid}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.UnshieldTxData{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
		Skip:  &offset,
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("found %v UnshieldTxData in %v", len(result), time.Since(startTime))
	return result, nil
}

func DBGetUnshieldTxDataByStatus(status string, offset, limit int64) ([]common.UnshieldTxData, error) {
	startTime := time.Now()
	var result []common.UnshieldTxData
	if limit == 0 {
		limit = int64(1000)
	}
	filter := bson.M{"status": bson.M{operator.Eq: status}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.UnshieldTxData{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
		Skip:  &offset,
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("found %v UnshieldTxData in %v", len(result), time.Since(startTime))
	return result, nil
}

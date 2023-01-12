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

func DBUpdatePappTxStatus(incTx string, status string, errStr string) error {
	filter := bson.M{"inctx": bson.M{operator.Eq: incTx}}
	update := bson.M{"$set": bson.M{"status": status, "error": errStr}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.PappTxData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(false))
	if err != nil {
		return err
	}
	return nil
}

func DBUpdatePappTxSubmitOutchainStatus(incTx string, status string) error {
	filter := bson.M{"inctx": bson.M{operator.Eq: incTx}}
	update := bson.M{"$set": bson.M{"outchain_status": status}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.PappTxData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(false))
	if err != nil {
		return err
	}
	return nil
}

func DBGetPappWaitingSubmitOutchain(offset, limit int64) ([]common.PappTxData, error) {
	startTime := time.Now()
	var result []common.PappTxData
	if limit == 0 {
		limit = int64(1000)
	}
	filter := bson.M{"outchain_status": bson.M{operator.In: []string{common.StatusWaiting, common.StatusSubmitting}}}
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

func DBSavePappTxData(txdata common.PappTxData) error {
	filter := bson.M{"inctx": bson.M{operator.Eq: txdata.IncTx}}
	update := bson.M{"$set": bson.M{"created_at": time.Now(), "status": txdata.Status, "networks": txdata.Networks, "type": txdata.Type, "inctxdata": txdata.IncTxData, "feetoken": txdata.FeeToken, "feeamount": txdata.FeeAmount, "burnttoken": txdata.BurntToken, "burntamount": txdata.BurntAmount, "pappswapinfo": txdata.PappSwapInfo, "isunifiedtoken": txdata.IsUnifiedToken, "fee_refundota": txdata.FeeRefundOTA, "fee_refundaddress": txdata.FeeRefundAddress, "refundsubmitted": txdata.RefundSubmitted, "refundpfee": txdata.RefundPrivacyFee, "pfeeamount": txdata.PFeeAmount, "outchain_status": txdata.OutchainStatus, "useragent": txdata.UserAgent}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.PappTxData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}
	return nil
}

func DBUpdatePappRefundPFee(incTx string, refundpfee bool) error {
	filter := bson.M{"inctx": bson.M{operator.Eq: incTx}}
	update := bson.M{"$set": bson.M{"refundpfee": refundpfee}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.PappTxData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(false))
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

func DBGetPappTxData(incTx string) (*common.PappTxData, error) {
	var result common.PappTxData

	filter := bson.M{"inctx": bson.M{operator.Eq: incTx}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	dbresult := mgm.Coll(&common.PappTxData{}).FindOne(ctx, filter)
	if dbresult.Err() != nil {
		return nil, dbresult.Err()
	}

	if err := dbresult.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func DBGetPappTxDataByStatusAndShardID(status string, shardid int, offset, limit int64) ([]common.PappTxData, error) {
	startTime := time.Now()
	var result []common.PappTxData
	if limit == 0 {
		limit = int64(1000)
	}
	filter := bson.M{"status": bson.M{operator.Eq: status}, "shardid": bson.M{operator.Eq: shardid}}
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

func DBGetPappTxDataByStatus(status string, offset, limit int64) ([]common.PappTxData, error) {
	startTime := time.Now()
	var result []common.PappTxData
	if limit == 0 {
		limit = int64(1000)
	}
	filter := bson.M{"status": bson.M{operator.Eq: status}}
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

func DBGetPappTxPendingOutchainSubmit(offset, limit int64) ([]common.PappTxData, error) {
	startTime := time.Now()
	var result []common.PappTxData
	if limit == 0 {
		limit = int64(1000)
	}
	filter := bson.M{"outchain_status": bson.M{operator.In: []string{common.StatusSubmitting, common.StatusPending, common.StatusWaiting, common.StatusSubmitFailed}}}
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

func DBGetExternalTxByIncTx(incTx string, network string) (*common.ExternalTxStatus, error) {
	var result common.ExternalTxStatus

	filter := bson.M{"increquesttx": bson.M{operator.Eq: incTx}, "network": bson.M{operator.Eq: network}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	dbresult := mgm.Coll(&common.ExternalTxStatus{}).FindOne(ctx, filter)
	if dbresult.Err() != nil {
		return nil, dbresult.Err()
	}

	if err := dbresult.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func DBGetPappVaultData(network string, pappType int) (*common.PappVaultData, error) {
	var result common.PappVaultData

	filter := bson.M{"network": bson.M{operator.Eq: network}, "type": bson.M{operator.Eq: pappType}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	dbresult := mgm.Coll(&common.PappVaultData{}).FindOne(ctx, filter)
	if dbresult.Err() != nil {
		return nil, dbresult.Err()
	}

	if err := dbresult.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func DBGetPappSupportedToken() ([]common.PappSupportedTokenData, error) {
	var result []common.PappSupportedTokenData
	filter := bson.M{}
	err := mgm.Coll(&common.PappSupportedTokenData{}).SimpleFindWithCtx(context.Background(), &result, filter)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func DBRetrievePendingDexTxs(offset, limit int64) ([]common.DexSwapTrackData, error) {
	startTime := time.Now()
	var result []common.DexSwapTrackData
	if limit == 0 {
		limit = int64(1000)
	}
	filter := bson.M{"status": bson.M{operator.In: []string{common.StatusPending, common.StatusExecuting}}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.DexSwapTrackData{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
		Skip:  &offset,
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("found %v DexSwapTrackData in %v", len(result), time.Since(startTime))
	return result, nil
}

func DBUpdateDexSwapTxStatus(incTx string, status string) error {
	filter := bson.M{"inctx": bson.M{operator.Eq: incTx}}
	update := bson.M{"$set": bson.M{"status": status}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.DexSwapTrackData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(false))
	if err != nil {
		return err
	}
	return nil
}

func DBSaveDexSwapTxData(txdata common.DexSwapTrackData) error {
	var doc interface{}
	doc = txdata
	_, err := mgm.Coll(&common.DexSwapTrackData{}).InsertOne(context.Background(), doc)
	if err != nil {
		return err
	}
	return nil
}

func DBDeleteDexSwap(txList []string) error {
	filter := bson.M{"inctx": bson.M{operator.In: txList}}
	_, err := mgm.Coll(&common.DexSwapTrackData{}).DeleteMany(context.Background(), filter)
	if err != nil {
		return err
	}
	return nil
}

func DBGetPappAPIKey(papp string) (string, error) {
	var result common.PAppAPIKeyData
	filter := bson.M{"app": bson.M{operator.Eq: papp}}
	dbresult := mgm.Coll(&common.PAppAPIKeyData{}).FindOne(context.Background(), filter)
	if dbresult.Err() != nil {
		return "", dbresult.Err()
	}
	if err := dbresult.Decode(&result); err != nil {
		return "", err
	}
	return result.Key, nil
}

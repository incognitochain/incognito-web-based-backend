package database

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/kamva/mgm/v3"
	"github.com/kamva/mgm/v3/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBInsertInterswapTxData(txdata common.InterSwapTxData) error {
	bytes, _ := json.Marshal(txdata)
	log.Printf("DBInsertInterswapTxData %+v", string(bytes))
	var doc interface{}
	doc = txdata
	_, err := mgm.Coll(&common.InterSwapTxData{}).InsertOne(context.Background(), doc)
	if err != nil {
		log.Printf("DBInsertInterswapTxData err %+v", err)
		return err
	}
	return nil
}

func DBSaveInterSwapTxData(txdata common.InterSwapTxData) (*primitive.ObjectID, error) {
	bytes, _ := json.Marshal(txdata)
	log.Printf("DBSaveInterSwapTxData %+v", string(bytes))
	filter := bson.M{"txid": bson.M{operator.Eq: txdata.TxID}}
	update := bson.M{"$set": bson.M{
		"created_at":              time.Now(),
		"txraw":                   txdata.TxRaw,
		"fromamount":              txdata.FromAmount,
		"fromtoken":               txdata.FromToken,
		"totoken":                 txdata.ToToken,
		"midtoken":                txdata.MidToken,
		"pathtype":                txdata.PathType,
		"final_minacceptedamount": txdata.FinalMinExpectedAmt,
		"slippage":                txdata.Slippage,
		"ota_refundfee":           txdata.OTARefundFee,
		"ota_refund":              txdata.OTARefund,
		"ota_fromtoken":           txdata.OTAFromToken,
		"ota_totoken":             txdata.OTAToToken,
		"withdrawaddress":         txdata.WithdrawAddress,
		"addon_txid":              txdata.AddOnTxID,
		"txidrefund":              txdata.TxIDRefund,
		"txidresponse":            txdata.TxIDResponse,
		"amountresponse":          txdata.AmountResponse,
		"tokenresponse":           txdata.TokenResponse,
		"shardid":                 txdata.ShardID,
		"txidoutchain":            txdata.TxIDOutchain,
		"papp_name":               txdata.PAppName,
		"papp_network":            txdata.PAppNetwork,
		"papp_contract":           txdata.PAppContract,
		"status":                  txdata.Status,
		"statusstr":               txdata.StatusStr,
		"useragent":               txdata.UserAgent,
		"error":                   txdata.Error,
	}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	result, err := mgm.Coll(&common.InterSwapTxData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return nil, err
	}
	docID := result.UpsertedID.(primitive.ObjectID)
	return &docID, nil
}

func DBUpdateInterswapTxStatus(txID string, status int, statusStr string, errStr string) error {
	filter := bson.M{"txid": bson.M{operator.Eq: txID}}
	update := bson.M{"$set": bson.M{"status": status, "statusstr": statusStr, "error": errStr}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.InterSwapTxData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(false))
	if err != nil {
		return err
	}
	return nil
}

func DBUpdateInterswapTxAddOnTxID(txID string, addOnTxID string) error {
	filter := bson.M{"txid": bson.M{operator.Eq: txID}}
	update := bson.M{"$set": bson.M{"addon_txid": addOnTxID}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.InterSwapTxData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(false))
	if err != nil {
		return err
	}
	return nil
}

func DBUpdateInterswapTxRefundTxID(txID string, refundTxID string) error {
	filter := bson.M{"txid": bson.M{operator.Eq: txID}}
	update := bson.M{"$set": bson.M{"txidrefund": refundTxID}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.InterSwapTxData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(false))
	if err != nil {
		return err
	}
	return nil
}

func DBUpdateInterswapTxInfo(txID string, updateInfo map[string]interface{}) error {
	filter := bson.M{"txid": bson.M{operator.Eq: txID}}
	m := bson.M{}
	for k, v := range updateInfo {
		m[k] = v
	}
	update := bson.M{"$set": m}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.InterSwapTxData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(false))
	if err != nil {
		return err
	}
	return nil
}

func DBRetrieveInterswapTxsByStatus(status []int, offset, limit int64) ([]common.InterSwapTxData, error) {
	startTime := time.Now()
	var result []common.InterSwapTxData
	if limit == 0 {
		limit = int64(1000)
	}
	filter := bson.M{"status": bson.M{operator.In: status}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.InterSwapTxData{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
		Skip:  &offset,
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("found %v InterSwapTxData in %v", len(result), time.Since(startTime))
	return result, nil
}

func DBRetrieveInterswapTxByTxID(txID string) (*common.InterSwapTxData, error) {
	result := common.InterSwapTxData{}

	filter := bson.M{"txid": bson.M{operator.Eq: txID}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	dbresult := mgm.Coll(&common.InterSwapTxData{}).FindOne(ctx, filter)
	if dbresult.Err() != nil {
		return nil, dbresult.Err()
	}

	if err := dbresult.Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

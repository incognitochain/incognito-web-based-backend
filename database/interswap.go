package database

import (
	"context"
	"time"

	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/kamva/mgm/v3"
	"github.com/kamva/mgm/v3/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBSaveInterSwapTxData(txdata common.InterSwapTxData) (*primitive.ObjectID, error) {
	filter := bson.M{"txid": bson.M{operator.Eq: txdata.TxID}}
	update := bson.M{"$set": bson.M{
		"created_at": time.Now(),
		"txraw":      txdata.TxRaw,
		// "tx"
		"addon_swapinfo": txdata.AddOnSwapInfo,
		"ota_refundfee":  txdata.OTARefundFee,
		"ota_fromtoken":  txdata.OTAFromToken,
		"ota_totoken":    txdata.OTAToToken,
		"status":         txdata.Status,
		"statusstr":      txdata.StatusStr,
		"useragent":      txdata.UserAgent,
		"error":          txdata.Error,
	}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	result, err := mgm.Coll(&common.InterSwapTxData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return nil, err
	}
	docID := result.UpsertedID.(primitive.ObjectID)
	return &docID, nil
}

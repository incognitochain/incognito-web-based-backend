package database

import (
	"context"
	"errors"
	"time"

	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/kamva/mgm/v3"
	"github.com/kamva/mgm/v3/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBSaveFeetTable(data common.ExternalNetworksFeeData) error {
	data.Creating()
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.ExternalNetworksFeeData{}).InsertOne(ctx, data)
	if err != nil {
		return err
	}
	return nil
}

func DBRetrieveFeeTable() (*common.ExternalNetworksFeeData, error) {
	var result []common.ExternalNetworksFeeData
	filter := bson.M{}
	limit := int64(1)
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(10)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.ExternalNetworksFeeData{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
		Sort:  bson.D{{"created_at", -1}},
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, errors.New("fee table not found")
	}
	return &result[0], nil
}

func DBRetrieveFeesTable(limit int64) ([]common.ExternalNetworksFeeData, error) {
	var result []common.ExternalNetworksFeeData
	filter := bson.M{}
	if limit == 0 {
		limit = 1
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(10)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.ExternalNetworksFeeData{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
		Sort:  bson.D{{"created_at", -1}},
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, errors.New("fee table not found")
	}
	return result, nil
}

func DBCreateRefundFeeRecord(data common.RefundFeeData) error {
	data.Creating()
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.RefundFeeData{}).InsertOne(ctx, data)
	if err != nil {
		return err
	}
	return nil
}

func DBUpdateRefundFeeRefundTx(refundtx, incReqTx, status, errStr string) error {
	filter := bson.M{"increquesttx": bson.M{operator.Eq: incReqTx}}
	update := bson.M{"$set": bson.M{"status": status, "refundtx": refundtx, "error": errStr, "updated_at": time.Now()}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.RefundFeeData{}).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func DBUpdatePappRefund(incReqTx string, refundsubmitted bool) error {
	filter := bson.M{"inctx": bson.M{operator.Eq: incReqTx}}
	update := bson.M{"$set": bson.M{"refundsubmitted": refundsubmitted}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.PappTxData{}).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

// func DBGetWaitingFeeRefund(limit int64) ([]common.RefundFeeData, error) {
// 	var result []common.RefundFeeData
// 	filter := bson.M{"status": bson.M{operator.Eq: common.StatusWaiting}}
// 	if limit == 0 {
// 		limit = 100
// 	}
// 	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*DB_OPERATION_TIMEOUT)
// 	err := mgm.Coll(&common.RefundFeeData{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
// 		Sort:  bson.D{{"created_at", -1}},
// 		Limit: &limit,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return result, nil
// }

func DBGetPappTxNeedFeeRefund(limit int64) ([]common.PappTxData, error) {
	var result []common.PappTxData
	filter := bson.M{"status": bson.M{operator.Eq: common.StatusRejected}, "refundsubmitted": bson.M{operator.Eq: false}}
	if limit == 0 {
		limit = 100
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.PappTxData{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
		Sort:  bson.D{{"created_at", -1}},
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func DBGetTxFeeRefundByReq(incReqTx string) (*common.RefundFeeData, error) {
	var result common.RefundFeeData

	filter := bson.M{"increquesttx": bson.M{operator.Eq: incReqTx}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	dbresult := mgm.Coll(&common.RefundFeeData{}).FindOne(ctx, filter)
	if dbresult.Err() != nil {
		return nil, dbresult.Err()
	}

	if err := dbresult.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func DBGetPendingFeeRefundTx(limit int64) ([]common.RefundFeeData, error) {
	var result []common.RefundFeeData
	filter := bson.M{"status": bson.M{operator.In: []string{common.StatusPending, common.StatusWaiting}}}
	if limit == 0 {
		limit = 100
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.RefundFeeData{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
		Sort:  bson.D{{"created_at", -1}},
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

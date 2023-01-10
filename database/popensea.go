package database

import (
	"context"
	"strings"
	"time"

	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/papps/popensea"
	"github.com/kamva/mgm/v3"
	"github.com/kamva/mgm/v3/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBSaveCollectionsInfo(list []popensea.CollectionDetail) error {
	for _, collection := range list {
		filter := bson.M{"address": bson.M{operator.Eq: strings.ToLower(collection.PrimaryAssetContracts[0].Address)}}
		update := bson.M{"$set": bson.M{"address": strings.ToLower(collection.PrimaryAssetContracts[0].Address), "name": collection.Name, "detail": collection, "updated_at": time.Now()}}
		ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
		_, err := mgm.Coll(&common.OpenseaCollectionData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
		if err != nil {
			return err
		}
	}
	return nil
}

func DBGetCollectionsInfo(address string) (*common.OpenseaCollectionData, error) {
	var result common.OpenseaCollectionData
	filter := bson.M{"address": bson.M{operator.Eq: strings.ToLower(address)}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	dbresult := mgm.Coll(&common.OpenseaCollectionData{}).FindOne(ctx, filter)
	if dbresult.Err() != nil {
		return nil, dbresult.Err()
	}

	if err := dbresult.Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func DBSaveNFTDetail(list []popensea.NFTDetail) error {
	for _, nft := range list {
		uid := nft.AssetContract.Address + "-" + nft.TokenID
		filter := bson.M{"uid": bson.M{operator.Eq: uid}}
		update := bson.M{"$set": bson.M{"uid": uid, "address": nft.AssetContract.Address, "token_id": nft.TokenID, "name": nft.AssetContract.Name, "detail": nft, "updated_at": time.Now()}}
		ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
		_, err := mgm.Coll(&common.OpenseaAssetData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
		if err != nil {
			return err
		}
	}
	return nil
}

func DBGetNFTDetail(address string, nftid string) (*common.OpenseaAssetData, error) {
	var result common.OpenseaAssetData
	uid := address + "-" + nftid
	filter := bson.M{"uid": bson.M{operator.Eq: uid}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	dbresult := mgm.Coll(&common.OpenseaAssetData{}).FindOne(ctx, filter)
	if dbresult.Err() != nil {
		return nil, dbresult.Err()
	}

	if err := dbresult.Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func DBGetCollectionNFTs(address string, limit, offset int64) ([]common.OpenseaAssetData, error) {
	var result []common.OpenseaAssetData
	if limit == 0 {
		limit = int64(1000)
	}
	filter := bson.M{"address": bson.M{operator.Eq: address}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.OpenseaAssetData{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
		Skip:  &offset,
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func DBGetDefaultCollectionList() ([]common.OpenseaDefaultCollectionData, error) {
	var result []common.OpenseaDefaultCollectionData
	limit := int64(1000)
	filter := bson.M{"verify": bson.M{operator.Eq: true}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.OpenseaDefaultCollectionData{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func DBGetPendingOpenseaOffer() ([]common.OpenseaOfferData, error) {
	var result []common.OpenseaOfferData
	limit := int64(1000)
	filter := bson.M{"status": bson.M{operator.Eq: popensea.OfferStatusPending}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(limit)*DB_OPERATION_TIMEOUT)
	err := mgm.Coll(&common.OpenseaOfferData{}).SimpleFindWithCtx(ctx, &result, filter, &options.FindOptions{
		Limit: &limit,
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func DBGetOpenseaOfferByOfferTx(txhash string) (*common.OpenseaOfferData, error) {
	var result common.OpenseaOfferData
	filter := bson.M{"offer_tx_inc": bson.M{operator.Eq: txhash}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(DB_OPERATION_TIMEOUT)*time.Second)

	dbresult := mgm.Coll(&common.OpenseaOfferData{}).FindOne(ctx, filter)
	if dbresult.Err() != nil {
		return nil, dbresult.Err()
	}
	if err := dbresult.Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func DBInsertOpenseaOfferData(data *common.OpenseaOfferData) error {
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(DB_OPERATION_TIMEOUT)*time.Second)
	_, err := mgm.Coll(&common.OpenseaOfferData{}).InsertOne(ctx, data)
	if err != nil {
		return err
	}
	return nil
}

func DBUpdateOpenseaOfferStatus(incTx string, status string) error {
	filter := bson.M{"offer_tx_inc": bson.M{operator.Eq: incTx}}
	update := bson.M{"$set": bson.M{"status": status}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.OpenseaOfferData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(false))
	if err != nil {
		return err
	}
	return nil
}

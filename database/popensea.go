package database

import (
	"context"
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
		filter := bson.M{"address": bson.M{operator.Eq: collection.PrimaryAssetContracts[0].Address}}
		update := bson.M{"$set": bson.M{"address": collection.PrimaryAssetContracts[0].Address, "name": collection.Name, "detail": collection, "updated_at": time.Now()}}
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
	filter := bson.M{"address": bson.M{operator.Eq: address}}
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

func DBGetCollectionNFTs(address string) (*common.OpenseaAssetData, error) {
	var result common.OpenseaAssetData
	filter := bson.M{"address": bson.M{operator.Eq: address}}
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

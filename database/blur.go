package database

import (
	"context"
	"fmt"
	"time"

	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/papps/pblur"
	"github.com/kamva/mgm/v3"
	"github.com/kamva/mgm/v3/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBBlurSaveCollection(list []pblur.CollectionDetail) error {
	for _, collection := range list {
		filter := bson.M{"collection_slug": bson.M{operator.Eq: collection.CollectionSlug}}
		update := bson.M{"$set": bson.M{
			"collection_slug":  collection.CollectionSlug,
			"contract_address": collection.ContractAddress,
			"image_url":        collection.ImageUrl,
			"name":             collection.Name,
			"total_supply":     collection.TotalSupply,
			"number_owners":    collection.NumberOwners,
			"floor_price":      collection.FloorPrice.Amount,

			"floor_price_one_day":    collection.FloorPriceOneDay.Amount,
			"floor_price_one_week":   collection.FloorPriceOneWeek.Amount,
			"volume_fifteen_minutes": collection.VolumeFifteenMinutes.Amount,

			"volume_one_day":  collection.VolumeOneDay.Amount,
			"volume_one_week": collection.VolumeOneWeek.Amount,

			"best_collection_bid":        collection.BestCollectionBid.Amount,
			"total_collection_bid_value": collection.TotalCollectionBidValue.Amount,
			"trait_frequencies":          collection.TraitFrequencies,

			"updated_at": time.Now()}}

		ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
		_, err := mgm.Coll(&common.BlurCollectionData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
		if err != nil {
			return err
		}
	}
	return nil
}

func DBBlurSaveNFTDetail(contractAddress string, list []pblur.NFTDetail) error {
	for _, nft := range list {
		uid := contractAddress + "-" + nft.TokenID
		filter := bson.M{"uid": bson.M{operator.Eq: uid}}
		update := bson.M{"$set": bson.M{"uid": uid, "contract_address": contractAddress, "token_id": nft.TokenID, "name": nft.Name, "price": nft.Price.Amount, "detail": nft, "updated_at": time.Now()}}
		ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
		_, err := mgm.Coll(&common.BlurAssetData{}).UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
		if err != nil {
			return err
		}
	}
	return nil
}

func DBBlurGetNFTDetail(address string, nftid string) (*common.BlurAssetData, error) {
	var result common.BlurAssetData
	uid := address + "-" + nftid
	filter := bson.M{"uid": bson.M{operator.Eq: uid}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	dbresult := mgm.Coll(&common.BlurAssetData{}).FindOne(ctx, filter)
	if dbresult.Err() != nil {
		return nil, dbresult.Err()
	}

	if err := dbresult.Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// get by id
func DBBlurGetCollectionDetail(slug string) (*common.BlurCollectionData, error) {

	p := &common.BlurCollectionData{}
	filter := bson.M{"collection_slug": bson.M{operator.Eq: slug}}
	err := mgm.Coll(p).First(filter, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func DBBlurGetCollectionNFTs(address string, filterObj *common.Filter) ([]common.BlurAssetData, error) {
	var result []common.BlurAssetData

	limit := int64(filterObj.Limit)
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	filterObj.Limit = int(limit)

	page := int64(filterObj.Page)

	if page <= 0 {
		page = 1
	}

	offset := page*limit - limit
	filterObj.Offset = int(offset)

	filterObj.Page = int(page)

	sort := "rarity_rank"
	order := -1

	if len(filterObj.Sort) > 0 {
		sort = filterObj.Sort
	}

	if filterObj.Order == "asc" {
		order = 1
	}
	query := filterObj.Query

	fmt.Println("query: ", query)
	fmt.Println("sort: ", sort)
	fmt.Println("order: ", order)
	fmt.Println("page: ", page)
	fmt.Println("offset: ", offset)
	fmt.Println("limit: ", limit)

	var err error
	var filter interface{}

	if len(filterObj.Query) > 2 {
		filter = bson.M{"contract_address": bson.M{operator.Eq: address}, "name": bson.M{operator.Regex: query, "$options": "i"}}
	} else {
		filter = bson.M{"contract_address": bson.M{operator.Eq: address}}
	}
	err = mgm.Coll(&common.BlurAssetData{}).SimpleFind(&result, filter, &options.FindOptions{
		Limit: &limit,
		Skip:  &offset,
		Sort:  bson.D{{sort, order}},
	})
	if err != nil {
		return result, err
	}
	return result, nil
}

func DBBlurGetCollectionList(filterObj *common.Filter) ([]common.BlurCollectionData, error) {
	var result []common.BlurCollectionData

	limit := int64(filterObj.Limit)
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	filterObj.Limit = int(limit)

	page := int64(filterObj.Page)

	if page <= 0 {
		page = 1
	}

	offset := page*limit - limit
	filterObj.Offset = int(offset)

	// page -= 1

	filterObj.Page = int(page)

	sort := "volume_one_day"
	order := -1

	if len(filterObj.Sort) > 0 {
		sort = filterObj.Sort
	}

	if filterObj.Order == "asc" {
		order = 1
	}
	query := filterObj.Query

	fmt.Println("query: ", query)
	fmt.Println("sort: ", sort)
	fmt.Println("order: ", order)
	fmt.Println("page: ", page)
	fmt.Println("offset: ", offset)
	fmt.Println("limit: ", limit)

	var err error
	var filter interface{}

	if len(filterObj.Query) > 2 {
		//filter = bson.D{{operator.Text, bson.D{{"$search", "+\"" + query + "\""}}}}

		// filter = bson.D{{operator.Text, bson.D{{"$search", "/" + query + "/"}}}}

		// filter = bson.M{"name": "/" + query + "/"}

		// filter = bson.M{"name": bson.M{operator.Regex: "/.*digi.*/"}}

		// filter = bson.D{{operator.Regex, "/Ape/"}}
		// filter = bson.M{"name": bson.M{operator.Expr: "/.*Ape.*/"}}
		// filter = bson.D{{operator.Text, bson.D{{"$search", "/" + query + "/"}}}}

		// filter = bson.M{"name": bson.M{operator.Eq: "/.*Ape.*/"}}

		// filter = bson.M{"name": "/Club/"}
		//db.users.find({ "name": { "$regex": "m", "$options": "i" } })
		filter = bson.M{"name": bson.M{operator.Regex: query, "$options": "i"}}

		// filter = bson.D{{"name", primitive.Regex{Pattern: "ape", Options: "i"}}}

	} else {
		filter = bson.M{}
	}
	err = mgm.Coll(&common.BlurCollectionData{}).SimpleFind(&result, filter, &options.FindOptions{
		Limit: &limit,
		Skip:  &offset,
		Sort:  bson.D{{sort, order}},
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

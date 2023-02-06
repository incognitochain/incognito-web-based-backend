package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/kamva/mgm/v3"
	"github.com/kamva/mgm/v3/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBPNftInsertPNftAssetDataTable(data *common.PNftAssetData) error {
	return mgm.Coll(data).Create(data)
}
func DBPNftInsertPNftCollectionDataTable(data *common.PNftCollectionData) error {
	return mgm.Coll(data).Create(data)
}

func DBPNftInsertListNftCacheTable(data *common.ListNftCache) error {
	return mgm.Coll(data).Create(data)
}

// get nft list cache by address
func DBPNftGetListNftCacheTableByAddress(address string) (*common.ListNftCache, error) {

	p := &common.ListNftCache{}
	filter := bson.M{"address": bson.M{operator.Eq: strings.ToLower(address)}}
	err := mgm.Coll(p).First(filter, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}
func DBPNftGetNFTDetail(address string, nftid string) (*common.PNftAssetData, error) {
	var result common.PNftAssetData
	uid := strings.ToLower(address) + "-" + strings.ToLower(nftid)
	filter := bson.M{"uid": bson.M{operator.Eq: uid}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	dbresult := mgm.Coll(&common.PNftAssetData{}).FindOne(ctx, filter)
	if dbresult.Err() != nil {
		return nil, dbresult.Err()
	}

	if err := dbresult.Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

func DBPNftGetNFTDetailByIDs(address string, nftids []string) ([]common.PNftAssetData, error) {
	var result []common.PNftAssetData

	fmt.Println("address: ", address)
	fmt.Println("nftids: ", nftids)

	filter := bson.M{"contract_address": bson.M{operator.Eq: strings.ToLower(address)}, "token_id": bson.M{operator.All: nftids}}
	err := mgm.Coll(&common.PNftAssetData{}).SimpleFind(&result, filter, &options.FindOptions{})
	if err != nil {
		return result, err
	}
	return result, nil
}

func DBPNftListingAsset(address, nftID string, listing bool) error {

	uid := strings.ToLower(address) + "-" + strings.ToLower(nftID)
	filter := bson.M{"uid": bson.M{operator.Eq: uid}}

	update := bson.M{"$set": bson.M{"listing": listing}}
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(1)*DB_OPERATION_TIMEOUT)
	_, err := mgm.Coll(&common.PNftAssetData{}).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func DBPNftUpdateAssetTable(data *common.PNftAssetData) error {
	return mgm.Coll(&common.PNftAssetData{}).Update(data)
}

// get collection by slug
func DBPNftGetCollectionDetail(slug string) (*common.PNftCollectionData, error) {

	p := &common.PNftCollectionData{}
	filter := bson.M{"collection_slug": bson.M{operator.Eq: slug}}
	err := mgm.Coll(p).First(filter, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// get collection by contract address
func DBBlurGetCollectionByAddressDetail(address string) (*common.PNftCollectionData, error) {

	p := &common.PNftCollectionData{}
	filter := bson.M{"contract_address": bson.M{operator.Eq: strings.ToLower(address)}}
	err := mgm.Coll(p).First(filter, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func DBPNftGetCollectionNFTs(address string, filterObj *common.Filter) ([]common.PNftAssetData, error) {
	var result []common.PNftAssetData

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
		filter = bson.M{"contract_address": bson.M{operator.Eq: address}, "token_id": bson.M{operator.Regex: query, "$options": "i"}}
	} else {
		filter = bson.M{"contract_address": bson.M{operator.Eq: address}}
	}
	err = mgm.Coll(&common.PNftAssetData{}).SimpleFind(&result, filter, &options.FindOptions{
		Limit: &limit,
		Skip:  &offset,
		Sort:  bson.D{{sort, order}},
	})
	if err != nil {
		return result, err
	}
	return result, nil
}

func DBPNftGetCollectionList(filterObj *common.Filter) ([]common.PNftCollectionData, error) {
	var result []common.PNftCollectionData

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
		filter = bson.M{"name": bson.M{operator.Regex: query, "$options": "i"}}
	} else {
		filter = bson.M{}
	}
	err = mgm.Coll(&common.PNftCollectionData{}).SimpleFind(&result, filter, &options.FindOptions{
		Limit: &limit,
		Skip:  &offset,
		Sort:  bson.D{{sort, order}},
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func DBPNftInsertSellOrder(data *common.PNftSellOrder) error {
	return mgm.Coll(data).Create(data)
}

func DBPNftGetNFTSellOrder(address string, nftids []string) ([]common.PNftSellOrder, error) {
	var result []common.PNftSellOrder

	fmt.Println("address: ", address)
	fmt.Println("nftids: ", nftids)

	filter := bson.M{"contract_address": bson.M{operator.Eq: address}, "token_id": bson.M{operator.All: nftids}}
	err := mgm.Coll(&common.PNftSellOrder{}).SimpleFind(&result, filter, &options.FindOptions{})
	if err != nil {
		return result, err
	}
	return result, nil
}

package common

import (
	"github.com/incognitochain/incognito-web-based-backend/papps/pnft"
	"github.com/kamva/mgm/v3"
)

type ListNftCache struct {
	mgm.DefaultModel `bson:",inline"`
	Address          string `bson:"address"`
	Data             string
}

type PNftCollectionData struct {
	mgm.DefaultModel `bson:",inline"`
	ContractAddress  string `bson:"contract_address"`
	Name             string `bson:"name"`
	CollectionSlug   string `bson:"collection_slug"`
	ImageUrl         string `bson:"image_url"`
	TotalSupply      int    `bson:"total_supply"`
	NumberOwners     int    `bson:"number_owners"`

	FloorPrice string `bson:"floor_price"`

	FloorPriceOneDay string `bson:"floor_price_one_day"`

	FloorPriceOneWeek string `bson:"floor_price_one_week"`

	VolumeFifteenMinutes string `bson:"volume_fifteen_minutes"`

	VolumeOneDay string `bson:"volume_one_day"`

	VolumeOneWeek string `bson:"volume_one_week"`

	BestCollectionBid string `bson:"best_collection_bid"`

	TotalCollectionBidValue string `bson:"total_collection_bid_value"`

	TraitFrequencies interface{} `bson:"trait_frequencies"`
}

type PNftAssetData struct {
	mgm.DefaultModel `bson:",inline"`
	UID              string         `bson:"uid"`
	ContractAddress  string         `bson:"contract_address"`
	TokenID          string         `bson:"token_id"`
	Name             string         `bson:"name"`
	Price            string         `bson:"price"`
	Detail           pnft.NFTDetail `bson:"detail"`
}

type Filter struct {
	Sort   string `json:"sort"`
	Order  string `json:"order"`
	Query  string `json:"query"`
	Page   int    `json:"page"`
	Offset int
	Limit  int `json:"limit"`
}

type PNftSellOrder struct {
	mgm.DefaultModel `bson:",inline"`
	IsCanceling      bool   `bson:"is_canceling"`
	OrderHash        string `bson:"order_hash"`
	Seller           string `bson:"seller"`
	ContractAddress  string `bson:"contract_address"`
	TokenID          string `bson:"token_id"`
	PaymentToken     string `bson:"payment_token"`
	Amount           string `bson:"amount"`
	Salt             string `bson:"salt"`
	OrderInput       string `bson:"order_input"` //this is bridge-eth/pnft.Input
	Network          int    `bson:"network"`
}

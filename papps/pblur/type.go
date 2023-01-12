package pblur

import "time"

// collection response from blur api:
type CollectionDetail struct {
	ContractAddress string `json:"contractAddress"`
	Name            string `json:"name"`
	CollectionSlug  string `json:"collectionSlug"`
	ImageUrl        string `json:"imageUrl"`
	TotalSupply     int    `json:"totalSupply"`
	NumberOwners    int    `json:"numberOwners"`

	FloorPrice struct {
		Amount string `json:"amount"`
		Unit   string `json:"unit"`
	} `json:"floorPrice"`

	FloorPriceOneDay struct {
		Amount string `json:"amount"`
		Unit   string `json:"unit"`
	} `json:"floorPriceOneDay"`

	FloorPriceOneWeek struct {
		Amount string `json:"amount"`
		Unit   string `json:"unit"`
	} `json:"floorPriceOneWeek"`

	VolumeFifteenMinutes struct {
		Amount string `json:"amount"`
		Unit   string `json:"unit"`
	} `json:"volumeFifteenMinutes"`

	VolumeOneDay struct {
		Amount string `json:"amount"`
		Unit   string `json:"unit"`
	} `json:"volumeOneDay"`

	VolumeOneWeek struct {
		Amount string `json:"amount"`
		Unit   string `json:"unit"`
	} `json:"volumeOneWeek"`

	BestCollectionBid struct {
		Amount string `json:"amount"`
		Unit   string `json:"unit"`
	} `json:"bestCollectionBid"`

	TotalCollectionBidValue struct {
		Amount string `json:"amount"`
		Unit   string `json:"unit"`
	} `json:"totalCollectionBidValue"`

	TraitFrequencies interface{} `json:"traitFrequencies"`
}

type NFTDetail struct {
	TokenID     string      `json:"tokenId"`
	Name        string      `json:"name"`
	ImageURL    string      `json:"imageUrl"`
	Traits      interface{} `json:"traits"`
	RarityScore float64     `json:"rarityScore"`
	RarityRank  int         `json:"rarityRank"`
	Price       struct {
		Amount      string    `json:"amount"`
		Unit        string    `json:"unit"`
		ListedAt    time.Time `json:"listedAt"`
		Marketplace string    `json:"marketplace"`
	} `json:"price"`
	HighestBid interface{} `json:"highestBid"`
	LastSale   struct {
		Amount   string    `json:"amount"`
		Unit     string    `json:"unit"`
		ListedAt time.Time `json:"listedAt"`
	} `json:"lastSale"`
	LastCostBasis struct {
		Amount   string    `json:"amount"`
		Unit     string    `json:"unit"`
		ListedAt time.Time `json:"listedAt"`
	} `json:"lastCostBasis"`
	Owner struct {
		Address  string      `json:"address"`
		Username interface{} `json:"username"`
	} `json:"owner"`
	NumberOwnedByOwner int  `json:"numberOwnedByOwner"`
	IsSuspicious       bool `json:"isSuspicious"`
}

// todo: update
// type NFTOrder struct {
// 	OrderHash string `json:"order_hash"`
// 	Chain     string `json:"chain"`
// 	Type      string `json:"type"`
// 	Price     struct {
// 		Current struct {
// 			Value    string `json:"value"`
// 			Currency string `json:"currency"`
// 		} `json:"current"`
// 	} `json:"price"`
// 	ProtocolData struct {
// 		Parameters struct {
// 			Offerer string `json:"offerer"`
// 			Offer   []struct {
// 				ItemType             int    `json:"itemType"`
// 				Token                string `json:"token"`
// 				IdentifierOrCriteria string `json:"identifierOrCriteria"`
// 				StartAmount          string `json:"startAmount"`
// 				EndAmount            string `json:"endAmount"`
// 			} `json:"offer"`
// 			Consideration []struct {
// 				ItemType             int    `json:"itemType"`
// 				Token                string `json:"token"`
// 				IdentifierOrCriteria string `json:"identifierOrCriteria"`
// 				StartAmount          string `json:"startAmount"`
// 				EndAmount            string `json:"endAmount"`
// 				Recipient            string `json:"recipient"`
// 			} `json:"consideration"`
// 			StartTime                       string `json:"startTime"`
// 			EndTime                         string `json:"endTime"`
// 			OrderType                       int    `json:"orderType"`
// 			Zone                            string `json:"zone"`
// 			ZoneHash                        string `json:"zoneHash"`
// 			Salt                            string `json:"salt"`
// 			ConduitKey                      string `json:"conduitKey"`
// 			TotalOriginalConsiderationItems int    `json:"totalOriginalConsiderationItems"`
// 			Counter                         int    `json:"counter"`
// 		} `json:"parameters"`
// 		Signature string `json:"signature"`
// 	} `json:"protocol_data"`
// }

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

type BuyPayload struct {
	TokenPrices []TokenPrice `json:"tokenPrices"`
	UserAddress string       `json:"userAddress"`
}

type TokenPrice struct {
	TokenID string `json:"tokenId"`
	Price   Price  `json:"price"`
}
type Price struct {
	Amount string `json:"amount"`
	Unit   string `json:"unit"`
}

type BuyDataResponse struct {
	Buys []struct {
		TxnData struct {
			To    string `json:"to"`
			Data  string `json:"data"`
			Value struct {
				Type string `json:"type"`
				Hex  string `json:"hex"`
			} `json:"value"`
		} `json:"txnData"`
		GasEstimate    int    `json:"gasEstimate"`
		AmountFromPool string `json:"amountFromPool"`
		IncludedTokens []struct {
			TokenID         string `json:"tokenId"`
			ContractAddress string `json:"contractAddress"`
		} `json:"includedTokens"`
	} `json:"buys"`
	CancelReasons []struct {
		TokenID string `json:"tokenId"`
		Reason  string `json:"reason"`
	} `json:"cancelReasons"`
}
type LoginData struct {
	Message       string    `json:"message"`
	WalletAddress string    `json:"walletAddress"`
	ExpiresOn     time.Time `json:"expiresOn"`
	Hmac          string    `json:"hmac"`
	Signature     string    `json:"signature"`
}

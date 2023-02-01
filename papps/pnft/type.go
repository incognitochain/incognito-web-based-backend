package pnft

import "time"

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
}

// just for crawl data from quicknode.
type Asset struct {
	Name              string `json:"name"`
	CollectionTokenID string `json:"collectionTokenId"`
	CollectionName    string `json:"collectionName"`
	ImageURL          string `json:"imageUrl"`
	CollectionAddress string `json:"collectionAddress"`
	Traits            []struct {
		Value     string `json:"value"`
		TraitType string `json:"trait_type"`
	} `json:"traits"`
	Chain       string `json:"chain"`
	Network     string `json:"network"`
	Description string `json:"description"`
	Provenance  []struct {
		BlockNumber string    `json:"blockNumber"`
		Date        time.Time `json:"date"`
		From        string    `json:"from"`
		To          string    `json:"to"`
		TxHash      string    `json:"txHash"`
	} `json:"provenance"`
	CurrentOwner string `json:"currentOwner"`
}

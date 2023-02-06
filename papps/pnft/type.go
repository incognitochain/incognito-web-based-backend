package pnft

import "time"

type NFTDetail struct {
	TokenID string `json:"tokenId"`
	Name    string `json:"name"`

	ImageURL             string      `json:"imageUrl"`
	BackgroundColor      interface{} `json:"background_color"`
	ImagePreviewURL      string      `json:"image_preview_url"`
	ImageThumbnailURL    string      `json:"image_thumbnail_url"`
	ImageOriginalURL     string      `json:"image_original_url"`
	AnimationURL         interface{} `json:"animation_url"`
	AnimationOriginalURL interface{} `json:"animation_original_url"`

	Traits      interface{} `json:"traits"`
	RarityScore float64     `json:"rarityScore"`
	RarityRank  int         `json:"rarityRank"`
	Price       Price       `json:"price"`
	HighestBid  interface{} `json:"highestBid"`
	LastSale    struct {
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

type Price struct {
	Amount      string    `json:"amount"`
	Unit        string    `json:"unit"`
	ListedAt    time.Time `json:"listedAt"`
	Marketplace string    `json:"marketplace"`
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

type OpenCollectionResp struct {
	BannerImageURL    string `json:"banner_image_url"`
	Description       string `json:"description"`
	ExternalURL       string `json:"external_url"`
	FeaturedImageURL  string `json:"featured_image_url"`
	ImageURL          string `json:"image_url"`
	LargeImageURL     string `json:"large_image_url"`
	Name              string `json:"name"`
	Slug              string `json:"slug"`
	TelegramURL       string `json:"telegram_url"`
	TwitterUsername   string `json:"twitter_username"`
	InstagramUsername string `json:"instagram_username"`
	WikiURL           string `json:"wiki_url"`
}

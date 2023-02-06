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

	Traits interface{} `json:"traits"`

	RarityScore float64 `json:"rarityScore"`
	RarityRank  int     `json:"rarityRank"`
	Price       Price   `json:"price"`

	ListingInfo string `bson:"listing_info"`

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

type Price struct {
	Amount      string    `json:"amount"`
	Unit        string    `json:"unit"`
	ListedAt    time.Time `json:"listedAt"`
	Marketplace string    `json:"marketplace"`
}

// just for crawl data from quicknode.
type QuicknodeNftDataResp struct {
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

type OpenSeaCollectionResp struct {
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

type MoralisNftDataResp struct {
	TokenAddress       string `json:"token_address"`
	TokenID            string `json:"token_id"`
	OwnerOf            string `json:"owner_of"`
	BlockNumber        string `json:"block_number"`
	BlockNumberMinted  string `json:"block_number_minted"`
	TokenHash          string `json:"token_hash"`
	Amount             string `json:"amount"`
	ContractType       string `json:"contract_type"`
	Name               string `json:"name"`
	Symbol             string `json:"symbol"`
	TokenURI           string `json:"token_uri"`
	Metadata           string `json:"metadata"`
	NormalizedMetadata *struct {
		Name         string      `json:"name"`
		Description  string      `json:"description"`
		AnimationURL interface{} `json:"animation_url"`
		ExternalLink interface{} `json:"external_link"`
		Image        string      `json:"image"`
		Attributes   []struct {
			TraitType   string      `json:"trait_type"`
			Value       string      `json:"value"`
			DisplayType interface{} `json:"display_type"`
			MaxValue    interface{} `json:"max_value"`
			TraitCount  int         `json:"trait_count"`
			Order       interface{} `json:"order"`
		} `json:"attributes"`
	} `json:"normalized_metadata"`
	LastTokenURISync time.Time `json:"last_token_uri_sync"`
	LastMetadataSync time.Time `json:"last_metadata_sync"`
	MinterAddress    string    `json:"minter_address"`
}

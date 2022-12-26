package popensea

import "time"

type CollectionDetail struct {
	Editors       []string `json:"editors"`
	PaymentTokens []struct {
		ID       int         `json:"id"`
		Symbol   string      `json:"symbol"`
		Address  string      `json:"address"`
		ImageURL string      `json:"image_url"`
		Name     interface{} `json:"name"`
		Decimals int         `json:"decimals"`
		EthPrice float64     `json:"eth_price"`
		UsdPrice float64     `json:"usd_price"`
	} `json:"payment_tokens"`
	PrimaryAssetContracts []struct {
		Address                     string      `json:"address"`
		AssetContractType           string      `json:"asset_contract_type"`
		CreatedDate                 string      `json:"created_date"`
		Name                        string      `json:"name"`
		NftVersion                  interface{} `json:"nft_version"`
		OpenseaVersion              interface{} `json:"opensea_version"`
		Owner                       int         `json:"owner"`
		SchemaName                  string      `json:"schema_name"`
		Symbol                      string      `json:"symbol"`
		TotalSupply                 string      `json:"total_supply"`
		Description                 interface{} `json:"description"`
		ExternalLink                interface{} `json:"external_link"`
		ImageURL                    interface{} `json:"image_url"`
		DefaultToFiat               bool        `json:"default_to_fiat"`
		DevBuyerFeeBasisPoints      int         `json:"dev_buyer_fee_basis_points"`
		DevSellerFeeBasisPoints     int         `json:"dev_seller_fee_basis_points"`
		OnlyProxiedTransfers        bool        `json:"only_proxied_transfers"`
		OpenseaBuyerFeeBasisPoints  int         `json:"opensea_buyer_fee_basis_points"`
		OpenseaSellerFeeBasisPoints int         `json:"opensea_seller_fee_basis_points"`
		BuyerFeeBasisPoints         int         `json:"buyer_fee_basis_points"`
		SellerFeeBasisPoints        int         `json:"seller_fee_basis_points"`
		PayoutAddress               interface{} `json:"payout_address"`
	} `json:"primary_asset_contracts"`
	Traits struct {
	} `json:"traits"`
	Stats struct {
		OneHourVolume         float64 `json:"one_hour_volume"`
		OneHourChange         float64 `json:"one_hour_change"`
		OneHourSales          float64 `json:"one_hour_sales"`
		OneHourSalesChange    float64 `json:"one_hour_sales_change"`
		OneHourAveragePrice   float64 `json:"one_hour_average_price"`
		OneHourDifference     float64 `json:"one_hour_difference"`
		SixHourVolume         float64 `json:"six_hour_volume"`
		SixHourChange         float64 `json:"six_hour_change"`
		SixHourSales          float64 `json:"six_hour_sales"`
		SixHourSalesChange    float64 `json:"six_hour_sales_change"`
		SixHourAveragePrice   float64 `json:"six_hour_average_price"`
		SixHourDifference     float64 `json:"six_hour_difference"`
		OneDayVolume          float64 `json:"one_day_volume"`
		OneDayChange          float64 `json:"one_day_change"`
		OneDaySales           float64 `json:"one_day_sales"`
		OneDaySalesChange     float64 `json:"one_day_sales_change"`
		OneDayAveragePrice    float64 `json:"one_day_average_price"`
		OneDayDifference      float64 `json:"one_day_difference"`
		SevenDayVolume        float64 `json:"seven_day_volume"`
		SevenDayChange        float64 `json:"seven_day_change"`
		SevenDaySales         float64 `json:"seven_day_sales"`
		SevenDayAveragePrice  float64 `json:"seven_day_average_price"`
		SevenDayDifference    float64 `json:"seven_day_difference"`
		ThirtyDayVolume       float64 `json:"thirty_day_volume"`
		ThirtyDayChange       float64 `json:"thirty_day_change"`
		ThirtyDaySales        float64 `json:"thirty_day_sales"`
		ThirtyDayAveragePrice float64 `json:"thirty_day_average_price"`
		ThirtyDayDifference   float64 `json:"thirty_day_difference"`
		TotalVolume           float64 `json:"total_volume"`
		TotalSales            float64 `json:"total_sales"`
		TotalSupply           float64 `json:"total_supply"`
		Count                 float64 `json:"count"`
		NumOwners             float64 `json:"num_owners"`
		AveragePrice          float64 `json:"average_price"`
		NumReports            float64 `json:"num_reports"`
		MarketCap             float64 `json:"market_cap"`
		FloorPrice            float64 `json:"floor_price"`
	} `json:"stats"`
	BannerImageURL          interface{} `json:"banner_image_url"`
	ChatURL                 interface{} `json:"chat_url"`
	CreatedDate             time.Time   `json:"created_date"`
	DefaultToFiat           bool        `json:"default_to_fiat"`
	Description             interface{} `json:"description"`
	DevBuyerFeeBasisPoints  string      `json:"dev_buyer_fee_basis_points"`
	DevSellerFeeBasisPoints string      `json:"dev_seller_fee_basis_points"`
	DiscordURL              interface{} `json:"discord_url"`
	DisplayData             struct {
		CardDisplayStyle string        `json:"card_display_style"`
		Images           []interface{} `json:"images"`
	} `json:"display_data"`
	ExternalURL                 interface{} `json:"external_url"`
	Featured                    bool        `json:"featured"`
	FeaturedImageURL            interface{} `json:"featured_image_url"`
	Hidden                      bool        `json:"hidden"`
	SafelistRequestStatus       string      `json:"safelist_request_status"`
	ImageURL                    interface{} `json:"image_url"`
	IsSubjectToWhitelist        bool        `json:"is_subject_to_whitelist"`
	LargeImageURL               interface{} `json:"large_image_url"`
	MediumUsername              interface{} `json:"medium_username"`
	Name                        string      `json:"name"`
	OnlyProxiedTransfers        bool        `json:"only_proxied_transfers"`
	OpenseaBuyerFeeBasisPoints  string      `json:"opensea_buyer_fee_basis_points"`
	OpenseaSellerFeeBasisPoints string      `json:"opensea_seller_fee_basis_points"`
	PayoutAddress               interface{} `json:"payout_address"`
	RequireEmail                bool        `json:"require_email"`
	ShortDescription            interface{} `json:"short_description"`
	Slug                        string      `json:"slug"`
	TelegramURL                 interface{} `json:"telegram_url"`
	TwitterUsername             interface{} `json:"twitter_username"`
	InstagramUsername           interface{} `json:"instagram_username"`
	WikiURL                     interface{} `json:"wiki_url"`
	IsNsfw                      bool        `json:"is_nsfw"`
	Fees                        struct {
		SellerFees  map[string]uint64 `json:"seller_fees"`
		OpenseaFees map[string]uint64 `json:"opensea_fees"`
	} `json:"fees"`
	IsRarityEnabled bool `json:"is_rarity_enabled"`
}

type NFTDetail struct {
	ID                   int         `json:"id"`
	NumSales             int         `json:"num_sales"`
	BackgroundColor      interface{} `json:"background_color"`
	ImageURL             string      `json:"image_url"`
	ImagePreviewURL      string      `json:"image_preview_url"`
	ImageThumbnailURL    string      `json:"image_thumbnail_url"`
	ImageOriginalURL     string      `json:"image_original_url"`
	AnimationURL         interface{} `json:"animation_url"`
	AnimationOriginalURL interface{} `json:"animation_original_url"`
	Name                 string      `json:"name"`
	Description          interface{} `json:"description"`
	ExternalLink         interface{} `json:"external_link"`
	AssetContract        struct {
		Address                     string      `json:"address"`
		AssetContractType           string      `json:"asset_contract_type"`
		CreatedDate                 string      `json:"created_date"`
		Name                        string      `json:"name"`
		NftVersion                  string      `json:"nft_version"`
		OpenseaVersion              interface{} `json:"opensea_version"`
		Owner                       int         `json:"owner"`
		SchemaName                  string      `json:"schema_name"`
		Symbol                      string      `json:"symbol"`
		TotalSupply                 string      `json:"total_supply"`
		Description                 string      `json:"description"`
		ExternalLink                string      `json:"external_link"`
		ImageURL                    string      `json:"image_url"`
		DefaultToFiat               bool        `json:"default_to_fiat"`
		DevBuyerFeeBasisPoints      int         `json:"dev_buyer_fee_basis_points"`
		DevSellerFeeBasisPoints     int         `json:"dev_seller_fee_basis_points"`
		OnlyProxiedTransfers        bool        `json:"only_proxied_transfers"`
		OpenseaBuyerFeeBasisPoints  int         `json:"opensea_buyer_fee_basis_points"`
		OpenseaSellerFeeBasisPoints int         `json:"opensea_seller_fee_basis_points"`
		BuyerFeeBasisPoints         int         `json:"buyer_fee_basis_points"`
		SellerFeeBasisPoints        int         `json:"seller_fee_basis_points"`
		PayoutAddress               interface{} `json:"payout_address"`
	} `json:"asset_contract"`
	Permalink  string `json:"permalink"`
	Collection struct {
		BannerImageURL          string      `json:"banner_image_url"`
		ChatURL                 interface{} `json:"chat_url"`
		CreatedDate             time.Time   `json:"created_date"`
		DefaultToFiat           bool        `json:"default_to_fiat"`
		Description             string      `json:"description"`
		DevBuyerFeeBasisPoints  string      `json:"dev_buyer_fee_basis_points"`
		DevSellerFeeBasisPoints string      `json:"dev_seller_fee_basis_points"`
		DiscordURL              interface{} `json:"discord_url"`
		DisplayData             struct {
			CardDisplayStyle string      `json:"card_display_style"`
			Images           interface{} `json:"images"`
		} `json:"display_data"`
		ExternalURL                 string      `json:"external_url"`
		Featured                    bool        `json:"featured"`
		FeaturedImageURL            interface{} `json:"featured_image_url"`
		Hidden                      bool        `json:"hidden"`
		SafelistRequestStatus       string      `json:"safelist_request_status"`
		ImageURL                    string      `json:"image_url"`
		IsSubjectToWhitelist        bool        `json:"is_subject_to_whitelist"`
		LargeImageURL               interface{} `json:"large_image_url"`
		MediumUsername              interface{} `json:"medium_username"`
		Name                        string      `json:"name"`
		OnlyProxiedTransfers        bool        `json:"only_proxied_transfers"`
		OpenseaBuyerFeeBasisPoints  string      `json:"opensea_buyer_fee_basis_points"`
		OpenseaSellerFeeBasisPoints string      `json:"opensea_seller_fee_basis_points"`
		PayoutAddress               interface{} `json:"payout_address"`
		RequireEmail                bool        `json:"require_email"`
		ShortDescription            interface{} `json:"short_description"`
		Slug                        string      `json:"slug"`
		TelegramURL                 interface{} `json:"telegram_url"`
		TwitterUsername             interface{} `json:"twitter_username"`
		InstagramUsername           interface{} `json:"instagram_username"`
		WikiURL                     interface{} `json:"wiki_url"`
		IsNsfw                      bool        `json:"is_nsfw"`
		Fees                        struct {
			SellerFees  map[string]uint64 `json:"seller_fees"`
			OpenseaFees map[string]uint64 `json:"opensea_fees"`
		} `json:"fees"`
		IsRarityEnabled bool `json:"is_rarity_enabled"`
	} `json:"collection"`
	Decimals          interface{} `json:"decimals"`
	TokenMetadata     string      `json:"token_metadata"`
	IsNsfw            bool        `json:"is_nsfw"`
	Owner             interface{} `json:"owner"`
	SeaportSellOrders []struct {
		CreatedDate    string `json:"created_date"`
		ClosingDate    string `json:"closing_date"`
		ListingTime    int    `json:"listing_time"`
		ExpirationTime int    `json:"expiration_time"`
		OrderHash      string `json:"order_hash"`
		ProtocolData   struct {
			Parameters struct {
				Offerer string `json:"offerer"`
				Offer   []struct {
					ItemType             int    `json:"itemType"`
					Token                string `json:"token"`
					IdentifierOrCriteria string `json:"identifierOrCriteria"`
					StartAmount          string `json:"startAmount"`
					EndAmount            string `json:"endAmount"`
				} `json:"offer"`
				Consideration []struct {
					ItemType             int    `json:"itemType"`
					Token                string `json:"token"`
					IdentifierOrCriteria string `json:"identifierOrCriteria"`
					StartAmount          string `json:"startAmount"`
					EndAmount            string `json:"endAmount"`
					Recipient            string `json:"recipient"`
				} `json:"consideration"`
				StartTime                       string `json:"startTime"`
				EndTime                         string `json:"endTime"`
				OrderType                       int    `json:"orderType"`
				Zone                            string `json:"zone"`
				ZoneHash                        string `json:"zoneHash"`
				Salt                            string `json:"salt"`
				ConduitKey                      string `json:"conduitKey"`
				TotalOriginalConsiderationItems int    `json:"totalOriginalConsiderationItems"`
				Counter                         int    `json:"counter"`
			} `json:"parameters"`
			Signature string `json:"signature"`
		} `json:"protocol_data"`
		ProtocolAddress string `json:"protocol_address"`
		Maker           struct {
			User          int    `json:"user"`
			ProfileImgURL string `json:"profile_img_url"`
			Address       string `json:"address"`
			Config        string `json:"config"`
		} `json:"maker"`
		Taker        interface{} `json:"taker"`
		CurrentPrice string      `json:"current_price"`
		MakerFees    []struct {
			Account struct {
				User          interface{} `json:"user"`
				ProfileImgURL string      `json:"profile_img_url"`
				Address       string      `json:"address"`
				Config        string      `json:"config"`
			} `json:"account"`
			BasisPoints string `json:"basis_points"`
		} `json:"maker_fees"`
		TakerFees       []interface{} `json:"taker_fees"`
		Side            string        `json:"side"`
		OrderType       string        `json:"order_type"`
		Cancelled       bool          `json:"cancelled"`
		Finalized       bool          `json:"finalized"`
		MarkedInvalid   bool          `json:"marked_invalid"`
		ClientSignature string        `json:"client_signature"`
		RelayID         string        `json:"relay_id"`
		CriteriaProof   interface{}   `json:"criteria_proof"`
	} `json:"seaport_sell_orders"`
	Creator struct {
		User struct {
			Username interface{} `json:"username"`
		} `json:"user"`
		ProfileImgURL string `json:"profile_img_url"`
		Address       string `json:"address"`
		Config        string `json:"config"`
	} `json:"creator"`
	Traits []struct {
		TraitType   string      `json:"trait_type"`
		Value       string      `json:"value"`
		DisplayType interface{} `json:"display_type"`
		MaxValue    interface{} `json:"max_value"`
		TraitCount  int         `json:"trait_count"`
		Order       interface{} `json:"order"`
	} `json:"traits"`
	LastSale struct {
		Asset struct {
			Decimals interface{} `json:"decimals"`
			TokenID  string      `json:"token_id"`
		} `json:"asset"`
		AssetBundle    interface{} `json:"asset_bundle"`
		EventType      string      `json:"event_type"`
		EventTimestamp string      `json:"event_timestamp"`
		AuctionType    interface{} `json:"auction_type"`
		TotalPrice     string      `json:"total_price"`
		PaymentToken   struct {
			Symbol   string `json:"symbol"`
			Address  string `json:"address"`
			ImageURL string `json:"image_url"`
			Name     string `json:"name"`
			Decimals int    `json:"decimals"`
			EthPrice string `json:"eth_price"`
			UsdPrice string `json:"usd_price"`
		} `json:"payment_token"`
		Transaction struct {
			BlockHash   string `json:"block_hash"`
			BlockNumber string `json:"block_number"`
			FromAccount struct {
				User struct {
					Username interface{} `json:"username"`
				} `json:"user"`
				ProfileImgURL string `json:"profile_img_url"`
				Address       string `json:"address"`
				Config        string `json:"config"`
			} `json:"from_account"`
			ID        int    `json:"id"`
			Timestamp string `json:"timestamp"`
			ToAccount struct {
				User          interface{} `json:"user"`
				ProfileImgURL string      `json:"profile_img_url"`
				Address       string      `json:"address"`
				Config        string      `json:"config"`
			} `json:"to_account"`
			TransactionHash  string `json:"transaction_hash"`
			TransactionIndex string `json:"transaction_index"`
		} `json:"transaction"`
		CreatedDate string `json:"created_date"`
		Quantity    string `json:"quantity"`
	} `json:"last_sale"`
	TopBid                  interface{} `json:"top_bid"`
	ListingDate             interface{} `json:"listing_date"`
	SupportsWyvern          bool        `json:"supports_wyvern"`
	RarityData              interface{} `json:"rarity_data"`
	TransferFee             interface{} `json:"transfer_fee"`
	TransferFeePaymentToken interface{} `json:"transfer_fee_payment_token"`
	TokenID                 string      `json:"token_id"`
}

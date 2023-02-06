package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/gosimple/slug"
	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/papps/pnft"
)

// TODO: Lam
// this func is called when user listing nfts:
func createNftAndCollectionToInsertDBWhenListing(orderList []common.PNftSellOrder) error {

	if len(orderList) == 0 {
		return errors.New("orderList empty")
	}

	sellerAddress := orderList[0].Seller
	// get nfts of seller:
	nftListOwner, err := database.DBPNftGetListNftCacheTableByAddress(sellerAddress)

	if err != nil {
		return err
	}
	if nftListOwner == nil {
		return errors.New("nftListOwner nil")
	}

	// convert nftList data string to struct:
	var moralisNftDataRespList []pnft.MoralisNftDataResp
	err = json.Unmarshal([]byte(nftListOwner.Data), &moralisNftDataRespList)
	if err != nil {
		return fmt.Errorf("err from Unmarshal nftListOwner.Data: %v", err.Error())
	}

	// check from PNftAssetData database:
	for _, order := range orderList {
		pNftData, _ := database.DBPNftGetNFTDetail(order.ContractAddress, order.TokenID)
		if err != nil {
			return err
		}
		if pNftData == nil {
			// insert now:
			for _, moralisNftData := range moralisNftDataRespList {
				if strings.EqualFold(moralisNftData.TokenAddress, order.ContractAddress) && strings.EqualFold(moralisNftData.TokenID, order.TokenID) {
					pNftData, _ = createMoralisNftToPNftNftForListing(order.ContractAddress, order.Amount, &moralisNftData)
					if pNftData != nil {
						// insert pnft to pnft marketplace now:
						err := database.DBPNftInsertPNftAssetDataTable(pNftData)
						if err != nil {
							fmt.Println("can not DBPNftInsertPNftAssetDataTable: ", err)
							return err
						}
						// create collection to insert db:
						// check exist first:
						collection, _ := database.DBBlurGetCollectionByAddressDetail(order.ContractAddress)
						if collection == nil {
							// insert now:
							// try to get collection from opensea first:
							osCollection, err := pnft.RetrieveGetCollectionInfoFromOpensea(config.OpenSeaAPI, "", order.ContractAddress)
							if err != nil {
								fmt.Println("can not RetrieveGetCollectionInfoFromOpensea: ", err, "skip....")
							}
							if osCollection != nil {
								collection, err = convertOpenSeaCollectionToPNftCollection(order.ContractAddress, osCollection)
								if err != nil {
									fmt.Println("can not convertOpenSeaCollectionToPNftCollection: ", err)
								}
								if collection != nil {
									err = database.DBPNftInsertPNftCollectionDataTable(collection)
									if err != nil {
										fmt.Println("can not DBPNftInsertPNftCollectionDataTable: ", err)
										return err
									}
								}
							}
						}
						if collection == nil {
							//create the collection from pNft data:
							collection, err = createPNftCollectionFromPNftAsset(order.ContractAddress, pNftData)
							if err != nil {
								fmt.Println("can not createPNftCollectionFromPNftAsset: ", err)
								return err
							}
							// insert the collection db:
							err = database.DBPNftInsertPNftCollectionDataTable(collection)
							if err != nil {
								fmt.Println("can not DBPNftInsertPNftCollectionDataTable: ", err)
								return err
							}
						}
					}
					break
				}
			}
		}

	}

	return nil
}

func convertOpenSeaCollectionToPNftCollection(contractAddress string, osCollection *pnft.OpenSeaCollectionResp) (*common.PNftCollectionData, error) {
	if osCollection == nil {
		return nil, errors.New("collection is nil")
	}
	return &common.PNftCollectionData{
		ContractAddress: contractAddress,
		Name:            osCollection.Name,
		CollectionSlug:  osCollection.Slug,

		ImageUrl:       osCollection.ImageURL,
		LargeImageURL:  osCollection.LargeImageURL,
		BannerImageURL: osCollection.BannerImageURL,

		ExternalURL:       osCollection.ExternalURL,
		TelegramURL:       osCollection.TelegramURL,
		TwitterUsername:   osCollection.TwitterUsername,
		InstagramUsername: osCollection.InstagramUsername,
		WikiURL:           osCollection.WikiURL,
	}, nil
}

func createPNftCollectionFromPNftAsset(contractAddress string, pnft *common.PNftAssetData) (*common.PNftCollectionData, error) {
	if pnft == nil {
		return nil, errors.New("pnft is nil")
	}
	return &common.PNftCollectionData{
		ContractAddress: contractAddress,
		Name:            pnft.Name,
		CollectionSlug:  slug.Make(pnft.Name),

		ImageUrl:       pnft.Detail.ImageURL,
		LargeImageURL:  pnft.Detail.ImageURL,
		BannerImageURL: "",

		ExternalURL:       "",
		TelegramURL:       "",
		TwitterUsername:   "",
		InstagramUsername: "",
		WikiURL:           "",
	}, nil
}

func convertOpenSeaNftToPNftAsset(contractAddress string, osNft *common.OpenseaAssetData) (*common.PNftAssetData, error) {
	if osNft == nil {
		return nil, errors.New("osNft is nil")
	}
	price := "0"
	if len(osNft.Detail.SeaportSellOrders) > 0 {
		price = osNft.Detail.SeaportSellOrders[0].CurrentPrice
	}

	priceInfo := pnft.Price{
		Amount: price,
		Unit:   "ETH",
		// ListedAt:    osNft.ListedAt,
		Marketplace: "opensea",
	}

	return &common.PNftAssetData{
		UID:             osNft.Address + osNft.TokenID,
		ContractAddress: osNft.Address,
		TokenID:         osNft.TokenID,
		Name:            osNft.Name,
		Price:           price,
		Detail: pnft.NFTDetail{
			TokenID: osNft.TokenID,
			Name:    osNft.Name,

			ImageURL: osNft.Detail.ImageURL,

			BackgroundColor:      osNft.Detail.BackgroundColor,
			ImagePreviewURL:      osNft.Detail.ImagePreviewURL,
			ImageThumbnailURL:    osNft.Detail.ImageThumbnailURL,
			ImageOriginalURL:     osNft.Detail.ImageOriginalURL,
			AnimationURL:         osNft.Detail.AnimationURL,
			AnimationOriginalURL: osNft.Detail.AnimationOriginalURL,

			Traits: osNft.Detail.Traits,
			// RarityScore: osNft.Detail.RarityScore,
			// RarityRank:  osNft.Detail.RarityRank,
			Price: priceInfo,
			// HighestBid: osNft.TokenID,
			// LastSale: map[string]interface{}{
			// 	Amount:   osNft.TokenID,
			// 	Unit:     osNft.TokenID,
			// 	ListedAt: osNft.TokenID,
			// },
			// LastCostBasis: map[string]interface{}{
			// 	Amount:   osNft.TokenID,
			// 	Unit:     osNft.TokenID,
			// 	ListedAt: osNft.TokenID,
			// },
		},
	}, nil
}

func createMoralisNftToPNftNftForListing(contractAddress, price string, moralistNft *pnft.MoralisNftDataResp) (*common.PNftAssetData, error) {
	if moralistNft == nil {
		return nil, errors.New("moralistNft is nil")
	}
	priceInfo := pnft.Price{
		Amount: price,
		Unit:   "ETH",
		// ListedAt:    moralistNft.ListedAt,
		Marketplace: "pnft",
	}
	imageUrl := ""
	var traits interface{}
	if moralistNft.NormalizedMetadata != nil {
		imageUrl = moralistNft.NormalizedMetadata.Image
		traits = moralistNft.NormalizedMetadata.Attributes
	}

	return &common.PNftAssetData{
		UID:             contractAddress + moralistNft.TokenID,
		ContractAddress: contractAddress,
		TokenID:         moralistNft.TokenID,
		Name:            moralistNft.Name,
		Price:           price,
		Detail: pnft.NFTDetail{
			TokenID:              moralistNft.TokenID,
			Name:                 moralistNft.Name,
			ImageURL:             imageUrl,
			ImagePreviewURL:      imageUrl,
			ImageThumbnailURL:    imageUrl,
			ImageOriginalURL:     imageUrl,
			AnimationURL:         imageUrl,
			AnimationOriginalURL: imageUrl,

			Traits: traits,
			// RarityScore: osNft.Detail.RarityScore,
			// RarityRank:  osNft.Detail.RarityRank,
			Price: priceInfo,
			// HighestBid: osNft.TokenID,
			// LastSale: map[string]interface{}{
			// 	Amount:   osNft.TokenID,
			// 	Unit:     osNft.TokenID,
			// 	ListedAt: osNft.TokenID,
			// },
			// LastCostBasis: map[string]interface{}{
			// 	Amount:   osNft.TokenID,
			// 	Unit:     osNft.TokenID,
			// 	ListedAt: osNft.TokenID,
			// },
		},
	}, nil
}

package submitproof

import (
	"fmt"
	"log"
	"strings"
	"time"

	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/papps/pblur"

	"github.com/incognitochain/incognito-web-based-backend/slacknoti"
)

func processPendingBlurTx(tx wcommon.PappTxData) error {
	txDetail, err := incClient.GetTxDetail(tx.IncTx)
	if err != nil {
		if strings.Contains(err.Error(), "RPC returns an error:") {
			err = database.DBUpdatePappTxStatus(tx.IncTx, wcommon.StatusSubmitFailed, err.Error())
			if err != nil {
				log.Println("DBUpdateShieldTxStatus err:", err)
				return err
			}
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[Blur]` submit Blur failed 😵 `%v` \n", tx.IncTx))
			return nil
		}
		return err
	}
	if txDetail.IsInBlock {
		status, err := checkBeaconBridgeAggUnshieldStatus(tx.IncTx)
		if err != nil {
			return err
		}

		switch status {
		case 0:
			err = database.DBUpdatePappTxStatus(tx.IncTx, wcommon.StatusRejected, "")
			if err != nil {
				return err
			}
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[Blur]` inctx `%v` rejected by beacon 😢\n", tx.IncTx))
		case 1:
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[Blur]` inctx `%v` accepted by beacon 👌\n", tx.IncTx))
			err = database.DBUpdatePappTxStatus(tx.IncTx, wcommon.StatusAccepted, "")
			if err != nil {
				return err
			}
			err = database.DBUpdatePappTxSubmitOutchainStatus(tx.IncTx, wcommon.StatusWaiting)
			if err != nil {
				return err
			}
			for _, network := range tx.Networks {
				_, err := SubmitOutChainTx(tx.IncTx, network, tx.IsUnifiedToken, false, wcommon.ExternalTxTypeBlur)
				if err != nil {
					return err
				}
			}
		default:
			if tx.Status != wcommon.StatusExecuting && tx.Status != wcommon.StatusAccepted {
				err = database.DBUpdatePappTxStatus(tx.IncTx, wcommon.StatusExecuting, "")
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func updateBlurCollection() {
	fmt.Println("start watcher ....")
	for {

		fmt.Println("started watcher ....")

		loadBlurAPIKey()

		filters := `{"sort":"VOLUME_ONE_DAY","order":"DESC"}`

		for {

			// get list collection:
			listCollection, err := pblur.RetrieveCollectionList(config.BlurAPI, config.BlurToken, filters)

			if err != nil {
				log.Println("RetrieveCollectionList err: ", err)
				break
			}

			if len(listCollection) == 0 {
				break
			}

			// update collection to db:
			err = database.DBBlurSaveCollection(listCollection)
			if err != nil {
				log.Println("DBSaveCollectionsInfo err: ", err)
			}

			// get detail:
			for _, collectionReq := range listCollection {

				// filter: {"cursor":{"price":{"unit":"ETH","time":"2023-01-08T21:16:52.000Z","amount":"0.54"},"tokenId":"3823"},"traits":[],"hasAsks":true}
				/*
					price": {
						"amount": "0.54",
						"unit": "ETH",
						"listedAt": "2023-01-08T21:16:52.000Z",
						"marketplace": "OPENSEA"
					},
				*/

				filterDetail := `{"traits":[],"hasAsks":true}`

				for {
					nftDetailList, err := pblur.RetrieveCollectionAssets(config.BlurAPI, config.BlurToken, collectionReq.CollectionSlug, filterDetail)
					if err != nil {
						log.Println("RetrieveAssetDetail err: ", err)
						break
					}
					if len(nftDetailList) == 0 {
						break
					}

					// update db:
					err = database.DBBlurSaveNFTDetail(collectionReq.ContractAddress, nftDetailList)
					if err != nil {
						log.Println("DBBlurSaveNFTDetail err: ", err)
					}

					lastItem := nftDetailList[len(nftDetailList)-1]
					filterDetail = fmt.Sprintf(`{"cursor":{"price":{"unit":"ETH","time":"%s","amount":"%s"},"tokenId":"%s"},"traits":[],"hasAsks":true}`, lastItem.Price.ListedAt, lastItem.Price.Amount, lastItem.TokenID)

					time.Sleep(1000 * time.Millisecond)
				}

			}

			// set filter for collection to continue get data:
			lastCollection := listCollection[len(listCollection)-1]
			filters = fmt.Sprintf(`{"sort":"VOLUME_ONE_DAY","order":"DESC","cursor":{"contractAddress":"%s","volumeOneDay":"%s"}}`, lastCollection.ContractAddress, lastCollection.VolumeOneDay.Amount)

			time.Sleep(1000 * time.Millisecond)
		}
		log.Println("done update Blur Collections detail")

		time.Sleep(10 * time.Minute)
	}

}

// func updateBlurCollectionAssets() {
// 	for {
// 		time.Sleep(10 * time.Second)
// 		defaultList, err := database.DBGetDefaultCollectionList()
// 		if err != nil {
// 			log.Println(err)
// 			continue
// 		}
// 		loadBlurAPIKey()
// 		if config.NetworkID == "mainnet" {
// 			t := time.Now()
// 			for _, collection := range defaultList {
// 				orderList := []pBlur.NFTOrder{}
// 				next := ""
// 				for {
// 					time.Sleep(500 * time.Millisecond)
// 					list, nextStr, err := pBlur.RetrieveCollectionListing(config.BlurAPIKey, collection.Slug, next)
// 					if err != nil {
// 						log.Println("RetrieveCollectionListing error: ", err)
// 						go slacknoti.SendSlackNoti(fmt.Sprintf("`[Blur]` can't retrieve %v collection listing", collection.Slug))
// 						break
// 					}
// 					log.Println("next", next, nextStr)
// 					if nextStr == next || nextStr == "" {
// 						if nextStr == "" && len(orderList) == 0 {
// 							for _, v := range list {
// 								if v.Price.Current.Currency == "eth" {
// 									orderList = append(orderList, v)
// 								}
// 							}
// 						}
// 						break
// 					}
// 					next = nextStr
// 					for _, v := range list {
// 						if v.Price.Current.Currency == "eth" {
// 							orderList = append(orderList, v)
// 						}
// 					}
// 					if len(orderList) >= 1000 {
// 						break
// 					}
// 				}
// 				log.Println("len(orderList)", collection.Slug, len(orderList))
// 				nftsToGetBatch := make([][]string, int(math.Ceil(float64(len(orderList))/30)))
// 				for idx, order := range orderList {
// 					nftid := order.ProtocolData.Parameters.Offer[0].IdentifierOrCriteria
// 					batch := int(math.Floor(float64(idx) / 30))
// 					nftsToGetBatch[batch] = append(nftsToGetBatch[batch], nftid)
// 				}
// 				for _, nftBatch := range nftsToGetBatch {
// 					time.Sleep(500 * time.Millisecond)
// 					assets, err := pBlur.RetrieveCollectionAssetByIDs(config.BlurAPIKey, collection.Address, nftBatch)
// 					if err != nil {
// 						log.Println("RetrieveCollectionAssetByIDs error: ", err)
// 						go slacknoti.SendSlackNoti(fmt.Sprintf("`[Blur]` can't retrieve %v collection assets", collection.Slug))
// 						continue
// 					}
// 					err = database.DBSaveNFTDetail(assets)
// 					if err != nil {
// 						log.Println("DBSaveNFTDetail error: ", err)
// 						continue
// 					}
// 				}
// 			}
// 			log.Println("done update Blur Collections Assets in", time.Since(t))
// 			go slacknoti.SendSlackNoti(fmt.Sprintf("`[Blur]`done update Blur Collections Assets in %v", time.Since(t)))
// 		}
// 	}
// }

// func updateBlurCollectionList() {
// for {
// 	offset := 0
// 	pBlur.RetrieveCollectionList(config.BlurAPI, "", 300)
// 	defaultList, err := database.DBGetDefaultCollectionList()
// 	if err != nil {
// 		log.Println(err)
// 		continue
// 	}
// 	collections := []pBlur.CollectionDetail{}
// 	for _, collection := range defaultList {
// 		collectionDetail, err := pBlur.RetrieveCollectionDetail(config.BlurAPI, config.BlurAPIKey, collection.Slug)
// 		if err != nil {
// 			log.Println(err)
// 			continue
// 		}
// 		collections = append(collections, *collectionDetail)
// 		time.Sleep(1500 * time.Millisecond)
// 	}

// 	err = database.DBSaveCollectionsInfo(collections)
// 	if err != nil {
// 		log.Println(err)
// 		continue
// 	}
// 	time.Sleep(20 * time.Second)
// }
// }

func loadBlurAPIKey() {
	if len(config.BlurToken) == 0 {
		blurAuthToken, err := database.DBGetPappAPIKey("BlurAuthToken")
		if err != nil {
			log.Println("DBGetPappAPIKey(Blur)", err)
			return
		}
		config.BlurToken = blurAuthToken
	}

}

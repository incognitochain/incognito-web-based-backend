package submitproof

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/papps/popensea"
	"github.com/incognitochain/incognito-web-based-backend/slacknoti"
)

func processPendingOpenseaTx(tx wcommon.PappTxData) error {
	txDetail, err := incClient.GetTxDetail(tx.IncTx)
	if err != nil {
		if strings.Contains(err.Error(), "RPC returns an error:") {
			err = database.DBUpdatePappTxStatus(tx.IncTx, wcommon.StatusSubmitFailed, err.Error())
			if err != nil {
				log.Println("DBUpdateShieldTxStatus err:", err)
				return err
			}
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[opensea]` submit opensea failed 😵 `%v` \n", tx.IncTx))
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
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[opensea]` inctx `%v` rejected by beacon 😢\n", tx.IncTx))
		case 1:
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[opensea]` inctx `%v` accepted by beacon 👌\n", tx.IncTx))
			err = database.DBUpdatePappTxStatus(tx.IncTx, wcommon.StatusAccepted, "")
			if err != nil {
				return err
			}
			err = database.DBUpdatePappTxSubmitOutchainStatus(tx.IncTx, wcommon.StatusWaiting)
			if err != nil {
				return err
			}
			for _, network := range tx.Networks {
				_, err := SubmitOutChainTx(tx.IncTx, network, tx.IsUnifiedToken, false, wcommon.ExternalTxTypeOpensea)
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

func updateOpenSeaCollectionDetail() {
	for {
		time.Sleep(8 * time.Second)
		defaultList, err := database.DBGetDefaultCollectionList()
		if err != nil {
			log.Println(err)
			continue
		}
		// collections := []popensea.CollectionDetail{}
		for _, collection := range defaultList {
			collectionDetail, err := popensea.RetrieveCollectionDetail(config.OpenSeaAPI, "", collection.Slug)
			if err != nil {
				log.Println(err)
				continue
			}
			err = database.DBSaveCollectionsInfo([]popensea.CollectionDetail{*collectionDetail})
			if err != nil {
				log.Println(err)
				continue
			}
			time.Sleep(1000 * time.Millisecond)
		}
		log.Println("done update OpenSea Collections detail")
	}
}

func updateOpenSeaCollectionAssets() {
	for {
		time.Sleep(10 * time.Second)
		defaultList, err := database.DBGetDefaultCollectionList()
		if err != nil {
			log.Println(err)
			continue
		}
		if config.NetworkID == "mainnet" {
			for _, collection := range defaultList {
				// collectionData, err := database.DBGetCollectionsInfo(collection.Address)
				// if err != nil {
				// 	log.Println(err)
				// 	continue
				// }
				orderList := []popensea.NFTOrder{}
				next := ""
				for {
					time.Sleep(500 * time.Millisecond)
					list, nextStr, err := popensea.RetrieveCollectionListing(config.OpenSeaAPIKey, collection.Slug, next)
					if err != nil {
						log.Println("RetrieveCollectionListing error: ", err)
						go slacknoti.SendSlackNoti(fmt.Sprintf("`[opensea]` can't retrieve %v collection listing", collection.Slug))
						break
					}
					if nextStr == next {
						break
					}
					next = nextStr
					orderList = append(orderList, list...)
				}
				log.Println("len(orderList)", len(orderList))
				nftsToGetBatch := make([][]string, int(math.Ceil(float64(len(orderList))/30)))
				for idx, order := range orderList {
					nftid := order.ProtocolData.Parameters.Offer[0].IdentifierOrCriteria
					batch := int(math.Floor(float64(idx) / 30))
					nftsToGetBatch[batch] = append(nftsToGetBatch[batch], nftid)
				}
				for _, nftBatch := range nftsToGetBatch {
					time.Sleep(500 * time.Millisecond)
					assets, err := popensea.RetrieveCollectionAssetByIDs(config.OpenSeaAPIKey, collection.Address, nftBatch)
					if err != nil {
						log.Println("RetrieveCollectionAssetByIDs error: ", err)
						continue
					}
					err = database.DBSaveNFTDetail(assets)
					if err != nil {
						log.Println("DBSaveNFTDetail error: ", err)
						continue
					}
				}
			}

		}
	}
}

// func updateOpenSeaCollectionList() {
// for {
// 	offset := 0
// 	popensea.RetrieveCollectionList(config.OpenSeaAPI, "", 300)
// 	defaultList, err := database.DBGetDefaultCollectionList()
// 	if err != nil {
// 		log.Println(err)
// 		continue
// 	}
// 	collections := []popensea.CollectionDetail{}
// 	for _, collection := range defaultList {
// 		collectionDetail, err := popensea.RetrieveCollectionDetail(config.OpenSeaAPI, config.OpenSeaAPIKey, collection.Slug)
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

package submitproof

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/incognitochain/go-incognito-sdk-v2/common/base58"
	"github.com/incognitochain/go-incognito-sdk-v2/metadata/bridge"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/papps/popensea"
	"github.com/incognitochain/incognito-web-based-backend/slacknoti"
)

func processPendingOpenseaTx(tx wcommon.PappTxData) error {

	txTypeStr := ""
	switch tx.Type {
	case wcommon.ExternalTxTypeOpenseaBuy:
		txTypeStr = "opensea"
	case wcommon.ExternalTxTypeOpenseaOffer:
		txTypeStr = "opensea-offer"
	case wcommon.ExternalTxTypeOpenseaOfferCancel:
		txTypeStr = "opensea-cancel"
	}
	txDetail, err := incClient.GetTxDetail(tx.IncTx)
	if err != nil {
		if strings.Contains(err.Error(), "RPC returns an error:") {
			err = database.DBUpdatePappTxStatus(tx.IncTx, wcommon.StatusSubmitFailed, err.Error())
			if err != nil {
				log.Println("DBUpdateShieldTxStatus err:", err)
				return err
			}
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[%v]` submit opensea failed 😵 `%v` \n", txTypeStr, tx.IncTx))
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
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[%v]` inctx `%v` rejected by beacon 😢\n", txTypeStr, tx.IncTx))
		case 1:
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[%v]` inctx `%v` accepted by beacon 👌\n", txTypeStr, tx.IncTx))
			err = database.DBUpdatePappTxStatus(tx.IncTx, wcommon.StatusAccepted, "")
			if err != nil {
				return err
			}
			err = database.DBUpdatePappTxSubmitOutchainStatus(tx.IncTx, wcommon.StatusWaiting)
			if err != nil {
				return err
			}
			for _, network := range tx.Networks {
				switch tx.Type {
				case wcommon.ExternalTxTypeOpenseaBuy:
					_, err := SubmitOutChainTx(tx.IncTx, network, tx.IsUnifiedToken, false, wcommon.ExternalTxTypeOpenseaBuy)
					if err != nil {
						return err
					}
				case wcommon.ExternalTxTypeOpenseaOffer:
					_, err := SubmitOutChainTx(tx.IncTx, network, tx.IsUnifiedToken, false, wcommon.ExternalTxTypeOpenseaOffer)
					if err != nil {
						return err
					}
				case wcommon.ExternalTxTypeOpenseaOfferCancel:
					_, err := SubmitOutChainTx(tx.IncTx, network, tx.IsUnifiedToken, false, wcommon.ExternalTxTypeOpenseaOfferCancel)
					if err != nil {
						return err
					}
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
		loadOpenseaAPIKey()
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
		loadOpenseaAPIKey()
		if config.NetworkID == "mainnet" {
			t := time.Now()
			for _, collection := range defaultList {
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
					log.Println("next", next, nextStr)
					if nextStr == next || nextStr == "" {
						if nextStr == "" && len(orderList) == 0 {
							for _, v := range list {
								if v.Price.Current.Currency == "eth" {
									orderList = append(orderList, v)
								}
							}
						}
						break
					}
					next = nextStr
					for _, v := range list {
						if v.Price.Current.Currency == "eth" {
							orderList = append(orderList, v)
						}
					}
					if len(orderList) >= 1000 {
						break
					}
				}
				log.Println("len(orderList)", collection.Slug, len(orderList))
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
						go slacknoti.SendSlackNoti(fmt.Sprintf("`[opensea]` can't retrieve %v collection assets", collection.Slug))
						continue
					}
					err = database.DBSaveNFTDetail(assets)
					if err != nil {
						log.Println("DBSaveNFTDetail error: ", err)
						continue
					}
				}
			}
			log.Println("done update OpenSea Collections Assets in", time.Since(t))
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[opensea]`done update OpenSea Collections Assets in %v", time.Since(t)))
		}
	}
}

func createExternalOpenseaClaimTx(incTxHash, sign, network string, txType int) (*wcommon.ExternalTxStatus, error) {
	var result wcommon.ExternalTxStatus
	//TODO: opensea
	// networkID := wcommon.GetNetworkID(network)
	networkInfo, err := database.DBGetBridgeNetworkInfo(network)
	if err != nil {
		return nil, err
	}

	networkChainId := networkInfo.ChainID

	networkChainIdInt := new(big.Int)
	networkChainIdInt.SetString(networkChainId, 10)

	incTxData, err := database.DBGetPappTxData(strings.Split(incTxHash, "_")[0])
	if err != nil {
		return nil, err
	}

	swapInfo := wcommon.PappSwapInfo{}

	err = json.Unmarshal([]byte(incTxData.PappSwapInfo), &swapInfo)
	if err != nil {
		return nil, err
	}

	offer := popensea.OrderComponents{}
	offerBytes, err := hex.DecodeString(swapInfo.AdditionalData)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(offerBytes, &offer)
	if err != nil {
		return nil, err
	}

	papps, err := database.DBRetrievePAppsByNetwork("eth")
	if err != nil {
		return nil, err
	}
	offerAdapterAddr, exist := papps.AppContracts["opensea-offer"]
	if !exist {
		return nil, err
	}
	openseaOfferAddr := common.HexToAddress(offerAdapterAddr)

	privKey, _ := crypto.HexToECDSA(config.EVMKey)
	i := 0
retry:
	if i == 10 {
		return nil, errors.New("submit tx outchain failed")
	}
	for _, endpoint := range networkInfo.Endpoints {
		evmClient, err := ethclient.Dial(endpoint)
		if err != nil {
			log.Println(err)
			continue
		}
		openseaOffer, err := popensea.NewOpenseaOffer(openseaOfferAddr, evmClient)
		if err != nil {
			log.Println("popensea.NewOpenseaOffer error", err)
			time.Sleep(1 * time.Second)
			continue
		}
		gasPrice, err := evmClient.SuggestGasPrice(context.Background())
		if err != nil {
			log.Println(err)
			continue
		}

		auth, err := bind.NewKeyedTransactorWithChainID(privKey, networkChainIdInt)
		if err != nil {
			log.Println(err)
			continue
		}

		gasPrice = gasPrice.Mul(gasPrice, big.NewInt(12))
		gasPrice = gasPrice.Div(gasPrice, big.NewInt(10))

		auth.GasPrice = gasPrice
		if network == "eth" {
			auth.GasLimit = wcommon.EVMGasLimitETH
		} else {
			if network == "bsc" {
				auth.GasLimit = wcommon.EVMGasLimitPancake
			}
			auth.GasLimit = wcommon.EVMGasLimit
		}

		result.Type = txType
		result.Network = network
		result.IncRequestTx = incTxHash

		// address, err := wcommon.GetEVMAddress(config.EVMKey)
		// if err != nil {
		// 	log.Println(err)
		// 	continue
		// }
		// account := common.HexToAddress(address)
		// pendingNonce, _ := evmClient.PendingNonceAt(context.Background(), account)
		// auth.Nonce = new(big.Int).SetUint64(pendingNonce)

		tx, err := openseaOffer.Claim(auth, offer)
		if err != nil {
			log.Println(err)
			if strings.Contains(err.Error(), "insufficient funds") {
				return nil, errors.New("submit tx outchain failed err insufficient funds")
			}
			continue
		}
		result.Txhash = tx.Hash().String()
		result.Status = wcommon.StatusPending
		result.Nonce = tx.Nonce()
		break
	}

	if result.Txhash == "" {
		i++
		time.Sleep(2 * time.Second)
		goto retry
	}

	return &result, nil
}

func createExternalOpenseaCancelTx(incTxHash, sign, network string, txType int) (*wcommon.ExternalTxStatus, error) {
	var result wcommon.ExternalTxStatus
	//TODO: opensea
	// networkID := wcommon.GetNetworkID(network)
	networkInfo, err := database.DBGetBridgeNetworkInfo(network)
	if err != nil {
		return nil, err
	}
	networkChainId := networkInfo.ChainID

	networkChainIdInt := new(big.Int)
	networkChainIdInt.SetString(networkChainId, 10)
	papps, err := database.DBRetrievePAppsByNetwork("eth")
	if err != nil {
		return nil, err
	}
	offerAdapterAddr, exist := papps.AppContracts["opensea-offer"]
	if !exist {
		return nil, err
	}
	openseaOfferAddr := common.HexToAddress(offerAdapterAddr)

	privKey, _ := crypto.HexToECDSA(config.EVMKey)
	i := 0
retry:
	if i == 10 {
		return nil, errors.New("submit tx outchain failed")
	}
	for _, endpoint := range networkInfo.Endpoints {
		evmClient, err := ethclient.Dial(endpoint)
		if err != nil {
			log.Println(err)
			continue
		}
		openseaOffer, err := popensea.NewOpenseaOffer(openseaOfferAddr, evmClient)
		if err != nil {
			log.Println("popensea.NewOpenseaOffer error", err)
			time.Sleep(1 * time.Second)
			continue
		}
		gasPrice, err := evmClient.SuggestGasPrice(context.Background())
		if err != nil {
			log.Println(err)
			continue
		}

		auth, err := bind.NewKeyedTransactorWithChainID(privKey, networkChainIdInt)
		if err != nil {
			log.Println(err)
			continue
		}

		gasPrice = gasPrice.Mul(gasPrice, big.NewInt(12))
		gasPrice = gasPrice.Div(gasPrice, big.NewInt(10))

		auth.GasPrice = gasPrice
		if network == "eth" {
			auth.GasLimit = wcommon.EVMGasLimitETH
		} else {
			if network == "bsc" {
				auth.GasLimit = wcommon.EVMGasLimitPancake
			}
			auth.GasLimit = wcommon.EVMGasLimit
		}

		result.Type = txType
		result.Network = network
		result.IncRequestTx = incTxHash
		_ = openseaOffer
		// address, err := wcommon.GetEVMAddress(config.EVMKey)
		// if err != nil {
		// 	log.Println(err)
		// 	continue
		// }
		// account := common.HexToAddress(address)
		// pendingNonce, _ := evmClient.PendingNonceAt(context.Background(), account)
		// auth.Nonce = new(big.Int).SetUint64(pendingNonce)

		// tx, err := evmproof.ExecuteWithBurnProof(c, auth, proof)
		// if err != nil {
		// 	log.Println(err)
		// 	if strings.Contains(err.Error(), "insufficient funds") {
		// 		return nil, errors.New("submit tx outchain failed err insufficient funds")
		// 	}
		// 	continue
		// }
		// result.Txhash = tx.Hash().String()
		// result.Status = wcommon.StatusPending
		// result.Nonce = tx.Nonce()
		break
	}

	if result.Txhash == "" {
		i++
		time.Sleep(2 * time.Second)
		goto retry
	}

	return &result, nil
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

func loadOpenseaAPIKey() {
	openseaKey, err := database.DBGetPappAPIKey("opensea")
	if err != nil {
		log.Println("DBGetPappAPIKey(opensea)", err)
		return
	}
	config.OpenSeaAPIKey = openseaKey
	openseaAPI, err := database.DBGetPappAPIKey("openseaAPI")
	if err != nil {
		log.Println("DBGetPappAPIKey(opensea)", err)
		return
	}
	config.OpenSeaAPI = openseaAPI
}

// TODO: opensea
func watchPendingOpenseaOffer() {
	for {
		time.Sleep(15 * time.Second)
		offerList, err := database.DBGetOpenseaOfferByStatus([]string{popensea.OfferStatusPending})
		if err != nil {
			log.Println("DBGetPendingOpenseaOffer err:", err)
			continue
		}
		network := "chain%7D"
		if config.NetworkID == "testnet" {
			network = "goerli"
		}
		for _, offer := range offerList {
			isTimedOut := false
			if time.Since(offer.TimeoutAt) > time.Second {
				isTimedOut = true
			}

			status, err := checkOpenseaOfferFilled(offer.OfferHash, nil)
			if err != nil {
				log.Println("checkOpenseaOfferFilled err:", err)
				continue
			}
			log.Println("checkOpenseaOfferFilled", offer.OfferTxInc, status, isTimedOut, offer.OfferSubmitted)
			//TODO: next
			switch status {
			case popensea.OfferStatusPending:
				if isTimedOut {
					//send tx cancel to opensea adapter as shield tx

				} else {
					if !offer.OfferSubmitted {
						incTxData, err := database.DBGetPappTxData(offer.OfferTxInc)
						if err != nil {
							log.Println("DBGetPappTxData err:", err)
							panic(err)
							continue
						}
						externalTxData, err := database.DBGetExternalTxByIncTx(offer.OfferTxInc, "eth")
						if err != nil {
							log.Println("DBGetPappTxData err:", err)
							panic(err)
							continue
						}
						rawTxBytes, _, err := base58.Base58Check{}.Decode(incTxData.IncTxData)
						if err != nil {
							log.Println("base58.Base58Check{}.Decode(inctxData.IncTxData) err:", err)
							panic(err)
							continue
						}

						mdRaw, _, _, _, err := extractDataFromRawTx(rawTxBytes)
						if err != nil {
							log.Println("extractDataFromRawTx err:", err)
							panic(err)
							continue
						}

						md, ok := mdRaw.(*bridge.BurnForCallRequest)
						if !ok {
							log.Println("mdRaw.(*bridge.BurnForCallRequest) invalid")
							panic(err)
							continue
						}

						openseaProxyAbi, _ := abi.JSON(strings.NewReader(popensea.OpenseaMetaData.ABI))
						openseaOfferAbi, _ := abi.JSON(strings.NewReader(popensea.OpenseaOfferMetaData.ABI))

						callData, _ := hex.DecodeString(md.Data[0].ExternalCalldata)

						signature := []byte{}
						if method, ok := openseaProxyAbi.Methods["forward"]; ok {
							params, err := method.Inputs.Unpack(callData[4:])
							if err != nil {
								log.Fatal(err)
							}
							if method2, ok := openseaOfferAbi.Methods["offer"]; ok {
								params2, err := method2.Inputs.Unpack(params[1].([]byte)[4:])
								if err != nil {
									log.Fatal(err)
								}
								signature = params2[2].([]byte)
							}
						}

						// dataOffer, err := openseaOfferAbi.Unpack("offer", dataForward[1].([]byte)[4:])
						// if err != nil {
						// 	log.Println("dataForward", len(dataForward))
						// 	log.Println("openseaOfferAbi.Unpack(forward) err:", err)
						// 	panic(err)
						// 	continue
						// }

						swapInfo := wcommon.PappSwapInfo{}

						err = json.Unmarshal([]byte(incTxData.PappSwapInfo), &swapInfo)
						if err != nil {
							log.Println("json.Unmarshal([]byte(inctxData.PappSwapInfo),&swapInfo) err:", err)
							panic(err)
							continue
						}

						offer := popensea.OrderComponents{}
						offerBytes, err := hex.DecodeString(swapInfo.AdditionalData)
						if err != nil {
							log.Println("hex.DecodeString(swapInfo.AdditionalData) err:", err)
							panic(err)
							continue
						}

						err = json.Unmarshal(offerBytes, &offer)
						if err != nil {
							log.Println("json.Unmarshal(offerBytes, &offer) err:", err)
							panic(err)
							continue
						}

						offerWithSig := popensea.OrderComponentsWithSig{
							Parameters: popensea.OrderComponentsParam{
								Offerer:                         offer.Offerer.Hex(),
								OrderType:                       int(offer.OrderType),
								Zone:                            offer.Zone.Hex(),
								ZoneHash:                        "0x" + hex.EncodeToString(offer.ZoneHash[:]),
								EndTime:                         fmt.Sprintf("%v", offer.EndTime.Int64()),
								StartTime:                       fmt.Sprintf("%v", offer.StartTime.Int64()),
								Salt:                            offer.Salt.Int64(),
								TotalOriginalConsiderationItems: len(offer.Consideration),
								ConduitKey:                      "0x" + hex.EncodeToString(offer.ConduitKey[:]),
								Nonce:                           int(externalTxData.Nonce),
								Counter:                         0,
							},
							Signature: "0x" + hex.EncodeToString(signature),
						}

						for _, o := range offer.Offer {
							offerWithSig.Parameters.Offer = append(offerWithSig.Parameters.Offer, struct {
								ItemType             int    "json:\"itemType\""
								Token                string "json:\"token\""
								IdentifierOrCriteria string "json:\"identifierOrCriteria\""
								StartAmount          string "json:\"startAmount\""
								EndAmount            string "json:\"endAmount\""
							}{ItemType: int(o.ItemType), Token: o.Token.Hex(), IdentifierOrCriteria: o.IdentifierOrCriteria.String(), StartAmount: o.StartAmount.String(), EndAmount: o.EndAmount.String()})
						}
						for _, cond := range offer.Consideration {
							offerWithSig.Parameters.Consideration = append(offerWithSig.Parameters.Consideration, struct {
								ItemType             int    "json:\"itemType\""
								Token                string "json:\"token\""
								IdentifierOrCriteria string "json:\"identifierOrCriteria\""
								StartAmount          string "json:\"startAmount\""
								EndAmount            string "json:\"endAmount\""
								Recipient            string "json:\"recipient\""
							}{
								ItemType: int(cond.ItemType), Token: cond.Token.Hex(), IdentifierOrCriteria: cond.IdentifierOrCriteria.String(), StartAmount: cond.StartAmount.String(), EndAmount: cond.EndAmount.String(), Recipient: cond.Recipient.Hex(),
							})
						}

						offerWithSigJson, _ := json.Marshal(offerWithSig)
						log.Println("offerWithSig", string(offerWithSigJson))

						err = popensea.SubmitOpenseaOffer(config.OpenSeaAPI, config.OpenSeaAPIKey, network, offerWithSig)
						if err != nil {
							log.Println("popensea.SubmitOpenseaOffer err:", err)
							panic(err)
							continue
						}
						panic(err)
					}
				}
			case popensea.OfferStatusFilled:
				err = database.DBUpdateOpenseaOfferStatus(offer.OfferTxInc, popensea.OfferStatusFilled)
				if err != nil {
					log.Println("DBUpdateOpenseaOfferStatus err:", err)
					continue
				}
			case popensea.OfferStatusCancelled:
				err = database.DBUpdateOpenseaOfferStatus(offer.OfferTxInc, popensea.OfferStatusCancelled)
				if err != nil {
					log.Println("DBUpdateOpenseaOfferStatus err:", err)
					continue
				}
			}
		}
	}
}

func watchFilledOpenseaOffer() {
	for {
		time.Sleep(15 * time.Second)
		offerList, err := database.DBGetOpenseaOfferByStatus([]string{popensea.OfferStatusFilled})
		if err != nil {
			log.Println("DBGetPendingOpenseaOffer err:", err)
			continue
		}

		for _, offer := range offerList {
			offerClaim := offer.OfferTxInc + "_claim"
			_, err := SubmitOutChainTx(offerClaim, "eth", false, false, wcommon.ExternalTxTypeOpenseaOfferClaim)
			if err != nil {
				log.Println("SubmitOutChainTx err:", err)
				continue
			}
		}
	}
}

func checkOpenseaOfferFilled(orderHash string, order *popensea.OrderComponents) (string, error) {
	status := "unknown"
	papps, err := database.DBRetrievePAppsByNetwork("eth")
	if err != nil {
		return status, err
	}
	if len(papps.AppContracts) == 0 {
		return status, errors.New("papps is empty")
	}

	seaportAddress, exist := papps.AppContracts["seaport"]
	if !exist {
		return status, errors.New("seaport isn't exist")
	}

	openseaOfferer, exist := papps.AppContracts["opensea-offer"]
	if !exist {
		return status, errors.New("seaport isn't exist")
	}

	networkInfo, err := database.DBGetBridgeNetworkInfo("eth")
	if err != nil {
		return status, err
	}

	for _, endpoint := range networkInfo.Endpoints {
		evmClient, err := ethclient.Dial(endpoint)
		if err != nil {
			log.Println("ethclient.Dial err:", err)
			continue
		}

		seaport, err := popensea.NewIopensea(common.HexToAddress(seaportAddress), evmClient)
		if err != nil {
			log.Println("NewIopensea err:", err)
			return status, err
		}

		openseaOfferAddr := common.HexToAddress(openseaOfferer)
		offerAdapter, err := popensea.NewOpenseaOffer(openseaOfferAddr, evmClient)
		if err != nil {
			log.Println("NewOpenseaOffer err:", err)
			return status, err
		}
		if orderHash != "" {
			orderHashBytes, err := hex.DecodeString(orderHash)
			if err != nil {
				log.Println("hex.DecodeString(orderHash) err:", err)
				return status, err
			}
			orderStatus, err := seaport.GetOrderStatus(nil, toByte32(orderHashBytes))
			if err != nil {
				log.Println("seaport.GetOrderStatus err:", err)
				return status, err
			}
			if orderStatus.IsCancelled {
				status = popensea.OfferStatusCancelled
				return status, nil
			} else {
				if orderStatus.TotalFilled.Int64() > 0 {
					status = popensea.OfferStatusFilled
					return status, nil
				}
				status = popensea.OfferStatusPending
				return status, nil
			}
		} else {
			//TODO: opensea ?
			seaport.GetOrderHash(nil, *order)
			_ = offerAdapter
		}
	}
	return status, nil
}

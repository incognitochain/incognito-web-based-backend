package interswap

import (
	"context"
	"log"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	"github.com/incognitochain/go-incognito-sdk-v2/wallet"

	"github.com/incognitochain/incognito-web-based-backend/utxomanager"

	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
)

var config wcommon.Config
var InterswapIncKeySets map[string]*wallet.KeyWallet
var UtxoManager *utxomanager.UTXOManager

func StartWorker(cfg wcommon.Config, serviceID uuid.UUID) error {
	network := cfg.NetworkID
	config = cfg

	// // start client
	// err := startPubsubClient(cfg.GGCProject, cfg.GGCAuth)
	// if err != nil {
	// 	return err
	// }

	// // init topic instance
	// interSwapTxTopic, err = startPubsubTopic(cfg.NetworkID + "_" + INTERSWAP_TX_TOPIC)
	// if err != nil {
	// 	panic(err)
	// }

	// init incognito client
	err := InitIncClient(network, cfg)
	if err != nil {
		return err
	}
	UtxoManager = utxomanager.NewUTXOManager(incClient)
	log.Printf("UtxoManager len: %v\n", UtxoManager.Caches)

	// submit OTA key to fullnode
	if len(cfg.ISIncPrivKeys) > 0 {
		for _, key := range cfg.ISIncPrivKeys {
			wl, err := wallet.Base58CheckDeserialize(key)
			if err != nil {
				panic(err)
			}
			if cfg.FullnodeAuthKey != "" {
				err = incClient.AuthorizedSubmitKey(wl.Base58CheckSerialize(wallet.OTAKeyType), cfg.FullnodeAuthKey, 0, false)
				if err != nil {
					if !strings.Contains(err.Error(), "has been submitted") {
						return err
					}
				}
			} else {
				err = incClient.SubmitKey(wl.Base58CheckSerialize(wallet.OTAKeyType))
				if err != nil {
					if !strings.Contains(err.Error(), "has been submitted") {
						return err
					}
				}
			}
		}

	}

	go watchInterswapPendingTx(config)

	incclient.Logger = incclient.NewLogger(true)
	log.Println("Done submit keys")

	// // init subscription
	// var interswapSub *pubsub.Subscription
	// interswapSubID := cfg.NetworkID + "_" + serviceID.String() + "_interswap"
	// interswapSub, err = psclient.CreateSubscription(context.Background(), interswapSubID,
	// 	pubsub.SubscriptionConfig{Topic: interSwapTxTopic})
	// if err != nil {
	// 	interswapSub = psclient.Subscription(interswapSubID)
	// }
	// log.Println("interswapSub.ID()", interswapSub.ID())

	// start subscription to receive msg and req workers execute something
	// go func() {
	// 	ctx := context.Background()
	// 	err := interswapSub.Receive(ctx, ProcessInterswapTxRequest)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }()

	return nil
}

func ProcessInterswapTxRequest(ctx context.Context, m *pubsub.Message) {
	taskDesc := m.Attributes["task"]
	switch taskDesc {
	case InterswapPdexPappTxTask:
		processInterswapPdexPappPathTask(ctx, m)
	}
}

type InterswapSubmitTxTask struct {
	TxID                string `json:"txid" bson:"txid"`
	TxRawBytes          []byte `json:"txraw" bson:"txraw"`
	FromToken           string `json:"fromtoken" bson:"fromtoken"`
	ToToken             string `json:"totoken" bson:"totoken"`
	MidToken            string `json:"midtoken" bson:"midtoken"`
	PathType            int    `json:"pathtype" bson:"pathtype"`
	FinalMinExpectedAmt uint64 `json:"final_minacceptedamount" bson:"final_minacceptedamount"`

	OTARefundFee string `json:"ota_refundfee" bson:"ota_refundfee"`
	OTARefund    string `json:"ota_refund" bson:"ota_refund"`
	OTAFromToken string `json:"ota_fromtoken" bson:"ota_fromtoken"`
	OTAToToken   string `json:"ota_totoken" bson:"ota_totoken"`

	AddOnTxID    string `json:"addon_txid" bson:"addon_txid"`
	PAppName     string `json:"papp_name" bson:"papp_name"`
	PAppNetwork  string `json:"papp_network" bson:"papp_network"`
	PAppContract string `json:"papp_contract" bson:"papp_contract"`

	Status    int    `json:"status" bson:"status"`
	StatusStr string `json:"statusstr" bson:"statusstr"`
	UserAgent string `json:"useragent" bson:"useragent"`
	Error     string `json:"error" bson:"error"`
}

func processInterswapPdexPappPathTask(ctx context.Context, m *pubsub.Message) {
	// task := InterswapSubmitTxTask{}
	// err := json.Unmarshal(m.Data, &task)
	// if err != nil {
	// 	log.Println("processInterswapPathType1Task error decoding message", err)
	// 	m.Ack()
	// 	return
	// }

	// // get tx by hash
	// txDetail, err := incClient.GetTxDetail(task.TxID)
	// if err != nil {
	// 	log.Println("GetTxDetail err", err)
	// 	if err != nil {
	// 		log.Println(err)
	// 		m.Nack()
	// 		return
	// 	}
	// } else {
	// 	if txDetail.IsInBlock {
	// 		err = database.DBUpdateInterswapTxStatus(task.TxID, FirstInBlock, StatusStr[FirstInBlock], "")
	// 		if err != nil {
	// 			log.Println(err)
	// 			m.Nack()
	// 			return
	// 		}
	// 	}
	// }

	// // get swap tx status by calling api
	// _, pdexStatus, err := CallGetPdexSwapTxStatus(task.TxID, task.MidToken)
	// if err != nil {
	// 	log.Println(err)
	// 	m.Nack()
	// 	return
	// }
	// if len(pdexStatus.RespondTxs) > 1 {
	// 	if pdexStatus.Status == "accepted" {
	// 		// parse tx response to get received UTXO
	// 		if len(pdexStatus.RespondTxs) != 1 {

	// 		}
	// 		amtMidToken := pdexStatus.RespondAmounts[0]

	// 		// validate

	// 		// re-estimate with addon tx

	// 		// create addon tx
	// 		data := metadataBridge.BurnForCallRequestData{}
	// 		incClient.CreateAndSendBurnForCallRequestTransaction(config.ISIncPrivKey, task.MidToken)

	// 		// // update addon swap info: amountFrom
	// 		// updatedAddonSwapInfo := task.AddOnSwapInfo

	// 		// // re-calculate AmountIn for AddOn tx
	// 		// midTokenAmt := pdexStatus.RespondAmounts[0]
	// 		// amountStrMidToken := convertToWithoutDecStr(pdexStatus.RespondAmounts[0], pdexStatus.RespondTokens[0])

	// 		// updatedAddonSwapInfo.AmountIn = amountStrMidToken
	// 		// updatedAddonSwapInfo.AmountInRaw = pdexStatus.RespondAmounts[0]

	// 		// // check minAcceptedAmount of AddOn tx is still valid or not

	// 	} else if pdexStatus.Status == "refund" {

	// 	} else {

	// 	}

	// }

	// get swap tx status by calling api
	// database.DBGetPappTxStatus()

	// pappSwapInfoStr, _ := json.MarshalIndent(task.PappSwapInfo, "", "\t")
	// data := wcommon.PappTxData{
	// 	IncTx:            task.TxHash,
	// 	IncTxData:        string(task.TxRawData),
	// 	Type:             wcommon.PappTypeSwap,
	// 	Status:           wcommon.StatusSubmitting,
	// 	IsUnifiedToken:   task.IsUnifiedToken,
	// 	FeeToken:         task.FeeToken,
	// 	FeeAmount:        task.FeeAmount,
	// 	PFeeAmount:       task.PFeeAmount,
	// 	BurntToken:       task.BurntToken,
	// 	BurntAmount:      task.BurntAmount,
	// 	PappSwapInfo:     string(pappSwapInfoStr),
	// 	Networks:         task.Networks,
	// 	FeeRefundOTA:     task.FeeRefundOTA,
	// 	FeeRefundAddress: task.FeeRefundAddress,
	// 	OutchainStatus:   wcommon.StatusWaiting,
	// 	UserAgent:        task.UserAgent,
	// }
	// docID, err := database.DBSavePappTxData(data)
	// if err != nil {
	// 	writeErr, ok := err.(mongo.WriteException)
	// 	if !ok {
	// 		log.Println("DBAddPappTxData err", err)
	// 		m.Nack()
	// 		return
	// 	}
	// 	if !writeErr.HasErrorCode(11000) {
	// 		log.Println("DBAddPappTxData err", err)
	// 		m.Nack()
	// 		return
	// 	}
	// }

	// txDetail, err := incClient.GetTxDetail(task.TxHash)
	// if err != nil {
	// 	log.Println("GetTxDetail err", err)
	// } else {
	// 	if txDetail.IsInMempool {
	// 		err = database.DBUpdatePappTxStatus(task.TxHash, wcommon.StatusPending, "")
	// 		if err != nil {
	// 			log.Println(err)
	// 			m.Nack()
	// 			return
	// 		}
	// 	}
	// 	if txDetail.IsInBlock {
	// 		err = database.DBUpdatePappTxStatus(task.TxHash, wcommon.StatusExecuting, "")
	// 		if err != nil {
	// 			log.Println(err)
	// 			m.Nack()
	// 			return
	// 		}
	// 	}
	// 	m.Ack()
	// 	return
	// }

	// var errSubmit error

	// if task.IsPRVTx {
	// 	errSubmit = incClient.SendRawTx(task.TxRawData)
	// } else {
	// 	errSubmit = incClient.SendRawTokenTx(task.TxRawData)
	// }

	// if errSubmit != nil {
	// 	err = database.DBUpdatePappTxStatus(task.TxHash, wcommon.StatusSubmitFailed, errSubmit.Error())
	// 	if err != nil {
	// 		log.Println(err)
	// 		m.Nack()
	// 		return
	// 	}
	// 	go slacknoti.SendSlackNoti(fmt.Sprintf("`[swaptx]` submit swaptx failed 😵 `%v`", task.TxHash))
	// 	m.Ack()
	// 	return
	// } else {
	// 	err = database.DBUpdatePappTxStatus(task.TxHash, wcommon.StatusPending, "")
	// 	if err != nil {
	// 		log.Println(err)
	// 		m.Nack()
	// 		return
	// 	}
	// }
	// go func() {
	// 	slackep := os.Getenv("SLACK_SWAP_ALERT")
	// 	if slackep != "" {
	// 		swapAlert := ""
	// 		pappTxData := data
	// 		if pappTxData.PappSwapInfo != "" {
	// 			networkID := wcommon.GetNetworkID(task.Networks[0])
	// 			tkInInfo, _ := getTokenInfo(task.PappSwapInfo.TokenIn)
	// 			amount := new(big.Float).SetInt(task.PappSwapInfo.TokenInAmount)
	// 			decimal := new(big.Float)
	// 			decimalInt, err := getTokenDecimalOnNetwork(tkInInfo, networkID)
	// 			if err != nil {
	// 				log.Println("getTokenDecimalOnNetwork1", err)
	// 				return
	// 			}
	// 			decimal.SetFloat64(math.Pow10(int(-decimalInt)))

	// 			amountInFloat := amount.Mul(amount, decimal).Text('f', -1)
	// 			tokenInSymbol := tkInInfo.Symbol

	// 			tkOutInfo, _ := getTokenInfo(task.PappSwapInfo.TokenOut)
	// 			amount = new(big.Float).SetInt(task.PappSwapInfo.MinOutAmount)

	// 			decimalInt, err = getTokenDecimalOnNetwork(tkOutInfo, networkID)
	// 			if err != nil {
	// 				log.Println("getTokenDecimalOnNetwork2", err)
	// 				return
	// 			}
	// 			decimal.SetFloat64(math.Pow10(int(-decimalInt)))
	// 			amountOutFloat := amount.Mul(amount, decimal).Text('f', -1)
	// 			tokenOutSymbol := tkOutInfo.Symbol

	// 			uaStr := parseUserAgent(task.UserAgent)

	// 			swapAlert = fmt.Sprintf("`[%v(%v) | %v]` swap submitting 🛰\n SwapID: `%v`\n Requested: `%v %v` to `%v %v`\n--------------------------------------------------------", task.PappSwapInfo.DappName, pappTxData.Networks[0], uaStr, docID.Hex(), amountInFloat, tokenInSymbol, amountOutFloat, tokenOutSymbol)
	// 			log.Println(swapAlert)
	// 			slacknoti.SendWithCustomChannel(swapAlert, slackep)
	// 		}
	// 	}
	// }()

	m.Ack()
	return
}

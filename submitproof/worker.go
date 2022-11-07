package submitproof

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"github.com/incognitochain/bridge-eth/bridge/vault"
	inccommon "github.com/incognitochain/go-incognito-sdk-v2/common"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	"github.com/incognitochain/go-incognito-sdk-v2/wallet"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/evmproof"
	"github.com/incognitochain/incognito-web-based-backend/slacknoti"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

func StartWorker(keylist []string, cfg wcommon.Config, serviceID uuid.UUID) error {
	config = cfg
	keyList = keylist
	network := cfg.NetworkID

	err := startPubsubClient(cfg.GGCProject, cfg.GGCAuth)
	if err != nil {
		return err
	}

	shieldTxTopic, err = startPubsubTopic(cfg.NetworkID + "_" + SHIELD_TX_TOPIC)
	if err != nil {
		panic(err)
	}

	pappTxTopic, err = startPubsubTopic(cfg.NetworkID + "_" + PAPP_TX_TOPIC)
	if err != nil {
		panic(err)
	}

	err = initIncClient(network)
	if err != nil {
		return err
	}

	if len(keyList) == 0 {
		return errors.New("no keys")
	}

	for _, v := range keyList {
		wl, err := wallet.Base58CheckDeserialize(v)
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

	if config.IncKey != "" {
		wl, err := wallet.Base58CheckDeserialize(config.IncKey)
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

		// err = genShardsAccount(config.IncKey)
		// if err != nil {
		// 	return err
		// }

		// for _, v := range incShardsAccount {
		// 	wl, err := wallet.Base58CheckDeserialize(v)
		// 	if err != nil {
		// 		panic(err)
		// 	}
		// 	err = incClient.SubmitKey(wl.Base58CheckSerialize(wallet.OTAKeyType))
		// 	if err != nil {
		// 		return err
		// 	}
		// }
	}

	incclient.Logger = incclient.NewLogger(true)
	log.Println("Done submit keys")

	var shieldSub *pubsub.Subscription
	shieldSubID := cfg.NetworkID + "_" + serviceID.String() + "_shield"
	shieldSub, err = psclient.CreateSubscription(context.Background(), shieldSubID,
		pubsub.SubscriptionConfig{Topic: shieldTxTopic})
	if err != nil {
		shieldSub = psclient.Subscription(shieldSubID)
	}
	log.Println("shieldSub.ID()", shieldSub.ID())

	var pappSub *pubsub.Subscription
	pappSubID := cfg.NetworkID + "_" + serviceID.String() + "_papp"
	pappSub, err = psclient.CreateSubscription(context.Background(), pappSubID,
		pubsub.SubscriptionConfig{Topic: pappTxTopic})
	if err != nil {
		pappSub = psclient.Subscription(pappSubID)
	}
	log.Println("pappSub.ID()", pappSub.ID())

	go func() {
		ctx := context.Background()
		err := shieldSub.Receive(ctx, ProcessShieldRequest)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		ctx := context.Background()
		err := pappSub.Receive(ctx, ProcessPappTxRequest)
		if err != nil {
			panic(err)
		}
	}()

	return nil
}

func ProcessShieldRequest(ctx context.Context, m *pubsub.Message) {
	task := SubmitProofShieldTask{}
	defer m.Ack()
	err := json.Unmarshal(m.Data, &task)
	if err != nil {
		// rejectDelivery(delivery, payload)
		log.Println("ProcessShieldRequest error decoding message", err)
		return
	}
	if time.Since(m.PublishTime) > time.Hour {
		errdb := database.DBUpdateShieldTxStatus(task.TxHash, task.NetworkID, wcommon.StatusSubmitFailed, "timeout")
		if errdb != nil {
			log.Println("DBUpdateShieldTxStatus error:", errdb)
			return
		}
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[shieldtx]` shield/redeposit timeout 😵 exttx `%v` network `%v`\n", task.TxHash, task.NetworkID))
		return
	}

	t := time.Now().Unix()
	key := keyList[t%int64(len(keyList))]
	incTx, paymentAddr, tokenID, linkedTokenID, err := submitProof(task.TxHash, task.TokenID, task.NetworkID, key)
	if err != nil {
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[submitProof]` create tx failed `%v`, tokenID `%v`, networkID `%v`, error: `%v`\n", task.TxHash, task.TokenID, task.NetworkID, err))
		errdb := database.DBUpdateShieldTxStatus(task.TxHash, task.NetworkID, wcommon.StatusSubmitFailed, err.Error())
		if errdb != nil {
			log.Println("DBUpdateShieldTxStatus error:", errdb)
			return
		}
		return
	}

	errdb := database.DBUpdateShieldOnChainTxInfo(task.TxHash, task.NetworkID, paymentAddr, incTx, tokenID, linkedTokenID)
	if errdb != nil {
		log.Println("DBUpdateShieldOnChainTxInfo error:", err)
		return
	}
	err = database.DBUpdateExternalTxSubmitedRedeposit(task.TxHash, true)
	if err != nil {
		log.Println("DBUpdateExternalTxSubmitedRedeposit error:", err)
		return
	}

	err = database.DBUpdateShieldTxStatus(task.TxHash, task.NetworkID, wcommon.StatusPending, "")
	if err != nil {
		log.Println("DBUpdateShieldTxStatus err:", err)
		return
	}
}

func ProcessPappTxRequest(ctx context.Context, m *pubsub.Message) {
	taskDesc := m.Attributes["task"]
	switch taskDesc {
	case PappSubmitIncTask:
		processSubmitPappIncTask(ctx, m)
	case PappSubmitExtTask:
		processSubmitPappExtTask(ctx, m)
	case PappSubmitFeeRefundTask:
		processSubmitRefundFeeTask(ctx, m)
	}
}

func processSubmitPappExtTask(ctx context.Context, m *pubsub.Message) {
	task := SubmitPappProofOutChainTask{}
	err := json.Unmarshal(m.Data, &task)
	if err != nil {
		log.Println("processSubmitPappExtTask error decoding message", err)
		m.Ack()
		return
	}

	if time.Since(m.PublishTime) > time.Hour {
		status := wcommon.ExternalTxStatus{
			IncRequestTx: task.IncTxhash,
			Type:         wcommon.PappTypeSwap,
			Status:       wcommon.StatusSubmitFailed,
			Network:      task.Network,
			Error:        "timeout",
		}
		err = database.DBSaveExternalTxStatus(&status)
		if err != nil {
			writeErr, ok := err.(mongo.WriteException)
			if !ok {
				log.Println("DBSaveExternalTxStatus err", err)
				m.Nack()
				return
			}
			if !writeErr.HasErrorCode(11000) {
				log.Println("DBSaveExternalTxStatus err", err)
				m.Nack()
				return
			}
		}
		err = database.DBUpdatePappTxSubmitOutchainStatus(task.IncTxhash, wcommon.StatusSubmitFailed)
		if err != nil {
			writeErr, ok := err.(mongo.WriteException)
			if !ok {
				log.Println("DBSaveExternalTxStatus err", err)
				m.Nack()
				return
			}
			if !writeErr.HasErrorCode(11000) {
				log.Println("DBSaveExternalTxStatus err", err)
				m.Nack()
				return
			}
		}
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[swaptx]` submitProofTx timeout 😵 inctx `%v` network `%v`\n", task.IncTxhash, task.Network))
		return
	}

	_, err = database.DBGetExternalTxStatusByIncTx(task.IncTxhash, task.Network)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Println("DBGetExternalTxStatusByIncTx err", err)
			m.Nack()
			return
		}
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[swaptx]` submitProofTx `%v` for network `%v`", task.IncTxhash, task.Network))
	} else {
		if task.IsRetry {
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[swaptx]` retry submitProofTx `%v` for network `%v` 🫡", task.IncTxhash, task.Network))
		}
	}

	status, err := createOutChainSwapTx(task.Network, task.IncTxhash, task.IsUnifiedToken)
	if err != nil {
		log.Println("createOutChainSwapTx error", err)
		time.Sleep(15 * time.Second)
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[swaptx]` submitProofTx `%v` for network `%v` failed 😵 err: %v", task.IncTxhash, task.Network, err))
		m.Ack()
		return
	}
	go slacknoti.SendSlackNoti(fmt.Sprintf("`[swaptx]` submitProofTx `%v` for network `%v` success 👌 txhash `%v`", task.IncTxhash, task.Network, status.Txhash))

	err = database.DBSaveExternalTxStatus(status)
	if err != nil {
		writeErr, ok := err.(mongo.WriteException)
		if !ok {
			log.Println("DBSaveExternalTxStatus err", err)
			m.Ack()
			return
		}
		if !writeErr.HasErrorCode(11000) {
			log.Println("DBSaveExternalTxStatus err", err)
			m.Ack()
			return
		}
	}

	err = database.DBUpdatePappTxSubmitOutchainStatus(task.IncTxhash, wcommon.StatusPending)
	if err != nil {
		log.Println("DBUpdatePappTxSubmitOutchainStatus err", err)
		m.Ack()
		return
	}

	m.Ack()
}

func processSubmitPappIncTask(ctx context.Context, m *pubsub.Message) {
	task := SubmitPappTxTask{}
	err := json.Unmarshal(m.Data, &task)
	if err != nil {
		log.Println("processSubmitPappIncTask error decoding message", err)
		m.Ack()
		return
	}
	pappSwapInfoStr, _ := json.MarshalIndent(task.PappSwapInfo, "", "\t")
	data := wcommon.PappTxData{
		IncTx:            task.TxHash,
		IncTxData:        string(task.TxRawData),
		Type:             wcommon.PappTypeSwap,
		Status:           wcommon.StatusSubmitting,
		IsUnifiedToken:   task.IsUnifiedToken,
		FeeToken:         task.FeeToken,
		FeeAmount:        task.FeeAmount,
		PFeeAmount:       task.PFeeAmount,
		BurntToken:       task.BurntToken,
		BurntAmount:      task.BurntAmount,
		PappSwapInfo:     string(pappSwapInfoStr),
		Networks:         task.Networks,
		FeeRefundOTA:     task.FeeRefundOTA,
		FeeRefundAddress: task.FeeRefundAddress,
		OutchainStatus:   wcommon.StatusWaiting,
		UserAgent:        task.UserAgent,
	}
	docID, err := database.DBSavePappTxData(data)
	if err != nil {
		writeErr, ok := err.(mongo.WriteException)
		if !ok {
			log.Println("DBAddPappTxData err", err)
			m.Nack()
			return
		}
		if !writeErr.HasErrorCode(11000) {
			log.Println("DBAddPappTxData err", err)
			m.Nack()
			return
		}
	}

	txDetail, err := incClient.GetTxDetail(task.TxHash)
	if err != nil {
		log.Println("GetTxDetail err", err)
	} else {
		if txDetail.IsInMempool {
			err = database.DBUpdatePappTxStatus(task.TxHash, wcommon.StatusPending, "")
			if err != nil {
				log.Println(err)
				m.Nack()
				return
			}
		}
		if txDetail.IsInBlock {
			err = database.DBUpdatePappTxStatus(task.TxHash, wcommon.StatusExecuting, "")
			if err != nil {
				log.Println(err)
				m.Nack()
				return
			}
		}
		m.Ack()
		return
	}

	var errSubmit error

	if task.IsPRVTx {
		errSubmit = incClient.SendRawTx(task.TxRawData)
	} else {
		errSubmit = incClient.SendRawTokenTx(task.TxRawData)
	}

	if errSubmit != nil {
		err = database.DBUpdatePappTxStatus(task.TxHash, wcommon.StatusSubmitFailed, errSubmit.Error())
		if err != nil {
			log.Println(err)
			m.Nack()
			return
		}
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[swaptx]` submit swaptx failed 😵 `%v`", task.TxHash))
		m.Ack()
		return
	} else {
		err = database.DBUpdatePappTxStatus(task.TxHash, wcommon.StatusPending, "")
		if err != nil {
			log.Println(err)
			m.Nack()
			return
		}
	}
	go func() {
		slackep := os.Getenv("SLACK_SWAP_ALERT")
		if slackep != "" {
			swapAlert := ""
			pappTxData := data
			if pappTxData.PappSwapInfo != "" {
				networkID := wcommon.GetNetworkID(task.Networks[0])
				tkInInfo, _ := getTokenInfo(task.PappSwapInfo.TokenIn)
				amount := new(big.Float).SetInt(task.PappSwapInfo.TokenInAmount)
				decimal := new(big.Float)
				decimalInt, err := getTokenDecimalOnNetwork(tkInInfo, networkID)
				if err != nil {
					log.Println("getTokenDecimalOnNetwork1", err)
					return
				}
				decimal.SetFloat64(math.Pow10(int(-decimalInt)))

				amountInFloat := amount.Mul(amount, decimal).Text('f', -1)
				tokenInSymbol := tkInInfo.Symbol

				tkOutInfo, _ := getTokenInfo(task.PappSwapInfo.TokenOut)
				amount = new(big.Float).SetInt(task.PappSwapInfo.MinOutAmount)

				decimalInt, err = getTokenDecimalOnNetwork(tkOutInfo, networkID)
				if err != nil {
					log.Println("getTokenDecimalOnNetwork2", err)
					return
				}
				decimal.SetFloat64(math.Pow10(int(-decimalInt)))
				amountOutFloat := amount.Mul(amount, decimal).Text('f', -1)
				tokenOutSymbol := tkOutInfo.Symbol

				uaStr := parseUserAgent(task.UserAgent)

				swapAlert = fmt.Sprintf("`[%v | %v]` swap submitting 🛰\n SwapID: `%v`\n Requested: `%v %v` to `%v %v`\n--------------------------------------------------------", task.PappSwapInfo.DappName, uaStr, docID.Hex(), amountInFloat, tokenInSymbol, amountOutFloat, tokenOutSymbol)
				log.Println(swapAlert)
				slacknoti.SendWithCustomChannel(swapAlert, slackep)
			}
		}
	}()

	m.Ack()
}

func createOutChainSwapTx(network string, incTxHash string, isUnifiedToken bool) (*wcommon.ExternalTxStatus, error) {
	var result wcommon.ExternalTxStatus

	// networkID := wcommon.GetNetworkID(network)
	networkInfo, err := database.DBGetBridgeNetworkInfo(network)
	if err != nil {
		return nil, err
	}

	pappAddress, err := database.DBGetPappVaultData(network, wcommon.PappTypeSwap)
	if err != nil {
		return nil, err
	}

	networkChainId := networkInfo.ChainID

	networkChainIdInt := new(big.Int)
	networkChainIdInt.SetString(networkChainId, 10)

	var proof *evmproof.DecodedProof
	// if isUnifiedToken {
	proof, err = evmproof.GetAndDecodeBurnProofUnifiedToken(config.FullnodeURL, incTxHash, 0)
	// } else {
	// 	proof, err = evmproof.GetAndDecodeBurnProofV2(config.FullnodeURL, incTxHash, "getburnplgprooffordeposittosc")
	// }
	if err != nil {
		return nil, err
	}
	if proof == nil {
		return nil, fmt.Errorf("could not get proof for network %s", networkChainId)
	}

	if len(proof.InstRoots) == 0 {
		return nil, fmt.Errorf("could not get proof for network %s", networkChainId)
	}

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

		c, err := vault.NewVault(common.HexToAddress(pappAddress.ContractAddress), evmClient)
		if err != nil {
			log.Println(err)
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

		result.Type = wcommon.PappTypeSwap
		result.Network = network
		result.IncRequestTx = incTxHash

		tx, err := evmproof.ExecuteWithBurnProof(c, auth, proof)
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

func processSubmitRefundFeeTask(ctx context.Context, m *pubsub.Message) {
	task := SubmitRefundFeeTask{}
	err := json.Unmarshal(m.Data, &task)
	if err != nil {
		log.Println("processSubmitRefundFeeTask error decoding message", err)
		m.Ack()
		return
	}
	i := 0
	defer m.Ack()
	go slacknoti.SendSlackNoti(fmt.Sprintf("`[refundfee]` Need refund fee for tx `%v`\n", task.IncReqTx))
retry:
	i++
	var errSubmit error
	var txhash string
	var txRaw []byte
	if i == 10 {
		errStr := ""
		if errSubmit != nil {
			errStr = errSubmit.Error()
		}
		err = database.DBUpdateRefundFeeRefundTx(txhash, task.IncReqTx, wcommon.StatusSubmitFailed, errStr)
		if err != nil {
			log.Println("DBUpdateRefundFeeRefundTx error ", err)
			return
		}
	}

	if task.Token != inccommon.PRVCoinID.String() {
		var tokenParam *incclient.TxTokenParam
		if task.PaymentAddress != "" {
			tokenParam = incclient.NewTxTokenParam(task.Token, 1, []string{task.PaymentAddress}, []uint64{task.Amount}, false, 0, nil)
		} else {
			tokenParam = incclient.NewTxTokenParam(task.Token, 1, []string{task.OTA}, []uint64{task.Amount}, false, 0, nil)
		}

		txParam := incclient.NewTxParam(config.IncKey, []string{}, []uint64{}, 100, tokenParam, nil, nil)

		txRaw, txhash, err = incClient.CreateRawTokenTransactionVer2(txParam)
		if err != nil {
			log.Println("CreateRawTokenTransactionVer2 error ", err)
			errSubmit = err
			goto retry
		}
		err = incClient.SendRawTokenTx(txRaw)
		if err != nil {
			log.Println("SendRawTokenTx error ", err)
			errSubmit = err
			goto retry
		}
	} else {
		var txParam *incclient.TxParam

		if task.PaymentAddress != "" {
			txParam = incclient.NewTxParam(config.IncKey, []string{task.PaymentAddress}, []uint64{task.Amount}, 0, nil, nil, nil)
		} else {
			txParam = incclient.NewTxParam(config.IncKey, []string{task.OTA}, []uint64{task.Amount}, 0, nil, nil, nil)
		}

		txRaw, txhash, err = incClient.CreateRawTransactionVer2(txParam)
		if err != nil {
			log.Println("CreateRawTransactionVer2 error ", err)
			errSubmit = err
			goto retry
		}
		err = incClient.SendRawTx(txRaw)
		if err != nil {
			log.Println("SendRawTx error ", err)
			errSubmit = err
			goto retry
		}
	}

	if errSubmit != nil {
		log.Println("processSubmitRefundFeeTask error ", errSubmit)
		time.Sleep(5 * time.Second)
		goto retry
	} else {
	retrySaved:
		err = database.DBUpdateRefundFeeRefundTx(txhash, task.IncReqTx, wcommon.StatusPending, "")
		if err != nil {
			time.Sleep(5 * time.Second)
			goto retrySaved
		}
	}
	go slacknoti.SendSlackNoti(fmt.Sprintf("`[refundfee]` refund fee tx submitted, `%v`, requestTx `%v`, isPrivacyFee `%v`\n", txhash, task.IncReqTx, task.IsPrivacyFeeRefund))

}
func speedupOutChainSwapTx(network int, evmTxHash string) error {
	//TODO
	return nil
}

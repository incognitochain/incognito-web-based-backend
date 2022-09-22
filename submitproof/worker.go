package submitproof

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"github.com/incognitochain/bridge-eth/bridge/vault"
	"github.com/incognitochain/go-incognito-sdk-v2/coin"
	inccommon "github.com/incognitochain/go-incognito-sdk-v2/common"
	inccrypto "github.com/incognitochain/go-incognito-sdk-v2/crypto"
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

	shieldTxTopic, err = startPubsubTopic(SHIELD_TX_TOPIC)
	if err != nil {
		panic(err)
	}

	pappTxTopic, err = startPubsubTopic(PAPP_TX_TOPIC)
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
		err = incClient.SubmitKey(wl.Base58CheckSerialize(wallet.OTAKeyType))
		if err != nil {
			return err
		}
	}

	if config.IncKey != "" {
		wl, err := wallet.Base58CheckDeserialize(config.IncKey)
		if err != nil {
			panic(err)
		}
		err = incClient.SubmitKey(wl.Base58CheckSerialize(wallet.OTAKeyType))
		if err != nil {
			return err
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
	shieldSub, err = psclient.CreateSubscription(context.Background(), serviceID.String()+"_shield",
		pubsub.SubscriptionConfig{Topic: shieldTxTopic})
	if err != nil {
		shieldSub = psclient.Subscription(serviceID.String() + "_shield")
	}
	log.Println("shieldSub.ID()", shieldSub.ID())

	var pappSub *pubsub.Subscription
	pappSub, err = psclient.CreateSubscription(context.Background(), serviceID.String()+"_papp",
		pubsub.SubscriptionConfig{Topic: pappTxTopic})
	if err != nil {
		pappSub = psclient.Subscription(serviceID.String() + "_papp")
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

	if time.Since(task.Time) > time.Hour {
		errdb := database.DBUpdateShieldTxStatus(task.TxHash, task.NetworkID, wcommon.StatusSubmitFailed, "timeout")
		if errdb != nil {
			log.Println("DBUpdateShieldTxStatus error:", errdb)
			return
		}
		return
	}

	t := time.Now().Unix()
	key := keyList[t%int64(len(keyList))]
	incTx, paymentAddr, tokenID, linkedTokenID, err := submitProof(task.TxHash, task.TokenID, task.NetworkID, key)

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
	// }

	if err != nil {
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[submitProof]` txhash `%v`, tokenID `%v`, networkID `%v`, error: `%v`\n", task.TxHash, task.TokenID, task.NetworkID, err))
		errdb := database.DBUpdateShieldTxStatus(task.TxHash, task.NetworkID, wcommon.StatusSubmitFailed, err.Error())
		if errdb != nil {
			log.Println("DBUpdateShieldTxStatus error:", errdb)
			return
		}
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

	status, err := createOutChainSwapTx(task.Network, task.IncTxhash, task.IsUnifiedToken)
	if err != nil {
		log.Println("createOutChainSwapTx error", err)
		m.Ack()
		return
	}

	err = database.DBSaveExternalTxStatus(status)
	if err != nil {
		writeErr, ok := err.(mongo.WriteException)
		if !ok {
			log.Println("DBAddPappTxData err", err)
			m.Ack()
			return
		}
		if !writeErr.HasErrorCode(11000) {
			log.Println("DBAddPappTxData err", err)
			m.Ack()
			return
		}
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

	data := wcommon.PappTxData{
		IncTx:            task.TxHash,
		IncTxData:        string(task.TxRawData),
		Type:             wcommon.PappTypeSwap,
		Status:           wcommon.StatusSubmitting,
		IsUnifiedToken:   task.IsUnifiedToken,
		FeeToken:         task.FeeToken,
		FeeAmount:        task.FeeAmount,
		BurntToken:       task.BurntToken,
		BurntAmount:      task.BurntAmount,
		Networks:         task.Networks,
		FeeRefundOTA:     task.FeeRefundOTA,
		FeeRefundOTASS:   task.FeeRefundOTASS,
		FeeRefundAddress: task.FeeRefundAddress,
		//TODO add ShardID: ,
	}
	err = database.DBSavePappTxData(data)
	if err != nil {
		writeErr, ok := err.(mongo.WriteException)
		if !ok {
			log.Println("DBAddPappTxData err", err)
			m.Ack()
			return
		}
		if !writeErr.HasErrorCode(11000) {
			log.Println("DBAddPappTxData err", err)
			m.Ack()
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

	privKey, _ := crypto.HexToECDSA(config.EVMKey)

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

		// nonce, err := getNonceByPrivateKey(evmClient, config.EVMKey)
		// if err != nil {
		// 	log.Println(err)
		// 	continue
		// }

		auth, err := bind.NewKeyedTransactorWithChainID(privKey, networkChainIdInt)
		if err != nil {
			log.Println(err)
			continue
		}

		gasPrice = gasPrice.Mul(gasPrice, big.NewInt(12))
		gasPrice = gasPrice.Div(gasPrice, big.NewInt(10))

		auth.GasPrice = gasPrice
		auth.GasLimit = wcommon.EVMGasLimit

		tx, err := evmproof.ExecuteWithBurnProof(c, auth, proof)
		if err != nil {
			log.Println(err)
			continue
		}
		result.Txhash = tx.Hash().String()
		result.Status = wcommon.StatusPending
		result.Type = wcommon.PappTypeSwap
		result.Network = network
		result.IncRequestTx = incTxHash
		break
	}

	if result.Txhash == "" {
		return nil, errors.New("submit tx outchain failed")
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

	otaReceiver := coin.OTAReceiver{}

	err = otaReceiver.FromString(task.OTA)
	if err != nil {
		log.Println("DBUpdateRefundFeeRefundTx error ", err)
		return
	}

	var sharedSecrets []inccrypto.Point
	otass, err := hex.DecodeString(task.OTASS)
	if err != nil {
		log.Println("DecodeString OTASS error ", err)
		return
	}

	err = json.Unmarshal(otass, &sharedSecrets)
	if err != nil {
		log.Println("DecodeString OTASS error ", err)
		return
	}

	otaReceiver.SharedSecrets = sharedSecrets

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

}
func speedupOutChainSwapTx(network int, evmTxHash string) error {
	//TODO
	return nil
}

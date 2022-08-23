package submitproof

import (
	"context"
	"encoding/json"
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
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	"github.com/incognitochain/go-incognito-sdk-v2/wallet"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/evmproof"
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

	swapTxTopic, err = startPubsubTopic(SWAP_TX_TOPIC)
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
	incclient.Logger = incclient.NewLogger(true)
	log.Println("Done submit keys")

	var shieldSub *pubsub.Subscription
	shieldSub, err = psclient.CreateSubscription(context.Background(), serviceID.String()+"_shield",
		pubsub.SubscriptionConfig{Topic: shieldTxTopic})
	if err != nil {
		shieldSub = psclient.Subscription(serviceID.String() + "_shield")
	}
	log.Println("shieldSub.ID()", shieldSub.ID())

	var swapSub *pubsub.Subscription
	swapSub, err = psclient.CreateSubscription(context.Background(), serviceID.String()+"_swap",
		pubsub.SubscriptionConfig{Topic: swapTxTopic})
	if err != nil {
		swapSub = psclient.Subscription(serviceID.String() + "_swap")
	}
	log.Println("swapSub.ID()", swapSub.ID())

	go func() {
		ctx := context.Background()
		err := shieldSub.Receive(ctx, ProcessShieldRequest)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		ctx := context.Background()
		err := swapSub.Receive(ctx, ProcessSwapRequest)
		if err != nil {
			panic(err)
		}
	}()
	return nil
}

func ProcessShieldRequest(ctx context.Context, m *pubsub.Message) {
	task := SubmitProofShieldTask{}
	err := json.Unmarshal(m.Data, &task)
	if err != nil {
		// rejectDelivery(delivery, payload)
		log.Println("ProcessShieldRequest error decoding message", err)
		m.Ack()
		return
	}

	if time.Since(task.Time) > time.Hour {
		errdb := database.DBUpdateShieldTxStatus(task.TxHash, task.NetworkID, wcommon.StatusSubmitFailed, "timeout")
		if errdb != nil {
			log.Println("DBUpdateShieldTxStatus error:", errdb)
			m.Nack()
			return
		}
		m.Ack()
		return
	}

	t := time.Now().Unix()
	key := keyList[t%int64(len(keyList))]
	incTx, paymentAddr, tokenID, linkedTokenID, err := submitProof(task.TxHash, task.TokenID, task.NetworkID, key)

	if tokenID != "" && linkedTokenID != "" {
		err = database.DBUpdateShieldOnChainTxInfo(task.TxHash, task.NetworkID, paymentAddr, incTx, tokenID, linkedTokenID)
		if err != nil {
			log.Println("DBUpdateShieldOnChainTxInfo error:", err)
			m.Nack()
			return
		}
	}

	if err != nil {
		if err.Error() == ProofAlreadySubmitError {
			errdb := database.DBUpdateShieldTxStatus(task.TxHash, task.NetworkID, wcommon.StatusSubmitFailed, err.Error())
			if errdb != nil {
				log.Println("DBUpdateShieldTxStatus error:", errdb)
				m.Nack()
				return
			}
			m.Ack()
			return
		}
		log.Println("submitProof error:", err) //
		return
	}

	err = database.DBUpdateShieldTxStatus(task.TxHash, task.NetworkID, wcommon.StatusPending, "")
	if err != nil {
		log.Println("DBUpdateShieldTxStatus err:", err)
		m.Nack()
		return
	}

	m.Ack()
}

func ProcessSwapRequest(ctx context.Context, m *pubsub.Message) {
	task := SubmitPappSwapTask{}
	err := json.Unmarshal(m.Data, &task)
	if err != nil {
		errdb := database.DBUpdatePappTxStatus(m.Attributes["txhash"], wcommon.StatusSubmitFailed, err.Error())
		if err != nil {
			log.Println("DBUpdatePappTxStatus err", errdb)
		}
		log.Println("ProcessShieldRequest error decoding message", err)
		m.Ack()
		return
	}

	var errSubmit error

	if task.IsPRVTx {
		errSubmit = incClient.SendRawTx(task.TxRawData)
	} else {
		errSubmit = incClient.SendRawTokenTx(task.TxRawData)
	}

	data := wcommon.PappTxData{
		IncTx:          task.TxHash,
		IncTxData:      string(task.TxRawData),
		Type:           wcommon.PappTypeSwap,
		Status:         wcommon.StatusSubmitting,
		IsUnifiedToken: task.IsUnifiedToken,
		FeeToken:       task.FeeToken,
		FeeAmount:      task.FeeAmount,
	}
	err = database.DBAddPappTxData(data)
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

	if errSubmit != nil {
		err = database.DBUpdatePappTxStatus(task.TxHash, wcommon.StatusSubmitFailed, errSubmit.Error())
		if err != nil {
			log.Println(err)
			m.Nack()
			return
		}
		m.Ack()
		return
	}

	m.Ack()
}

func createOutChainSwapTx(network int, incTxHash string, isUnifiedToken bool) (*wcommon.ExternalTxStatus, error) {
	var result wcommon.ExternalTxStatus

	networkName := wcommon.GetNetworkName(network)

	networkInfo, err := database.DBGetBridgeNetworkInfo(networkName)
	if err != nil {
		return nil, err
	}

	pappAddress, err := database.DBGetPappContractData(networkName, wcommon.PappTypeSwap)
	if err != nil {
		return nil, err
	}

	networkChainId := networkInfo.ChainID

	networkChainIdInt := new(big.Int)
	networkChainIdInt.SetString(networkChainId, 10)

	var proof *evmproof.DecodedProof
	if isUnifiedToken {
		proof, err = evmproof.GetAndDecodeBurnProofUnifiedToken(config.FullnodeURL, incTxHash, 0, uint(network))
	} else {
		proof, err = evmproof.GetAndDecodeBurnProofV2(config.FullnodeURL, incTxHash, "getburnplgprooffordeposittosc")
	}
	if err != nil {
		return nil, err
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

		auth.GasPrice = gasPrice

		tx, err := evmproof.ExecuteWithBurnProof(c, auth, proof)
		if err != nil {
			log.Println(err)
			continue
		}
		result.Txhash = tx.Hash().String()
		result.Status = wcommon.StatusPendingOutchain
		result.Type = wcommon.PappTypeSwap
		result.Network = networkName
		result.IncRequestTx = incTxHash
		break
	}

	if result.Txhash == "" {
		return nil, errors.New("submit tx outchain failed")
	}

	return &result, nil
}

func speedupOutChainSwapTx(network int, evmTxHash string) error {
	//TODO
	return nil
}

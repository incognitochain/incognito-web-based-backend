package submitproof

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	"github.com/incognitochain/go-incognito-sdk-v2/wallet"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/pkg/errors"
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
		errdb := database.DBUpdateShieldTxStatus(task.Txhash, task.NetworkID, wcommon.ShieldStatusSubmitFailed, "timeout")
		if errdb != nil {
			log.Println("DBUpdateShieldTxStatus error:", errdb)
			return
		}
		m.Ack()
		return
	}

	t := time.Now().Unix()
	key := keyList[t%int64(len(keyList))]
	incTx, paymentAddr, tokenID, linkedTokenID, err := submitProof(task.Txhash, task.TokenID, task.NetworkID, key)

	if tokenID != "" && linkedTokenID != "" {
		err = database.DBUpdateShieldOnChainTxInfo(task.Txhash, task.NetworkID, paymentAddr, incTx, tokenID, linkedTokenID)
		if err != nil {
			log.Println("DBUpdateShieldOnChainTxInfo error:", err)
			return
		}
	}

	if err != nil {
		if err.Error() == ProofAlreadySubmitError {
			errdb := database.DBUpdateShieldTxStatus(task.Txhash, task.NetworkID, wcommon.ShieldStatusSubmitFailed, err.Error())
			if errdb != nil {
				log.Println("DBUpdateShieldTxStatus error:", errdb)
				return
			}
			m.Ack()
			return
		}
		log.Println("submitProof error:", err) //
		return
	}

	err = database.DBUpdateShieldTxStatus(task.Txhash, task.NetworkID, wcommon.ShieldStatusPending, "")
	if err != nil {
		log.Println("DBUpdateShieldTxStatus err:", err)
		return
	}

	m.Ack()
}

func ProcessSwapRequest(ctx context.Context, m *pubsub.Message) {
	//TODO
}

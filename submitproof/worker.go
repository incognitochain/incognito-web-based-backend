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
		pubsub.SubscriptionConfig{Topic: shieldTxTopic})
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
		err = updateShieldTxStatus(task.Txhash, task.NetworkID, ShieldStatusSubmitFailed)
		if err != nil {
			log.Println("updateShieldTxStatus error:", err)
			return
		}
		m.Ack()
		return
	}

	t := time.Now().Unix()
	key := keyList[t%int64(len(keyList))]
	incTx, paymentAddr, tokenID, linkedTokenID, err := submitProof(task.Txhash, task.TokenID, task.NetworkID, key)
	if err != nil {
		if err.Error() == ProofAlreadySubmitError {
			m.Ack()
			return
		}

		log.Println("submitProof error:", err) //
		return
	}

	err = updateShieldTxStatus(txhash, networkID, ShieldStatusPending)
	if err != nil {
		log.Println("error123:", err)
		return "", "", "", "", err
	}

	m.Ack()
	// if incTx != "" {
	// 	watchQueue, err := rdmq.OpenQueue(MqWatchTx)
	// 	if err != nil {
	// 		log.Printf("rejected %s", payload)
	// 		return
	// 	}

	// 	task := WatchShieldProofTask{
	// 		PaymentAddress: paymentAddr,
	// 		Txhash:         task.Txhash,
	// 		IncTx:          incTx,
	// 		TokenID:        task.TokenID,
	// 		NetworkID:      task.NetworkID,
	// 		Time:           time.Now(),
	// 	}
	// 	taskBytes, _ := json.Marshal(task)

	// 	err = watchQueue.PublishBytes(taskBytes)
	// 	if err != nil {
	// 		log.Printf("PublishBytes %s", payload)
	// 		return
	// 	}
	// }

}

func ProcessSwapRequest(ctx context.Context, m *pubsub.Message) {

}

// func createConsumer(keylist []string) error {
// 	taskQueue, err := rdmq.OpenQueue(MqSubmitTx)
// 	if err != nil {
// 		return err
// 	}
// 	err = taskQueue.StartConsuming(10, time.Second)
// 	if err != nil {
// 		return err
// 	}

// 	for i := 0; i < len(keylist); i++ {
// 		name := fmt.Sprintf("consumer %d", i)
// 		if _, err := taskQueue.AddConsumer(name, NewSubmitProofConsumer(keylist[i], INC_NetworkID)); err != nil {
// 			panic(err)
// 		}
// 	}

// 	return nil
// }

// func NewSubmitProofConsumer(key string, network int) *SubmitProofConsumer {
// 	return &SubmitProofConsumer{
// 		UseKey:    key,
// 		NetworkID: network,
// 	}
// }

// func (consumer *SubmitProofConsumer) Consume(delivery rmq.Delivery) {
// 	payload := delivery.Payload()
// 	log.Println("start consume task...")
// 	task := SubmitProofShieldTask{}
// 	err := json.Unmarshal([]byte(payload), &task)
// 	if err != nil {
// 		rejectDelivery(delivery, payload)
// 	}

// 	if time.Since(task.Time) > time.Hour {
// 		err = updateShieldTxStatus(task.Txhash, task.NetworkID, ShieldStatusSubmitFailed)
// 		if err != nil {
// 			log.Println("updateShieldTxStatus error:", err)
// 			return
// 		}
// 		ackDelivery(delivery, payload)
// 		return
// 	}
// 	incTx, paymentAddr, err := submitProof(task.Txhash, task.TokenID, task.NetworkID, consumer.UseKey)
// 	if err != nil {
// 		if err.Error() == ProofAlreadySubmitError {
// 			ackDelivery(delivery, payload)
// 			return
// 		}
// 		rejectDelivery(delivery, payload)
// 		return
// 	}

// 	ackDelivery(delivery, payload)
// 	if incTx != "" {
// 		watchQueue, err := rdmq.OpenQueue(MqWatchTx)
// 		if err != nil {
// 			log.Printf("rejected %s", payload)
// 			return
// 		}

// 		task := WatchShieldProofTask{
// 			PaymentAddress: paymentAddr,
// 			Txhash:         task.Txhash,
// 			IncTx:          incTx,
// 			TokenID:        task.TokenID,
// 			NetworkID:      task.NetworkID,
// 			Time:           time.Now(),
// 		}
// 		taskBytes, _ := json.Marshal(task)

// 		err = watchQueue.PublishBytes(taskBytes)
// 		if err != nil {
// 			log.Printf("PublishBytes %s", payload)
// 			return
// 		}
// 	}
// }

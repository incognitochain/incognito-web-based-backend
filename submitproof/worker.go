package submitproof

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/adjust/rmq/v4"
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

	err := connectDB(cfg.DatabaseURLs)
	if err != nil {
		return err
	}

	err = connectMQ(serviceID, cfg.DatabaseURLs)
	if err != nil {
		return err
	}

	switch network {
	case "mainnet":
		incClient, err = incclient.NewMainNetClient()
	case "testnet-2": // testnet2
		incClient, err = incclient.NewTestNetClient()
	case "testnet-1":
		incClient, err = incclient.NewTestNet1Client()
	case "devnet":
		return errors.New("unsupported network")
	}
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
	go func() {
		err := consumeTasks(keyList)
		if err != nil {
			log.Fatalln(err)
		}
	}()
	return nil
}

func consumeTasks(keylist []string) error {
	taskQueue, err := rdmq.OpenQueue(MqSubmitTx)
	if err != nil {
		return err
	}
	err = taskQueue.StartConsuming(10, time.Second)
	if err != nil {
		return err
	}

	for i := 0; i < len(keylist); i++ {
		name := fmt.Sprintf("consumer %d", i)
		if _, err := taskQueue.AddConsumer(name, NewConsumer(keylist[i])); err != nil {
			panic(err)
		}
	}

	return nil
}
func NewConsumer(key string) *TaskConsumer {
	return &TaskConsumer{
		UseKey: key,
	}
}

func (consumer *TaskConsumer) Consume(delivery rmq.Delivery) {
	payload := delivery.Payload()
	log.Println("start consume task...")
	task := SubmitProofTask{}
	err := json.Unmarshal([]byte(payload), &task)
	if err != nil {
		if err := delivery.Reject(); err != nil {
			log.Printf("failed to reject %s: %s", payload, err)
			return
		} else {
			log.Printf("rejected %s", payload)
			return
		}
	}
	incTx, err := submitProof(task.Txhash, task.TokenID, task.NetworkID, consumer.UseKey)
	if err != nil {
		if err := delivery.Reject(); err != nil {
			log.Printf("failed to reject %s: %s", payload, err)
			return
		} else {
			log.Printf("rejected %s", payload)
			return
		}
	}

	if err := delivery.Ack(); err != nil {
		log.Printf("failed to ack %s: %s", payload, err)
	} else {
		log.Printf("acked %s", payload)
	}
	if incTx != "" {
		watchQueue, err := rdmq.OpenQueue(MqWatchTx)
		if err != nil {
			log.Printf("rejected %s", payload)
			return
		}

		task := WatchProofTask{
			Txhash: task.Txhash,
			IncTx:  incTx,
		}
		taskBytes, _ := json.Marshal(task)

		err = watchQueue.PublishBytes(taskBytes)
		if err != nil {
			log.Printf("PublishBytes %s", payload)
			return
		}
	}

	// debugf("start consume %s", payload)
	// time.Sleep(consumeDuration)

	// consumer.count++
	// if consumer.count%reportBatchSize == 0 {
	// 	duration := time.Now().Sub(consumer.before)
	// 	consumer.before = time.Now()
	// 	perSecond := time.Second / (duration / reportBatchSize)
	// 	log.Printf("%s consumed %d %s %d", consumer.name, consumer.count, payload, perSecond)
	// }

	// if consumer.count%reportBatchSize > 0 {
	// 	if err := delivery.Ack(); err != nil {
	// 		debugf("failed to ack %s: %s", payload, err)
	// 	} else {
	// 		debugf("acked %s", payload)
	// 	}
	// } else { // reject one per batch
	// 	if err := delivery.Reject(); err != nil {
	// 		debugf("failed to reject %s: %s", payload, err)
	// 	} else {
	// 		debugf("rejected %s", payload)
	// 	}
	// }
}

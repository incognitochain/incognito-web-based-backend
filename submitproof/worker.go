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
	go func() {
		err := createConsumer(keyList)
		if err != nil {
			log.Fatalln(err)
		}
	}()
	return nil
}

func createConsumer(keylist []string) error {
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
		if _, err := taskQueue.AddConsumer(name, NewSubmitProofConsumer(keylist[i], INC_NetworkID)); err != nil {
			panic(err)
		}
	}

	return nil
}
func NewSubmitProofConsumer(key string, network int) *SubmitProofConsumer {
	return &SubmitProofConsumer{
		UseKey:    key,
		NetworkID: network,
	}
}

func (consumer *SubmitProofConsumer) Consume(delivery rmq.Delivery) {
	payload := delivery.Payload()
	log.Println("start consume task...")
	task := SubmitProofShieldTask{}
	err := json.Unmarshal([]byte(payload), &task)
	if err != nil {
		rejectDelivery(delivery, payload)
	}

	incTx, err := submitProof(task.Txhash, task.TokenID, task.NetworkID, consumer.UseKey)
	if err != nil {
		rejectDelivery(delivery, payload)
	}

	ackDelivery(delivery, payload)
	if incTx != "" {
		watchQueue, err := rdmq.OpenQueue(MqWatchTx)
		if err != nil {
			log.Printf("rejected %s", payload)
			return
		}

		task := WatchProofTask{
			Txhash:    task.Txhash,
			IncTx:     incTx,
			TokenID:   task.TokenID,
			NetworkID: task.NetworkID,
			Time:      time.Now(),
		}
		taskBytes, _ := json.Marshal(task)

		err = watchQueue.PublishBytes(taskBytes)
		if err != nil {
			log.Printf("PublishBytes %s", payload)
			return
		}
	}
}

package submitproof

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/adjust/rmq/v4"
	"github.com/google/uuid"
	"github.com/incognitochain/incognito-web-based-backend/common"
)

func StartWatcher(cfg common.Config, serviceID uuid.UUID) error {
	config = cfg
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

	taskQueue, err := rdmq.OpenQueue(MqWatchTx)
	if err != nil {
		return err
	}
	err = taskQueue.StartConsuming(10, time.Second)
	if err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		_, err = taskQueue.AddConsumerFunc(fmt.Sprintf("submitwatcher-%v", i), watchSubmittedTx)
		if err != nil {
			return err
		}
	}

	go watchUnackTask()
	go retryFailedTask()
	return nil
}

func watchUnackTask() {
	cleaner := rmq.NewCleaner(rdmq)
	for range time.Tick(60 * time.Second) {
		returned, err := cleaner.Clean()
		if err != nil {
			log.Printf("failed to clean: %s", err)
			continue
		}
		log.Printf("cleaned %d", returned)
	}
}

func retryFailedTask() {
	for range time.Tick(30 * time.Second) {
		queue, err := rdmq.OpenQueue(MqSubmitTx)
		if err != nil {
			panic(err)
		}
		returned, err := queue.ReturnRejected(math.MaxInt64)
		if err != nil {
			panic(err)
		}

		log.Printf("queue returner returned %d rejected deliveries", returned)
	}
}

func watchSubmittedTx(delivery rmq.Delivery) {
	payload := delivery.Payload()
	log.Println("start consume task...")
	task := WatchProofTask{}
	err := json.Unmarshal([]byte(payload), &task)
	if err != nil {
		rejectDelivery(delivery, payload)
	}

	incClient.GetRawMemPool()
	isInBlock, err := incClient.CheckTxInBlock(task.IncTx)
	if err != nil {
		rejectDelivery(delivery, payload)
	}

	if isInBlock {
		status, err := incClient.CheckShieldStatus(task.IncTx)
		if err != nil {
			log.Printf("CheckShieldStatus err", err)
			rejectDelivery(delivery, payload)
		}
		switch status {
		case 1:
			err = updateShieldTxStatus(task.Txhash, task.NetworkID, task.TokenID, ShieldStatusPending)
			if err != nil {
				log.Println("error123:", err)
				rejectDelivery(delivery, payload)
			}
		case 2:
			err = updateShieldTxStatus(task.Txhash, task.NetworkID, task.TokenID, ShieldStatusAccepted)
			if err != nil {
				log.Println("error123:", err)
				rejectDelivery(delivery, payload)
			}
			ackDelivery(delivery, payload)
			return
		case 3:
			err = updateShieldTxStatus(task.Txhash, task.NetworkID, task.TokenID, ShieldStatusRejected)
			if err != nil {
				log.Println("error123:", err)
				rejectDelivery(delivery, payload)
			}
		}
		delivery.Push()
	} else {
		rejectDelivery(delivery, payload)
	}
}

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

	err := connectDB(cfg.DatabaseURLs, cfg.DBUSER, cfg.DBPASS)
	if err != nil {
		return err
	}

	err = connectMQ(serviceID, cfg.DatabaseURLs, cfg.DBUSER, cfg.DBPASS)
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
		log.Printf("queue MqSubmitTx returner returned %d rejected deliveries", returned)

		queue, err = rdmq.OpenQueue(MqWatchTx)
		if err != nil {
			panic(err)
		}
		returned, err = queue.ReturnRejected(math.MaxInt64)
		if err != nil {
			panic(err)
		}

		log.Printf("queue MqWatchTx returner returned %d rejected deliveries", returned)
	}
}

func watchSubmittedTx(delivery rmq.Delivery) {
	payload := delivery.Payload()
	log.Println("start consume task...")
	task := WatchShieldProofTask{}
	err := json.Unmarshal([]byte(payload), &task)
	if err != nil {
		rejectDelivery(delivery, payload)
		return
	}

	incClient.GetRawMemPool()
	isInBlock, err := incClient.CheckTxInBlock(task.IncTx)
	if err != nil {
		rejectDelivery(delivery, payload)
	}

	if isInBlock {
		var status int
		if task.IsPunified {
			statusShield, err := incClient.CheckUnifiedShieldStatus(task.IncTx)
			if err != nil {
				log.Println("CheckShieldStatus err", err)
				rejectDelivery(delivery, payload)
				return
			}
			if statusShield.Status == 0 {
				status = 3
			} else {
				status = 2
			}
		} else {
			status, err = incClient.CheckShieldStatus(task.IncTx)
			if err != nil {
				log.Println("CheckShieldStatus err", err)
				rejectDelivery(delivery, payload)
			}
		}

		switch status {
		case 1:
			err = updateShieldTxStatus(task.Txhash, task.NetworkID, ShieldStatusPending)
			if err != nil {
				log.Println("error123:", err)
				rejectDelivery(delivery, payload)
			}
		case 2:
			err = updateShieldTxStatus(task.Txhash, task.NetworkID, ShieldStatusAccepted)
			if err != nil {
				log.Println("error123:", err)
				rejectDelivery(delivery, payload)
			}
			ackDelivery(delivery, payload)
			faucetPRV(task.PaymentAddress)
			return
		case 3:
			err = updateShieldTxStatus(task.Txhash, task.NetworkID, ShieldStatusRejected)
			if err != nil {
				log.Println("error123:", err)
				rejectDelivery(delivery, payload)
			}
			ackDelivery(delivery, payload)
			return
		}
		delivery.Push()
	} else {
		rejectDelivery(delivery, payload)
	}
}

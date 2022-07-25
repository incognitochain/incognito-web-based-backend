package submitproof

import (
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

	err := connectDB(cfg.DatabaseURLs)
	if err != nil {
		return err
	}

	err = connectMQ(serviceID, cfg.DatabaseURLs)
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

	_, err = taskQueue.AddConsumerFunc("submitwatcher", func(delivery rmq.Delivery) {
		fmt.Println(delivery.Payload())
	})
	if err != nil {
		return err
	}
	go watchUnackTask()
	go retryFailedTask()
	return nil
}

func watchUnackTask() {
	cleaner := rmq.NewCleaner(rdmq)

	for range time.Tick(15 * time.Second) {
		returned, err := cleaner.Clean()
		if err != nil {
			log.Printf("failed to clean: %s", err)
			continue
		}
		log.Printf("cleaned %d", returned)
	}
}

func retryFailedTask() {
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

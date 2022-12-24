package main

import (
	"fmt"
	"log"
	"syscall"

	"github.com/google/uuid"
	"github.com/incognitochain/incognito-web-based-backend/api"
	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/feeestimator"
	"github.com/incognitochain/incognito-web-based-backend/interswap"
	"github.com/incognitochain/incognito-web-based-backend/slacknoti"
	"github.com/incognitochain/incognito-web-based-backend/submitproof"
)

var serviceID uuid.UUID

func init() {
	id := uuid.New()
	log.SetPrefix(fmt.Sprintf("pid:%d ", syscall.Getpid()))
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	serviceID = id
}

func main() {
	err := loadConfig()
	if err != nil {
		panic(err)
	}

	if config.SlackMonitor != "" {
		go slacknoti.StartSlackHook()
	}

	err = database.ConnectDB(config.Mongodb, config.Mongo, config.NetworkID)
	if err != nil {
		panic(err)
	}

	keylist, err := loadKeyList()
	if err != nil {
		log.Println(err)
	}
	switch config.Mode {
	case common.MODE_FEEESTIMATOR:
		if err := feeestimator.StartService(config); err != nil {
			log.Fatalln(err)
		}
	case common.MODE_TXSUBMITWATCHER:
		if err := submitproof.StartWatcher(keylist, config, serviceID); err != nil {
			log.Fatalln(err)
		}
	case common.MODE_TXSUBMITWORKER:
		go func() {
			if err := submitproof.StartWorker(keylist, config, serviceID); err != nil {
				log.Fatalln(err)
			}
		}()
	case common.MODE_INTERSWAP:
		go func() {
			// InterSwap start service
			if err := interswap.StartWorker(config, serviceID); err != nil {
				log.Fatalln(err)
			}
		}()
	case common.MODE_API:
		go func() {
			if err := submitproof.StartAssigner(config, serviceID); err != nil {
				log.Fatalln(err)
			}
		}()
		go api.StartAPIservice(config)
	}
	select {}
}

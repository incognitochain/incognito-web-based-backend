package main

import (
	"log"

	"github.com/google/uuid"
	"github.com/incognitochain/incognito-web-based-backend/api"
	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/submitproof"
)

var serviceID uuid.UUID

func init() {
	id := uuid.New()
	serviceID = id
}

func main() {
	err := loadConfig()
	if err != nil {
		panic(err)
	}
	switch config.Mode {
	case common.MODE_TXSUBMITWATCHER:
		if err := submitproof.StartWatcher(config, serviceID); err != nil {
			log.Fatalln(err)
		}
	case common.MODE_TXSUBMITWORKER:
		keylist, err := loadKeyList()
		if err != nil {
			log.Println(err)
		}
		go func() {
			if err := submitproof.StartWorker(keylist, config, serviceID); err != nil {
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

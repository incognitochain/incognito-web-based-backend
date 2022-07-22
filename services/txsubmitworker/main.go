package main

import (
	"log"

	"github.com/google/uuid"
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
	keylist, err := loadKeyList()
	if err != nil {
		log.Println(err)
	}
	err = submitproof.StartWorker(keylist, config, serviceID)
	if err != nil {
		panic(err)
	}
	select {}
}

package main

import (
	"log"

	"github.com/incognitochain/incognito-web-based-backend/api"
	"github.com/incognitochain/incognito-web-based-backend/submitproof"
)

func main() {
	err := loadConfig()
	if err != nil {
		panic(err)
	}
	keylist, err := loadKeyList()
	if err != nil {
		log.Println(err)
	}
	err = submitproof.Start(keylist, config)
	if err != nil {
		panic(err)
	}
	go api.StartAPIservice(config)
	select {}
}

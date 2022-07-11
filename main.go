package main

import (
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
		panic(err)
	}
	err = submitproof.Start(keylist, "testnet")
	if err != nil {
		panic(err)
	}
	go api.StartAPIservice(config)
	select {}
}

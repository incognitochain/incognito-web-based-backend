package main

import "github.com/incognitochain/incognito-web-based-backend/api"

func main() {
	err := loadConfig()
	if err != nil {
		panic(err)
	}
	go api.StartAPIservice(config)
	select {}
}

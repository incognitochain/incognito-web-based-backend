package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/incognitochain/incognito-web-based-backend/common"
)

var config common.Config

func loadConfig() error {
	data, err := ioutil.ReadFile("./cfg.json")
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			log.Println("cfg.json isn't exist")
			config = common.DefaultConfig
			return nil
		}
		log.Fatalln(err)
	}

	var tempCfg common.Config
	if data != nil {
		err = json.Unmarshal(data, &tempCfg)
		if err != nil {
			return err
		}
		config = tempCfg
	} else {
		log.Println("cfg.json is empty")
		config = common.DefaultConfig
	}
	return nil
}

func loadKey() error {
	evmKeyEncrypted := os.Getenv("EVMKEY_ENC")
	incKeyEncrypted := os.Getenv("INCKEY_ENC")
	decryptKeyEncrypted := os.Getenv("DECKEY")

	if decryptKeyEncrypted == "" {
		return nil
	} else {
		log.Println("using env encrypted key")
	}

	if evmKeyEncrypted != "" {

	}

	if incKeyEncrypted != "" {

	}
	return nil
}

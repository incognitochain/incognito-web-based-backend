package database

import (
	"context"
	"log"
	"strings"

	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/kamva/mgm/v3"
)

func DBCreateDefaultNetworkInfo(network string) {

	isTestnet := strings.Contains(network, "testnet")

	if !checkBridgeNetworkDataExist() {
		createDefaultBridgeData(isTestnet)
	}
	if !checkPappsEndpointExist() {
		createDefaultPappsEndpoint(isTestnet)
	}

	if !checkPappVaultDataExist() {
		createDefaultVaultData(isTestnet)
	}
}

func checkBridgeNetworkDataExist() bool {
	len, err := mgm.Coll(&common.BridgeNetworkData{}).EstimatedDocumentCount(context.Background())
	if err != nil {
		log.Println(err)
		return false
	}
	if len == 0 {
		return false
	}
	return true
}

func checkPappVaultDataExist() bool {
	len, err := mgm.Coll(&common.PappVaultData{}).EstimatedDocumentCount(context.Background())
	if err != nil {
		log.Println(err)
		return false
	}
	if len == 0 {
		return false
	}
	return true
}

func checkPappsEndpointExist() bool {
	len, err := mgm.Coll(&common.PAppsEndpointData{}).EstimatedDocumentCount(context.Background())
	if err != nil {
		log.Println(err)
		return false
	}
	if len == 0 {
		return false
	}
	return true
}

func createDefaultBridgeData(isTestnet bool) {
	docs := []interface{}{}
	if isTestnet {
		for _, data := range common.TestnetBridgeNetworkData {
			docs = append(docs, data)
		}
	} else {
		for _, data := range common.MainnetBridgeNetworkData {
			docs = append(docs, data)
		}
	}

	_, err := mgm.Coll(&common.BridgeNetworkData{}).InsertMany(context.Background(), docs)
	if err != nil {
		log.Println(err)
	}
}

func createDefaultVaultData(isTestnet bool) {

	docs := []interface{}{}
	if isTestnet {
		for _, data := range common.TestnetIncognitoVault {
			docs = append(docs, data)
		}
	} else {
		for _, data := range common.MainnetIncognitoVault {
			docs = append(docs, data)
		}
	}

	_, err := mgm.Coll(&common.PappVaultData{}).InsertMany(context.Background(), docs)
	if err != nil {
		log.Println(err)
	}
}

func createDefaultPappsEndpoint(isTestnet bool) {
	docs := []interface{}{}
	if isTestnet {
		for _, data := range common.TestnetPappsEndpointData {
			docs = append(docs, data)
		}
	} else {
		for _, data := range common.MainnetPappsEndpointData {
			docs = append(docs, data)
		}
	}

	_, err := mgm.Coll(&common.PAppsEndpointData{}).InsertMany(context.Background(), docs)
	if err != nil {
		log.Println(err)
	}
}

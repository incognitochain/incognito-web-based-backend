package feeestimator

import (
	"errors"
	"time"

	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
)

const (
	checkFeeInterval = 15 * time.Second
)

// 						 network	app		 endpoint
var pAppsList = make(map[string]map[string][]string)

func StartService(cfg common.Config, papps map[string]map[string][]string) error {
	return checkFee()
}

func checkFee() error {
	for range time.Tick(checkFeeInterval) {
		feeList := make(map[string]interface{})
		for network, _ := range pAppsList {
			getFee(network)
		}

		saveFeeData(feeList)
	}
	return nil
}

func getFee(network string) error {

	switch network {
	case common.NETWORK_ETH:
	case common.NETWORK_BSC:
	case common.NETWORK_PLG:
	case common.NETWORK_FTM:
	default:
		return errors.New("unsupported network")
	}

	return nil
}

func saveFeeData(data map[string]interface{}) error {
	var feeData common.ExternalNetworksFeeData
	feeData.Creating()
	feeData.Fees = data
	return database.DBSaveFeetTable(feeData)
}

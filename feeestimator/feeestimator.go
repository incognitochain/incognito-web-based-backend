package feeestimator

import (
	"time"

	"github.com/incognitochain/incognito-web-based-backend/common"
)

const (
	checkFeeInterval = 15 * time.Second
)

func StartService(network string, cfg common.Config) error {

	return checkFee()
}

func checkFee() error {
	for range time.Tick(checkFeeInterval) {

	}
	return nil
}

func getFee(network string) error {
	return nil
}

func saveFeeTable()

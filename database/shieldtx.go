package database

import "github.com/incognitochain/incognito-web-based-backend/common"

func DBCreateShieldTxData(externalTx string, networkID int) error {
	return nil
}

func DBRetrievePendingShieldTxs() ([]common.ShieldTxData, error) {
	var result []common.ShieldTxData
	return result, nil
}

func DBRetrieveFailedShieldTxs() ([]common.ShieldTxData, error) {
	var result []common.ShieldTxData
	return result, nil
}

func DBGetShieldTxStatusByExternalTx(externalTx string, networkID int) (*common.ShieldTxData, error) {
	var result common.ShieldTxData
	return &result, nil
}

func DBUpdateShieldTxStatus(externalTx string, networkID int, status string, err string) error {
	return nil
}

func DBUpdateShieldOnChainTxInfo(externalTx string, networkID int, incTx string, tokenID string, linkedTokenID string) error {
	return nil
}

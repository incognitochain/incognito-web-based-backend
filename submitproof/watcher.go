package submitproof

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/pkg/errors"
)

func StartWatcher(cfg common.Config, serviceID uuid.UUID) error {
	config = cfg
	// network := cfg.NetworkID
	watchPendingTx()

	return nil
}

func watchPendingTx() {
	for {
		txList, err := database.DBRetrievePendingShieldTxs(0, 0)
		if err != nil {
			log.Println("DBRetrievePendingShieldTxs err:", err)
		}
		for _, txdata := range txList {
			err := processPendingShieldTxs(txdata)
			if err != nil {
				log.Println("processPendingShieldTxs err:", txdata.IncTx)
			}
		}
		time.Sleep(10 * time.Second)
	}
}

func processPendingShieldTxs(txdata common.ShieldTxData) error {
	isInBlock, err := incClient.CheckTxInBlock(txdata.IncTx)
	if err != nil {
		log.Println("CheckTxInBlock err:", err)
		return err
	}

	if isInBlock {
		var status int
		if txdata.TokenID == txdata.UTokenID {
			statusShield, err := incClient.CheckUnifiedShieldStatus(txdata.IncTx)
			if err != nil {
				log.Println("CheckShieldStatus err", err)
				return err
			}
			if statusShield.Status == 0 {
				status = 3
			} else {
				status = 2
			}
		} else {
			status, err = incClient.CheckShieldStatus(txdata.IncTx)
			if err != nil {
				log.Println("CheckShieldStatus err", err)
				return err
			}
		}

		switch status {
		case 1:
			err = database.DBUpdateShieldTxStatus(txdata.ExternalTx, txdata.NetworkID, common.ShieldStatusPending, "")
			if err != nil {
				log.Println("DBUpdateShieldTxStatus err:", err)
				return err
			}
		case 2:
			err = database.DBUpdateShieldTxStatus(txdata.ExternalTx, txdata.NetworkID, common.ShieldStatusAccepted, "")
			if err != nil {
				log.Println("DBUpdateShieldTxStatus err:", err)
				return err
			}
			faucetPRV(txdata.PaymentAddress)
			return nil
		case 3:
			err = database.DBUpdateShieldTxStatus(txdata.ExternalTx, txdata.NetworkID, common.ShieldStatusRejected, "rejected by chain")
			if err != nil {
				log.Println("DBUpdateShieldTxStatus err:", err)
				return err
			}
			return nil
		}
	}
	return errors.New("tx not finalized")
}

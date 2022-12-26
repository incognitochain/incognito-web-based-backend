package submitproof

import (
	"fmt"
	"log"
	"strings"
	"time"

	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/papps/popensea"
	"github.com/incognitochain/incognito-web-based-backend/slacknoti"
)

func processPendingOpenseaTx(tx wcommon.PappTxData) error {
	txDetail, err := incClient.GetTxDetail(tx.IncTx)
	if err != nil {
		if strings.Contains(err.Error(), "RPC returns an error:") {
			err = database.DBUpdatePappTxStatus(tx.IncTx, wcommon.StatusSubmitFailed, err.Error())
			if err != nil {
				log.Println("DBUpdateShieldTxStatus err:", err)
				return err
			}
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[opensea]` submit opensea failed 😵 `%v` \n", tx.IncTx))
			return nil
		}
		return err
	}
	if txDetail.IsInBlock {
		status, err := checkBeaconBridgeAggUnshieldStatus(tx.IncTx)
		if err != nil {
			return err
		}

		switch status {
		case 0:
			err = database.DBUpdatePappTxStatus(tx.IncTx, wcommon.StatusRejected, "")
			if err != nil {
				return err
			}
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[opensea]` inctx `%v` rejected by beacon 😢\n", tx.IncTx))
		case 1:
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[opensea]` inctx `%v` accepted by beacon 👌\n", tx.IncTx))
			err = database.DBUpdatePappTxStatus(tx.IncTx, wcommon.StatusAccepted, "")
			if err != nil {
				return err
			}
			err = database.DBUpdatePappTxSubmitOutchainStatus(tx.IncTx, wcommon.StatusWaiting)
			if err != nil {
				return err
			}
			for _, network := range tx.Networks {
				_, err := SubmitOutChainTx(tx.IncTx, network, tx.IsUnifiedToken, false, wcommon.ExternalTxTypeOpensea)
				if err != nil {
					return err
				}
			}
		default:
			if tx.Status != wcommon.StatusExecuting && tx.Status != wcommon.StatusAccepted {
				err = database.DBUpdatePappTxStatus(tx.IncTx, wcommon.StatusExecuting, "")
				if err != nil {
					return err
				}
			}
		}

	}
	return nil
}

func updateOpenSeaCollections() {
	for {
		time.Sleep(8 * time.Second)
		result, err := popensea.RetrieveCollectionList(config.OpenSeaAPI, config.OpenSeaAPIKey, 20, 0)
		if err != nil {
			log.Println(err)
			continue
		}
		err = database.DBSaveCollectionsInfo(result)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func watchCollections() {
	for {
		time.Sleep(8 * time.Second)
	}
}

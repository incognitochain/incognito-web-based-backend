package submitproof

import (
	"context"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/pkg/errors"
)

func StartWatcher(cfg wcommon.Config, serviceID uuid.UUID) error {
	config = cfg
	// network := cfg.NetworkID
	go watchPendingIncTx()
	go watchPendingExternalTx()

	return nil
}

func watchPendingExternalTx() {
	for {
		networks, err := database.DBGetBridgeNetworkInfos()
		if err != nil {
			log.Println("DBGetBridgeNetworkInfos err:", err)
		}
		for _, networkInfo := range networks {
			currentEVMHeight, err := getEVMBlockHeight(networkInfo.Endpoints)
			if err != nil {
				log.Fatalln("getEVMBlockHeight err:", err)
			}
			txList, err := database.DBRetrievePendingExternalTx(networkInfo.Network, 0, 0)
			if err != nil {
				log.Println("DBRetrievePendingExternalTx err:", err)
			}
			for _, tx := range txList {
				err := processPendingExternalTxs(tx, currentEVMHeight, 15, networkInfo.Endpoints)
				if err != nil {
					log.Println("processPendingExternalTxs err:", err)
				}
			}
		}

		time.Sleep(10 * time.Second)
	}
}

func watchPendingIncTx() {
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

func processPendingShieldTxs(txdata wcommon.ShieldTxData) error {
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
			err = database.DBUpdateShieldTxStatus(txdata.ExternalTx, txdata.NetworkID, wcommon.StatusPending, "")
			if err != nil {
				log.Println("DBUpdateShieldTxStatus err:", err)
				return err
			}
		case 2:
			err = database.DBUpdateShieldTxStatus(txdata.ExternalTx, txdata.NetworkID, wcommon.StatusAccepted, "")
			if err != nil {
				log.Println("DBUpdateShieldTxStatus err:", err)
				return err
			}
			faucetPRV(txdata.PaymentAddress)
			return nil
		case 3:
			err = database.DBUpdateShieldTxStatus(txdata.ExternalTx, txdata.NetworkID, wcommon.StatusRejected, "rejected by chain")
			if err != nil {
				log.Println("DBUpdateShieldTxStatus err:", err)
				return err
			}
			return nil
		}
	}
	return errors.New("tx not finalized")
}

func processPendingExternalTxs(tx wcommon.ExternalTxStatus, currentEVMHeight uint64, finalizeRange uint64, endpoints []string) error {
	for _, endpoint := range endpoints {
		evmClient, _ := ethclient.Dial(endpoint)
		txHash := common.Hash{}
		err := txHash.UnmarshalText([]byte(tx.Txhash))
		if err != nil {
			return err
		}
		txReceipt, err := evmClient.TransactionReceipt(context.Background(), txHash)
		if err != nil {
			return err
		}
		if currentEVMHeight >= txReceipt.BlockNumber.Uint64()+finalizeRange {
			// todo update status
		}
		return nil
	}
	return errors.New("no endpoints reachable")
}

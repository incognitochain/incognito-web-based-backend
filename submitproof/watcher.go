package submitproof

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"github.com/incognitochain/bridge-eth/bridge/vault"
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
				err := processPendingExternalTxs(tx, currentEVMHeight, uint64(networkInfo.ConfirmationBlocks), networkInfo.Endpoints)
				if err != nil {
					log.Println("processPendingExternalTxs err:", err)
				}
			}
		}

		time.Sleep(10 * time.Second)
	}
}

func watchPendingIncTx() {
	go func() {
		for {
			txList, err := database.DBRetrievePendingPappTxs(wcommon.PappTypeSwap, 0, 0)
			if err != nil {
				log.Println("DBRetrievePendingShieldTxs err:", err)
			}
			for _, txdata := range txList {
				err := processPendingSwapTx(txdata)
				if err != nil {
					log.Println("processPendingShieldTxs err:", txdata.IncTx)
				}
			}
			time.Sleep(10 * time.Second)
		}
	}()

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
	networkID := wcommon.GetNetworkID(tx.Network)
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
			// todo update status to db
			valueBuf := encodeBufferPool.Get().(*bytes.Buffer)
			defer encodeBufferPool.Put(valueBuf)

			vaultABI, err := abi.JSON(strings.NewReader(vault.VaultABI))
			if err != nil {
				fmt.Println("abi.JSON", err.Error())
				return err
			}
			isRedeposit := false
			for _, d := range txReceipt.Logs {
				switch len(d.Data) {
				case 256, 288:
					unpackResult, err := vaultABI.Unpack("Redeposit", d.Data)
					if err != nil {
						log.Println("unpackResult err", err)
						continue
					}
					if len(unpackResult) < 3 {
						err = errors.New(fmt.Sprintf("Unpack event not match data needed %v\n", unpackResult))
						log.Println("len(unpackResult) err", err)
						continue
					}
					isRedeposit = true
				default:
				}
			}
			if isRedeposit {
				result, err := SubmitShieldProof(tx.Txhash, networkID, "")
				if err != nil {
					return err
				}
				fmt.Println("result", result)
			}
		}
		return nil
	}
	return errors.New("no endpoints reachable")
}

func processPendingSwapTx(tx wcommon.PappTxData) error {

	submitProofExternalChain()
	return nil
}

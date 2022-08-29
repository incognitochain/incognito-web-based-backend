package submitproof

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"github.com/incognitochain/bridge-eth/bridge/vault"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/pkg/errors"
)

func StartWatcher(cfg wcommon.Config, serviceID uuid.UUID) error {
	config = cfg
	network := cfg.NetworkID

	err := initIncClient(network)
	if err != nil {
		return err
	}

	err = StartAssigner(config, serviceID)
	if err != nil {
		return err
	}

	go watchPendingShieldTx()
	go watchPendingPappTx()
	go watchPendingExternalTx()
	go watchEVMAccountBalance()

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

func watchPendingPappTx() {
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
}

func watchPendingShieldTx() {
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
		var logResult string
		if currentEVMHeight >= txReceipt.BlockNumber.Uint64()+finalizeRange {

			err = database.DBUpdateExternalTxStatus(tx.Txhash, wcommon.StatusAccepted, "")
			if err != nil {
				return err
			}

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
					unpackResult, err := vaultABI.Unpack("ExecuteFnLog", d.Data) // same as Redeposit
					if err != nil {
						log.Println("unpackResult2 err", err)
						continue
					} else {
						logResult = fmt.Sprintf("%s", unpackResult)
					}
				}
			}
			otherInfo := wcommon.ExternalTxSwapResult{
				LogResult:   logResult,
				IsRedeposit: isRedeposit,
			}

			otherInfoBytes, _ := json.MarshalIndent(otherInfo, "", "\t")

			err = database.DBUpdateExternalTxOtherInfo(tx.Txhash, string(otherInfoBytes))
			if err != nil {
				return err
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
	txDetail, err := incClient.GetTxDetail(tx.IncTx)
	if err != nil {
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
		case 1:
			for _, network := range tx.Networks {
				_, err := SendOutChainPappTx(tx.IncTx, network, tx.IsUnifiedToken)
				if err != nil {
					return err
				}
			}
			err = database.DBUpdatePappTxStatus(tx.IncTx, wcommon.StatusAccepted, "")
			if err != nil {
				return err
			}
		default:
			if tx.Status != wcommon.StatusExecuting {
				err = database.DBUpdatePappTxStatus(tx.IncTx, wcommon.StatusExecuting, "")
				if err != nil {
					return err
				}
			}
		}

	}
	return nil
}

func watchEVMAccountBalance() {
	for {
		networks, err := database.DBGetBridgeNetworkInfos()
		if err != nil {
			log.Println("DBGetBridgeNetworkInfos err:", err)
		}
		privKey, _ := crypto.HexToECDSA(config.EVMKey)
		keyAddr := crypto.PubkeyToAddress(privKey.PublicKey)
		for _, networkInfo := range networks {
			for _, endpoint := range networkInfo.Endpoints {
				evmClient, err := ethclient.Dial(endpoint)
				if err != nil {
					log.Println(err)
					continue
				}
				gasLeft, err := evmClient.BalanceAt(context.Background(), keyAddr, nil)
				if err != nil {
					log.Println(err)
					continue
				}
				log.Printf("network %v has %v gas left\n", networkInfo.Network, gasLeft.Uint64())
				break
			}
		}
		time.Sleep(1 * time.Minute)
	}

}

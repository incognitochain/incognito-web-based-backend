package submitproof

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/google/uuid"
	"github.com/incognitochain/bridge-eth/bridge/vault"
	inccommon "github.com/incognitochain/go-incognito-sdk-v2/common"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/slacknoti"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

func StartWatcher(keylist []string, cfg wcommon.Config, serviceID uuid.UUID) error {
	config = cfg
	keyList = keylist
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
	go watchPendingUnshieldTx()
	go watchPendingPappTx()
	go watchPendingExternalTx()
	go watchIncAccountBalance()
	go watchEVMAccountBalance()
	go watchRedepositExternalTx()
	go watchUnshieldTxNeedFeeRefund()
	go watchPappTxNeedFeeRefund()
	go watchPendingFeeRefundTx()
	go forwardCollectedFee()
	go watchVaultState()
	go trackDexSwap()

	return nil
}

func trackDexSwap() {
	for {
		txList, err := database.DBRetrievePendingDexTxs(0, 0)
		if err != nil {
			log.Println("DBRetrievePendingDexTxs error", err)
			continue
		}
		markDelete := []string{}
		for _, tx := range txList {
			uaStr := parseUserAgent(tx.UserAgent)
			switch tx.Status {
			case wcommon.StatusPending:
				txDetail, err := incClient.GetTxDetail(tx.IncTx)
				if err != nil {
					log.Println("CheckTxInBlock err:", err)
				}
				if txDetail != nil {
					if txDetail.IsInBlock {
						err = database.DBUpdateDexSwapTxStatus(tx.IncTx, wcommon.StatusExecuting)
						if err != nil {
							log.Println("DBUpdateDexSwapTxStatus error", err)
							continue
						}
						slackep := os.Getenv("SLACK_PDEX_ALERT")
						if slackep != "" {
							tkInInfo, _ := getTokenInfo(tx.TokenSell)
							tkOutInfo, _ := getTokenInfo(tx.TokenBuy)

							tokenInSymbol := tkInInfo.Symbol
							tokenOutSymbol := tkOutInfo.Symbol

							swapAlert := fmt.Sprintf("`[%v | %v]` swap submitting 🛰\n SwapID: `%v`\n Requested: `%v %v` to `%v %v`\n--------------------------------------------------------", "pdex", uaStr, tx.ID.Hex(), tx.AmountIn, tokenInSymbol, tx.MinAmountOut, tokenOutSymbol)
							slacknoti.SendWithCustomChannel(swapAlert, slackep)
						}
					}
				} else {
					markDelete = append(markDelete, tx.IncTx)
				}
			case wcommon.StatusExecuting:
				isCompleted, outAmount, err := getPdexSwapTxStatus(tx.IncTx, tx.TokenBuy)
				if err != nil {
					log.Println("getPdexSwapTxStatus error", err)
					continue
				}
				if isCompleted {
					err = database.DBUpdateDexSwapTxStatus(tx.IncTx, wcommon.StatusAccepted)
					if err != nil {
						log.Println("DBUpdateDexSwapTxStatus error", err)
						continue
					}

					slackep := os.Getenv("SLACK_PDEX_ALERT")
					if slackep != "" {
						tkInInfo, _ := getTokenInfo(tx.TokenSell)
						tkOutInfo, _ := getTokenInfo(tx.TokenBuy)
						tokenInSymbol := tkInInfo.Symbol
						tokenOutSymbol := tkOutInfo.Symbol
						var swapAlert string
						if outAmount != "" {
							swapAlert = fmt.Sprintf("`[%v | %v]` swap was success 🎉\n SwapID: `%v`\n Requested: `%v %v` to `%v %v` | received: `%v %v`\n--------------------------------------------------------", "pdex", uaStr, tx.ID.Hex(), tx.AmountIn, tokenInSymbol, tx.MinAmountOut, tokenOutSymbol, outAmount, tokenOutSymbol)
						} else {
							swapAlert = fmt.Sprintf("`[%v | %v]` swap was reverted 😢\n SwapID: `%v`\n Requested: `%v %v` to `%v %v`\n--------------------------------------------------------", "pdex", uaStr, tx.ID.Hex(), tx.AmountIn, tokenInSymbol, tx.MinAmountOut, tokenOutSymbol)
						}
						slacknoti.SendWithCustomChannel(swapAlert, slackep)
					}
				}
			}
		}

		err = database.DBDeleteDexSwap(markDelete)
		if err != nil {
			log.Println("DBDeleteDexSwap error", err)
			continue
		}
		time.Sleep(10 * time.Second)
	}
}

func forwardCollectedFee() {
	for {
		// shardNums, err := incClient.GetActiveShard()
		// if err != nil {
		// 	log.Println("GetActiveShard", err)
		// 	continue
		// }
		time.Sleep(5 * time.Second)
		pendingToken, err := getPendingPappsFee(-1)
		if err != nil {
			log.Println("getPendingPappsFee", err)
			continue
		}

		pendingTokenUnshield, err := getPendingUnshieldsFee(-1)
		if err != nil {
			log.Println("getPendingUnshieldsFee", err)
			continue
		}

		for token, amount := range pendingTokenUnshield {
			pendingToken[token] += amount
		}

		coins, _, err := incClient.GetAllUTXOsV2(config.IncKey)
		if err != nil {
			log.Println("GetAllUTXOsV2", err)
			continue
		}

		amountToSend := make(map[string]uint64)
		totalBalance := make(map[string]uint64)
		for tokenID, coinList := range coins {
			for _, v := range coinList {
				totalBalance[tokenID] += v.GetValue()
			}
		}

		for tokenID, amount := range totalBalance {
			if tokenID == inccommon.PRVCoinID.String() {
				if amount <= 1000000 { // 1000000 0,001 PRV
					continue
				} else {
					amount -= 1000000
				}
			}
			if pendingAmount, exist := pendingToken[tokenID]; exist {
				amount = amount - pendingAmount
				amountToSend[tokenID] = amount
			} else {
				amountToSend[tokenID] = amount
			}
		}

		collectFeeTk := make(map[string]float64)
		for tkID, tkAmount := range amountToSend {
			tkInfo, err := getTokenInfo(tkID)
			if err != nil {
				log.Println("getTokenInfo", tkID, err)
				continue
			}
			amount := new(big.Float).SetUint64(tkAmount)
			decimal := new(big.Float).SetFloat64(math.Pow10(-tkInfo.PDecimals))
			afl64, _ := amount.Mul(amount, decimal).Float64()
			collectFeeTk[tkInfo.Name] = afl64
		}

		collectFeeTkBytes, err := json.MarshalIndent(collectFeeTk, "", "\t")
		if err != nil {
			log.Println("GetAllUTXOsV2", err)
			continue
		}
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[collectedfee]` we have collected\n %v", string(collectFeeTkBytes)))

		if config.CentralIncPaymentAddress != "" {
			for tokenID, amount := range amountToSend {
				time.Sleep(30 * time.Second)
				txhash := ""
				if tokenID == inccommon.PRVCoinID.String() {
					txhash, err = incClient.CreateAndSendRawTransaction(config.IncKey, []string{config.CentralIncPaymentAddress}, []uint64{amount}, 2, nil)
					if err != nil {
						log.Println("GetAllUTXOsV2", err)
						continue
					}
				} else {
					txhash, err = incClient.CreateAndSendRawTokenTransaction(config.IncKey, []string{config.CentralIncPaymentAddress}, []uint64{amount}, tokenID, 2, nil)
					if err != nil {
						log.Println("GetAllUTXOsV2", err)
						continue
					}
				}

				go func(tkID string, tkAmount uint64) {
					tkInfo, _ := getTokenInfo(tkID)
					amount := new(big.Float).SetUint64(tkAmount)
					decimal := new(big.Float).SetFloat64(math.Pow10(-tkInfo.PDecimals))
					afl64, _ := amount.Mul(amount, decimal).Float64()
					slacknoti.SendSlackNoti(fmt.Sprintf("`[withdrawFee]` withdraw `%f %v` fee to central wallet txhash `%v`", afl64, tkInfo.Symbol, txhash))
				}(tokenID, amount)
			}
		}

		time.Sleep(6 * time.Hour)

	}
}

func watchIncAccountBalance() {
	for {
		for _, key := range keyList {
			bl, err := incClient.GetBalance(key, inccommon.PRVCoinID.String())
			if err != nil {
				log.Println("GetBalance", err)
				continue
			}
			log.Println("PRV left:", bl, key)
		}
		time.Sleep(10 * time.Minute)
	}
}

func watchPendingFeeRefundTx() {
	for {
		txList, err := database.DBGetPendingFeeRefundTx(0)
		if err != nil {
			log.Println("DBGetPendingFeeRefundTx err:", err)
		}

		for _, tx := range txList {
			status := tx.RefundStatus
			switch status {
			case wcommon.StatusWaiting:
				_, err := SubmitTxFeeRefund(tx.IncRequestTx, tx.RefundOTA, tx.RefundAddress, tx.RefundToken, tx.RefundAmount, tx.RefundPrivacyFee)
				if err != nil {
					log.Println("SubmitTxFeeRefund err:", err)
					continue
				} else {
					err = database.DBUpdateRefundFeeRefundTx(tx.RefundTx, tx.IncRequestTx, wcommon.StatusSubmitting, "")
					if err != nil {
						log.Println("DBUpdateRefundFeeRefundTx err:", err)
						continue
					}
				}
			case wcommon.StatusPending:
				txDetail, err := incClient.GetTxDetail(tx.RefundTx)
				if err != nil {
					log.Println("CheckTxInBlock err:", err)
				}

				if txDetail == nil {
					if time.Since(tx.UpdatedAt) > 1*time.Hour {
						err = database.DBUpdateRefundFeeRefundTx(tx.RefundTx, tx.IncRequestTx, wcommon.StatusSubmitFailed, "timeout")
						if err != nil {
							log.Println("DBUpdateRefundFeeRefundTx err:", err)
							continue
						}
						go slacknoti.SendSlackNoti(fmt.Sprintf("`[refundfee]` inctx fee refund have submited failed 😵, incReqTx `%v`, incRefund `%v`, isPrivacyFee `%v`\n", tx.IncRequestTx, tx.RefundTx, tx.RefundPrivacyFee))
					}
				} else {
					if txDetail.IsInBlock {
						err = database.DBUpdateRefundFeeRefundTx(tx.RefundTx, tx.IncRequestTx, wcommon.StatusAccepted, "")
						if err != nil {
							log.Println("DBUpdateRefundFeeRefundTx err:", err)
							continue
						}
						go slacknoti.SendSlackNoti(fmt.Sprintf("`[refundfee]` inctx fee refund have accepted 👌, incReqTx `%v`, incRefund `%v`, isPrivacyFee `%v`\n", tx.IncRequestTx, tx.RefundTx, tx.RefundPrivacyFee))
					}
				}
			}
		}
		time.Sleep(20 * time.Second)
	}
}

func watchPappTxNeedFeeRefund() {
	for {
		txList, err := database.DBGetPappTxNeedFeeRefund(0)
		if err != nil {
			log.Println("DBGetPappTxNeedFeeRefund err:", err)
		}
		for _, tx := range txList {
			rftx, err := database.DBGetTxFeeRefundByReq(tx.IncTx)
			if err != nil {
				if err != mongo.ErrNoDocuments {
					log.Println("DBGetTxFeeRefundByReq", err)
					continue
				}
			}
			if rftx != nil {
				err = database.DBUpdatePappRefund(tx.IncTx, true)
				if err != nil {
					log.Println("DBGetTxFeeRefundByReq", err)
					continue
				}
			}
			data := wcommon.RefundFeeData{
				IncRequestTx:     tx.IncTx,
				RefundAmount:     tx.FeeAmount,
				RefundToken:      tx.FeeToken,
				RefundOTA:        tx.FeeRefundOTA,
				RefundAddress:    tx.FeeRefundAddress,
				RefundPrivacyFee: false,
				RefundStatus:     wcommon.StatusWaiting,
			}

			err = database.DBCreateRefundFeeRecord(data)
			if err != nil {
				log.Println("DBGetTxFeeRefundByReq", err)
				continue
			}
		}

		txList, err = database.DBGetPappTxNeedPrivacyFeeRefund(0)
		if err != nil {
			log.Println("DBGetPappTxNeedFeeRefund err:", err)
		}
		for _, tx := range txList {
			rftx, err := database.DBGetTxFeeRefundByReq(tx.IncTx)
			if err != nil {
				if err != mongo.ErrNoDocuments {
					log.Println("DBGetTxFeeRefundByReq", err)
					continue
				}
			}
			if rftx != nil {
				err = database.DBUpdatePappRefund(tx.IncTx, true)
				if err != nil {
					log.Println("DBGetTxFeeRefundByReq", err)
					continue
				}
			}
			data := wcommon.RefundFeeData{
				IncRequestTx:     tx.IncTx,
				RefundAmount:     tx.PFeeAmount,
				RefundToken:      tx.FeeToken,
				RefundOTA:        tx.FeeRefundOTA,
				RefundAddress:    tx.FeeRefundAddress,
				RefundPrivacyFee: true,
				RefundStatus:     wcommon.StatusWaiting,
			}

			err = database.DBCreateRefundFeeRecord(data)
			if err != nil {
				log.Println("DBGetTxFeeRefundByReq", err)
				continue
			}
		}

		time.Sleep(30 * time.Second)
	}
}

func watchRedepositExternalTx() {
	for {
		txList, err := database.DBRetrievePendingRedepositExternalTx(0, 0)
		if err != nil {
			log.Println("DBRetrievePendingExternalTx err:", err)
		}
		for _, tx := range txList {
			networkID := wcommon.GetNetworkID(tx.Network)
			if _, err := database.DBGetShieldTxStatusByExternalTx(tx.Txhash, networkID); err == mongo.ErrNoDocuments {
				_, err := SubmitShieldProof(tx.Txhash, networkID, "", TxTypeRedeposit, false)
				if err != nil {
					log.Println(err)
					continue
				}

			}
		}
		time.Sleep(20 * time.Second)
	}
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
				log.Println("getEVMBlockHeight err:", networkInfo.Network, err)
				go slacknoti.SendSlackNoti(fmt.Sprintf("[externaltx] alert!!! can't get block height for network %v ⚠️", networkInfo.Network))
				continue
			}
			txList, err := database.DBRetrievePendingExternalTx(networkInfo.Network, 0, 0)
			if err != nil {
				log.Println("DBRetrievePendingExternalTx err:", err)
				continue
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
		txList, err := database.DBRetrievePendingPappTxs(wcommon.ExternalTxTypeSwap, 0, 0)
		if err != nil {
			log.Println("DBRetrievePendingPappTxs Swap err:", err)
		}
		for _, txdata := range txList {
			err := processPendingSwapTx(txdata)
			if err != nil {
				log.Println("processPendingShieldTxs Swap err:", txdata.IncTx)
			}
		}

		txList, err = database.DBRetrievePendingPappTxs(wcommon.ExternalTxTypeOpensea, 0, 0)
		if err != nil {
			log.Println("DBRetrievePendingPappTxs OpenSea err:", err)
		}
		for _, txdata := range txList {
			err := processPendingOpenseaTx(txdata)
			if err != nil {
				log.Println("processPendingOpenseaTx OpenSea err:", txdata.IncTx)
			}
		}

		txList, err = database.DBGetPappTxPendingOutchainSubmit(0, 0)
		if err != nil {
			log.Println("DBGetPappTxPendingOutchainSubmit err:", err)
		}
		for _, txdata := range txList {
			tx, err := database.DBGetExternalTxByIncTx(txdata.IncTx, txdata.Networks[0])
			if err != nil {
				log.Println("DBGetExternalTxByIncTx err:", err)
				continue
			}
			if tx != nil {
				if tx.Status == wcommon.StatusAccepted {
					err = database.DBUpdatePappTxSubmitOutchainStatus(txdata.IncTx, wcommon.StatusAccepted)
					if err != nil {
						log.Println("DBGetExternalTxByIncTx err:", err)
						continue
					}
				}
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
				log.Println("processPendingShieldTxs err:", txdata.IncTx, err)
			}
		}
		time.Sleep(10 * time.Second)
	}
}

func processPendingShieldTxs(txdata wcommon.ShieldTxData) error {
	isInBlock, err := incClient.CheckTxInBlock(txdata.IncTx)
	if err != nil {
		if strings.Contains(err.Error(), "RPC returns an error:") {
			err = database.DBUpdateShieldTxStatus(txdata.ExternalTx, txdata.NetworkID, wcommon.StatusSubmitFailed, err.Error())
			if err != nil {
				log.Println("DBUpdateShieldTxStatus err:", err)
				return err
			}
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[shieldtx]` submit shield failed 😵 inctx `%v` network `%v` externaltx `%v` \n", txdata.IncTx, txdata.NetworkID, txdata.ExternalTx))
			return nil
		}
		log.Println("CheckTxInBlock err:", err)
		return err
	}

	if isInBlock {
		var status int
		if txdata.TokenID != txdata.UTokenID {
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
			go faucetPRV(txdata.PaymentAddress)
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[shieldtx]` inctx shield/redeposit have accepted 👌, exttx `%v`, inctx `%v`\n", txdata.ExternalTx, txdata.IncTx))
			return nil
		case 3:
			err = database.DBUpdateShieldTxStatus(txdata.ExternalTx, txdata.NetworkID, wcommon.StatusRejected, "rejected by chain")
			if err != nil {
				log.Println("DBUpdateShieldTxStatus err:", err)
				return err
			}
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[shieldtx]` inctx shield/redeposit have rejected needed check 😵, exttx `%v`\n", txdata.ExternalTx))
			return nil
		}
	}
	return nil
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
			valueBuf := encodeBufferPool.Get().(*bytes.Buffer)
			defer encodeBufferPool.Put(valueBuf)

			vaultABI, err := abi.JSON(strings.NewReader(vault.VaultABI))
			if err != nil {
				fmt.Println("abi.JSON", err.Error())
				return err
			}
			isRedeposit := false
			tokenContract := ""
			amount := new(big.Int)
			for _, d := range txReceipt.Logs {
				switch len(d.Data) {
				case 96:
					unpackResult, err := vaultABI.Unpack("Withdraw", d.Data)
					if err != nil {
						fmt.Println("Unpack2", err)
						continue
					}
					if len(unpackResult) < 3 {
						err = errors.New(fmt.Sprintf("Unpack event not match data needed %v\n", unpackResult))
						fmt.Println("len(unpackResult)2", err)
						continue
					}
					fmt.Println("96", unpackResult[0].(common.Address).String(), unpackResult[1].(common.Address).String(), unpackResult[2].(*big.Int))
					tokenContract = unpackResult[0].(common.Address).String()
					amount = unpackResult[2].(*big.Int)
				case 256, 288:
					topicHash := strings.ToLower(d.Topics[0].String())
					if !strings.Contains(topicHash, "00b45d95b5117447e2fafe7f34def913ff3ba220e4b8688acf37ae2328af7a3d") {
						continue
					}
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
					tokenContract = unpackResult[0].(common.Address).String()
					amount = unpackResult[2].(*big.Int)
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
				LogResult:     logResult,
				IsRedeposit:   isRedeposit,
				IsReverted:    (len(txReceipt.Logs) >= 2) && (len(txReceipt.Logs) <= 3) && (tx.Type != 2),
				IsFailed:      (txReceipt.Status == 0),
				TokenContract: tokenContract,
				Amount:        amount,
			}

			otherInfoBytes, _ := json.MarshalIndent(otherInfo, "", "\t")

			err = database.DBUpdateExternalTxOtherInfo(tx.Txhash, string(otherInfoBytes))
			if err != nil {
				return err
			}
			if isRedeposit {
				err = database.DBUpdateExternalTxWillRedeposit(tx.Txhash, true)
				if err != nil {
					return err
				}
				_, err := SubmitShieldProof(tx.Txhash, networkID, "", TxTypeRedeposit, false)
				if err != nil {
					return err
				}
			}
			if otherInfo.IsReverted {
				err = database.DBUpdatePappRefundPFee(tx.IncRequestTx, true)
				if err != nil {
					return err
				}
			}

			err = database.DBUpdateExternalTxStatus(tx.Txhash, wcommon.StatusAccepted, "")
			if err != nil {
				return err
			}

			txtype := ""
			switch tx.Type {
			case wcommon.ExternalTxTypeSwap:
				txtype = "swaptx"
				err = database.DBUpdatePappTxSubmitOutchainStatus(tx.IncRequestTx, wcommon.StatusAccepted)
				if err != nil {
					return err
				}
				break
			case wcommon.ExternalTxTypeUnshield:
				txtype = "unshield"
				err = database.DBUpdateUnshieldTxSubmitOutchainStatus(tx.IncRequestTx, wcommon.StatusAccepted)
				if err != nil {
					return err
				}
				break
			default:
				txtype = "unknown"
			}
			if otherInfo.IsFailed {
				go slacknoti.SendSlackNoti(fmt.Sprintf("`[%v]` tx outchain have failed outcome needed check 😵, exttx `%v`, network `%v`\n", txtype, tx.Txhash, tx.Network))
			} else {
				if tx.Type == wcommon.ExternalTxTypeSwap {
					go func() {
					retry:
						slackep := os.Getenv("SLACK_SWAP_ALERT")
						if slackep != "" {
							swapAlert := ""
							txIncRequest := tx.IncRequestTx
							pappTxData, err := database.DBGetPappTxData(txIncRequest)
							if err != nil {
								log.Println("DBGetPappTxData", err)
								time.Sleep(5 * time.Second)
								goto retry
							}
							if pappTxData.PappSwapInfo != "" {
								pappSwapInfo := wcommon.PappSwapInfo{}

								err = json.Unmarshal([]byte(pappTxData.PappSwapInfo), &pappSwapInfo)
								if err != nil {
									log.Println("Unmarshal", err)
									return
								}
								tkInInfo, err := getTokenInfo(pappSwapInfo.TokenIn)
								if err != nil {
									log.Println("getTokenInfo1", err)
									time.Sleep(5 * time.Second)
									goto retry
								}
								amount := new(big.Float).SetInt(pappSwapInfo.TokenInAmount)
								decimal := new(big.Float)
								decimalInt, err := getTokenDecimalOnNetwork(tkInInfo, networkID)
								if err != nil {
									log.Println("getTokenDecimalOnNetwork2", err)
									return
								}
								decimal.SetFloat64(math.Pow10(int(-decimalInt)))

								amountInFloat := amount.Mul(amount, decimal).Text('f', -1)
								tokenInSymbol := tkInInfo.Symbol

								tkOutInfo, err := getTokenInfo(pappSwapInfo.TokenOut)
								if err != nil {
									log.Println("getTokenInfo2", err)
									time.Sleep(5 * time.Second)
									goto retry
								}
								amount = new(big.Float).SetInt(pappSwapInfo.MinOutAmount)
								decimalInt, err = getTokenDecimalOnNetwork(tkOutInfo, networkID)
								if err != nil {
									log.Println("getTokenDecimalOnNetwork2", err)
									return
								}
								decimal.SetFloat64(math.Pow10(int(-decimalInt)))
								amountOutFloat := amount.Mul(amount, decimal).Text('f', -1)
								tokenOutSymbol := tkOutInfo.Symbol

								if otherInfo.IsReverted {
									swapAlert = fmt.Sprintf("`[%v(%v)]` swap was reverted 😢\n SwapID: `%v`\n Requested: `%v %v` to `%v %v`\n--------------------------------------------------------", pappSwapInfo.DappName, tx.Network, pappTxData.ID.Hex(), amountInFloat, tokenInSymbol, amountOutFloat, tokenOutSymbol)
								} else {
									amount = new(big.Float).SetInt(otherInfo.Amount)

									if tkOutInfo.CurrencyType == wcommon.UnifiedCurrencyType {
										for _, ctk := range tkOutInfo.ListUnifiedToken {
											netID, _ := wcommon.GetNetworkIDFromCurrencyType(ctk.CurrencyType)
											isNative := false
											if wcommon.GetNativeNetworkCurrencyType(wcommon.GetNetworkName(netID)) == ctk.CurrencyType {
												isNative = true
											}
											if wcommon.CheckIsWrappedNativeToken(ctk.ContractID, netID) {
												isNative = true
											}
											if netID == networkID {
												if isNative {
													decimal = new(big.Float).SetFloat64(math.Pow10(-int(ctk.Decimals)))
												} else {
													if otherInfo.IsRedeposit {
														decimal = new(big.Float).SetFloat64(math.Pow10(-int(ctk.PDecimals)))
													} else {
														decimal = new(big.Float).SetFloat64(math.Pow10(-int(ctk.Decimals)))
													}
												}
												break
											}
										}
									} else {
										netID, _ := wcommon.GetNetworkIDFromCurrencyType(tkOutInfo.CurrencyType)
										isNative := false
										if wcommon.GetNativeNetworkCurrencyType(wcommon.GetNetworkName(netID)) == tkOutInfo.CurrencyType {
											isNative = true
										}
										if wcommon.CheckIsWrappedNativeToken(tkOutInfo.ContractID, netID) {
											isNative = true
										}
										if isNative {
											decimal = new(big.Float).SetFloat64(math.Pow10(-int(tkOutInfo.Decimals)))
										} else {
											if otherInfo.IsRedeposit {
												decimal = new(big.Float).SetFloat64(math.Pow10(-int(tkOutInfo.PDecimals)))
											} else {
												decimal = new(big.Float).SetFloat64(math.Pow10(-int(tkOutInfo.Decimals)))
											}
										}
									}

									uaStr := parseUserAgent(pappTxData.UserAgent)
									// decimal = new(big.Float).SetFloat64(math.Pow10(int(-decimalInt)))
									realOutFloat := amount.Mul(amount, decimal).Text('f', -1)
									swapAlert = fmt.Sprintf("`[%v(%v) | %v]` swap was success 🎉\n SwapID: `%v`\n Requested: `%v %v` to `%v %v` | received: `%v %v`\n--------------------------------------------------------", pappSwapInfo.DappName, tx.Network, uaStr, pappTxData.ID.Hex(), amountInFloat, tokenInSymbol, amountOutFloat, tokenOutSymbol, realOutFloat, tokenOutSymbol)
								}
								log.Println(swapAlert)
								slacknoti.SendWithCustomChannel(swapAlert, slackep)
							}
						}
					}()
				}
			}
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[%v]` tx outchain accepted exttx `%v`, network `%v`, incReqTx `%v`\n outcome of tx: `%v`\n", txtype, tx.Txhash, tx.Network, tx.IncRequestTx, string(otherInfoBytes)))
		}
		return nil
	}
	return errors.New("no endpoints reachable")
}

func processPendingSwapTx(tx wcommon.PappTxData) error {
	txDetail, err := incClient.GetTxDetail(tx.IncTx)
	if err != nil {
		if strings.Contains(err.Error(), "RPC returns an error:") {
			err = database.DBUpdatePappTxStatus(tx.IncTx, wcommon.StatusSubmitFailed, err.Error())
			if err != nil {
				log.Println("DBUpdateShieldTxStatus err:", err)
				return err
			}
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[swaptx]` submit swaptx failed 😵 `%v` \n", tx.IncTx))
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
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[swaptx]` inctx `%v` rejected by beacon 😢\n", tx.IncTx))
		case 1:
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[swaptx]` inctx `%v` accepted by beacon 👌\n", tx.IncTx))
			err = database.DBUpdatePappTxStatus(tx.IncTx, wcommon.StatusAccepted, "")
			if err != nil {
				return err
			}
			err = database.DBUpdatePappTxSubmitOutchainStatus(tx.IncTx, wcommon.StatusWaiting)
			if err != nil {
				return err
			}
			for _, network := range tx.Networks {
				_, err := SubmitOutChainTx(tx.IncTx, network, tx.IsUnifiedToken, false, wcommon.ExternalTxTypeSwap)
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

func watchEVMAccountBalance() {
	for {
		networks, err := database.DBGetBridgeNetworkInfos()
		if err != nil {
			log.Println("DBGetBridgeNetworkInfos err:", err)
		}
		privKey, _ := crypto.HexToECDSA(config.EVMKey)
		keyAddr := crypto.PubkeyToAddress(privKey.PublicKey)

		feeData, err := database.DBRetrieveFeeTable()
		if err != nil {
			log.Println("DBRetrieveFeeTable err:", err)
		}

		balanceResult := make(map[string]string)

		for _, networkInfo := range networks {
			for _, endpoint := range networkInfo.Endpoints {
				evmClient, err := ethclient.Dial(endpoint)
				if err != nil {
					balanceResult[networkInfo.Network] = err.Error()
					log.Println(err)
					continue
				}
				feeLeft, err := evmClient.BalanceAt(context.Background(), keyAddr, nil)
				if err != nil {
					balanceResult[networkInfo.Network] = err.Error()
					log.Println(err)
					continue
				}
				log.Printf("network %v has %v gas left\n", networkInfo.Network, feeLeft.Uint64())

				gasPrice, ok := feeData.GasPrice[networkInfo.Network]
				if !ok {
					balanceResult[networkInfo.Network] = "no gasprice"
					log.Printf("network %v have no gasprice\n", networkInfo.Network)
					continue
				}
				gasPriceBig := new(big.Int).SetUint64(gasPrice)
				gasLimitBig := new(big.Int).SetUint64(wcommon.EVMGasLimit)

				feeFloat := new(big.Float).SetInt(feeLeft)
				feeFloat.Mul(feeFloat, new(big.Float).SetFloat64(math.Pow10(-18)))

				feeLeft = feeLeft.Div(feeLeft, gasPriceBig)
				txLeft := feeLeft.Div(feeLeft, gasLimitBig)

				log.Printf("network %v estimted has %v txs left (\n", networkInfo.Network, txLeft.Uint64())

				if txLeft.Uint64() <= wcommon.MinEVMTxs {
					balanceResult[networkInfo.Network] = fmt.Sprintf("%f low fee ⚠️", feeFloat)
				} else {
					balanceResult[networkInfo.Network] = fmt.Sprintf("%f", feeFloat)
				}
				break
			}
		}
		slacktext := "`[networkfee]`\n"
		for network, v := range balanceResult {
			t := fmt.Sprintf("%v: %v\n", network, v)
			slacktext = slacktext + t
		}

		go slacknoti.SendSlackNoti(slacktext)
		time.Sleep(30 * time.Minute)
	}

}

func getPendingPappsFee(shardID int) (map[string]uint64, error) {
	result := make(map[string]uint64)
	var txList []wcommon.PappTxData
	var err error
	if shardID == -1 {
		txList, err = database.DBGetPappTxDataByStatus(wcommon.StatusExecuting, 0, 0)
		if err != nil {
			return nil, err
		}
	} else {
		txList, err = database.DBGetPappTxDataByStatusAndShardID(wcommon.StatusExecuting, shardID, 0, 0)
		if err != nil {
			return nil, err
		}
	}

	for _, v := range txList {
		result[v.FeeToken] += v.FeeAmount
	}

	txList, err = database.DBGetPappTxPendingOutchainSubmit(0, 0)
	if err != nil {
		return nil, err
	}
	for _, v := range txList {
		result[v.FeeToken] += v.FeeAmount
	}

	txRefundFeeWaitList, err := database.DBGetPendingFeeRefundTx(0)
	if err != nil {
		log.Println("DBGetPappTxNeedFeeRefund err:", err)
	}

	for _, tx := range txRefundFeeWaitList {
		status := tx.RefundStatus
		switch status {
		case wcommon.StatusWaiting, wcommon.StatusSubmitFailed, wcommon.StatusPending:
			// _, err := SubmitTxFeeRefund(tx.IncRequestTx, tx.RefundOTA, tx.RefundOTASS, tx.RefundAddress, tx.RefundToken, tx.RefundAmount)
			result[tx.RefundToken] += tx.RefundAmount
		}
	}

	return result, nil
}

func getPendingUnshieldsFee(shardID int) (map[string]uint64, error) {
	result := make(map[string]uint64)
	var txList []wcommon.UnshieldTxData
	var err error
	if shardID == -1 {
		txList, err = database.DBGetUnshieldTxDataByStatus(wcommon.StatusExecuting, 0, 0)
		if err != nil {
			return nil, err
		}
	} else {
		txList, err = database.DBGetUnshieldTxDataByStatusAndShardID(wcommon.StatusExecuting, shardID, 0, 0)
		if err != nil {
			return nil, err
		}
	}

	for _, v := range txList {
		result[v.FeeToken] += v.FeeAmount
	}

	txList, err = database.DBGetUnshieldTxPendingOutchainSubmit(0, 0)
	if err != nil {
		return nil, err
	}
	for _, v := range txList {
		result[v.FeeToken] += v.FeeAmount
	}

	txRefundFeeWaitList, err := database.DBGetPendingFeeRefundTx(0)
	if err != nil {
		log.Println("DBGetPendingFeeRefundTx err:", err)
	}

	for _, tx := range txRefundFeeWaitList {
		status := tx.RefundStatus
		switch status {
		case wcommon.StatusWaiting, wcommon.StatusSubmitFailed, wcommon.StatusPending:
			result[tx.RefundToken] += tx.RefundAmount
		}
	}

	return result, nil
}

func watchPendingUnshieldTx() {
	for {
		txList, err := database.DBRetrievePendingUnshieldTxs(0, 0)
		if err != nil {
			log.Println("DBRetrievePendingUnshieldTxs err:", err)
		}
		for _, txdata := range txList {
			err := processPendingUnshieldTx(txdata)
			if err != nil {
				log.Println("processPendingShieldTxs err:", txdata.IncTx)
			}
		}

		txList, err = database.DBGetUnshieldTxPendingOutchainSubmit(0, 0)
		if err != nil {
			log.Println("DBGetUnshieldTxPendingOutchainSubmit err:", err)
		}
		for _, txdata := range txList {
			tx, err := database.DBGetExternalTxByIncTx(txdata.IncTx, txdata.Networks[0])
			if err != nil {
				log.Println("DBGetExternalTxByIncTx err:", err)
				continue
			}
			if tx != nil {
				if tx.Status == wcommon.StatusAccepted {
					err = database.DBUpdateUnshieldTxSubmitOutchainStatus(txdata.IncTx, wcommon.StatusAccepted)
					if err != nil {
						log.Println("DBUpdateUnshieldTxSubmitOutchainStatus err:", err)
						continue
					}
				}
			}
		}
		time.Sleep(10 * time.Second)
	}
}

func processPendingUnshieldTx(tx wcommon.UnshieldTxData) error {
	txDetail, err := incClient.GetTxDetail(tx.IncTx)
	if err != nil {
		if strings.Contains(err.Error(), "RPC returns an error:") {
			err = database.DBUpdateUnshieldTxStatus(tx.IncTx, wcommon.StatusSubmitFailed, err.Error())
			if err != nil {
				log.Println("DBUpdateUnshieldTxStatus err:", err)
				return err
			}
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[unshield]` submit unshield failed 😵 `%v` \n", tx.IncTx))
			return nil
		}
		return err
	}
	if txDetail.IsInBlock {
		if tx.IsUnifiedToken {
			status, err := checkBeaconBridgeAggUnshieldStatus(tx.IncTx)
			if err != nil {
				return err
			}

			switch status {
			case 0:
				err = database.DBUpdateUnshieldTxStatus(tx.IncTx, wcommon.StatusRejected, "")
				if err != nil {
					return err
				}
				go slacknoti.SendSlackNoti(fmt.Sprintf("`[unshield]` inctx `%v` rejected by beacon 😢\n", tx.IncTx))
			case 1, 3:
				go slacknoti.SendSlackNoti(fmt.Sprintf("`[unshield]` inctx `%v` accepted by beacon 👌\n", tx.IncTx))
				err = database.DBUpdateUnshieldTxStatus(tx.IncTx, wcommon.StatusAccepted, "")
				if err != nil {
					return err
				}
				err = database.DBUpdateUnshieldTxSubmitOutchainStatus(tx.IncTx, wcommon.StatusWaiting)
				if err != nil {
					return err
				}
				for _, network := range tx.Networks {
					_, err := SubmitOutChainTx(tx.IncTx, network, tx.IsUnifiedToken, false, wcommon.ExternalTxTypeUnshield)
					if err != nil {
						return err
					}
				}
			default:
				if tx.Status != wcommon.StatusExecuting && tx.Status != wcommon.StatusAccepted {
					err = database.DBUpdateUnshieldTxStatus(tx.IncTx, wcommon.StatusExecuting, "")
					if err != nil {
						return err
					}
				}
			}
		} else {
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[unshield]` inctx `%v` accepted by beacon 👌\n", tx.IncTx))
			err = database.DBUpdateUnshieldTxStatus(tx.IncTx, wcommon.StatusAccepted, "")
			if err != nil {
				return err
			}
			err = database.DBUpdateUnshieldTxSubmitOutchainStatus(tx.IncTx, wcommon.StatusWaiting)
			if err != nil {
				return err
			}
			for _, network := range tx.Networks {
				_, err := SubmitOutChainTx(tx.IncTx, network, tx.IsUnifiedToken, false, wcommon.ExternalTxTypeUnshield)
				if err != nil {
					return err
				}
			}
		}

	}
	return nil
}

func watchUnshieldTxNeedFeeRefund() {
	for {
		txList, err := database.DBGetUnshieldTxNeedFeeRefund(0)
		if err != nil {
			log.Println("DBGetUnshieldTxNeedFeeRefund err:", err)
		}
		for _, tx := range txList {
			rftx, err := database.DBGetTxFeeRefundByReq(tx.IncTx)
			if err != nil {
				if err != mongo.ErrNoDocuments {
					log.Println("DBGetTxFeeRefundByReq", err)
					continue
				}
			}
			if rftx != nil {
				err = database.DBUpdateUnshieldRefund(tx.IncTx, true)
				if err != nil {
					log.Println("DBUpdateUnshieldRefund", err)
					continue
				}
			}
			data := wcommon.RefundFeeData{
				IncRequestTx:     tx.IncTx,
				RefundAmount:     tx.FeeAmount,
				RefundToken:      tx.FeeToken,
				RefundOTA:        tx.FeeRefundOTA,
				RefundAddress:    tx.FeeRefundAddress,
				RefundPrivacyFee: false,
				RefundStatus:     wcommon.StatusWaiting,
			}

			err = database.DBCreateRefundFeeRecord(data)
			if err != nil {
				log.Println("DBGetTxFeeRefundByReq", err)
				continue
			}
		}

		time.Sleep(30 * time.Second)
	}
}

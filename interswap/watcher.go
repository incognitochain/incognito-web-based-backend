package interswap

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/incognitochain/go-incognito-sdk-v2/coin"
	metadataBridge "github.com/incognitochain/go-incognito-sdk-v2/metadata/bridge"
	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/slacknoti"
)

// func

func watchInterswapPendingTx() {
	for {
		firstPendingTxs, err := database.DBRetrieveTxsByStatus([]int{FirstPending}, 0, 0)
		if err != nil {
			log.Println("DBRetrieveTxsByStatus err:", err)
		}
		for _, txdata := range firstPendingTxs {
			err := processInterswapPendingFirstTx(txdata)
			if err != nil {
				log.Println("processPendingShieldTxs err:", txdata.TxID)
			}
		}

		secondPendingTxs, err := database.DBRetrieveTxsByStatus([]int{SecondPending}, 0, 0)
		if err != nil {
			log.Println("DBRetrieveTxsByStatus err:", err)
		}
		for _, txdata := range secondPendingTxs {
			err := processInterswapPendingFirstTx(txdata)
			if err != nil {
				log.Println("processPendingShieldTxs err:", txdata.TxID)
			}
		}

		// txList, err = database.DBGetPappTxPendingOutchainSubmit(0, 0)
		// if err != nil {
		// 	log.Println("DBGetPappTxPendingOutchainSubmit err:", err)
		// }
		// for _, txdata := range txList {
		// 	tx, err := database.DBGetExternalTxByIncTx(txdata.IncTx, txdata.Networks[0])
		// 	if err != nil {
		// 		log.Println("DBGetExternalTxByIncTx err:", err)
		// 		continue
		// 	}
		// 	if tx != nil {
		// 		if tx.Status == wcommon.StatusAccepted {
		// 			err = database.DBUpdatePappTxSubmitOutchainStatus(txdata.IncTx, wcommon.StatusAccepted)
		// 			if err != nil {
		// 				log.Println("DBGetExternalTxByIncTx err:", err)
		// 				continue
		// 			}
		// 		}
		// 	}
		// }
		time.Sleep(10 * time.Second)
	}
}

func processInterswapPendingFirstTx(txData common.InterSwapTxData) error {
	interswapTxID := txData.TxID

	// check tx by hash
	txDetail, err := incClient.GetTxDetail(interswapTxID)
	if err != nil {
		if strings.Contains(err.Error(), "RPC returns an error:") {
			err = database.DBUpdateInterswapTxStatus(interswapTxID, SubmitFailed, StatusStr[SubmitFailed], err.Error())
			if err != nil {
				log.Println("DBUpdateShieldTxStatus err:", err)
				return err
			}
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[swaptx]` submit swaptx failed 😵 `%v` \n", interswapTxID))
			return nil
		}
		return err
	}
	if !txDetail.IsInBlock {
		return nil
	}

	switch txData.PathType {
	case PdexToPApp:
		{
			_, pdexStatus, err := CallGetPdexSwapTxStatus(interswapTxID, txData.MidToken)
			if err != nil {
				log.Printf("CallGetPdexSwapTxStatus TxID %v error %v ", interswapTxID, err)
				return err
			}

			if len(pdexStatus.RespondTxs) > 1 {
				if pdexStatus.Status == "accepted" {
					// parse tx response to get received UTXO
					if len(pdexStatus.RespondTxs) != 1 {
						log.Println("CallGetPdexSwapTxStatus error", err)
						return err
					}

					// TODO: validate

					amtMidToken := pdexStatus.RespondAmounts[0]
					amtMidStr, err := convertToWithoutDecStr(amtMidToken, txData.MidToken)
					if err != nil {
						log.Println("convert amount mid token to string error", err)
						return err
					}

					// re-estimate with addon tx
					addonParamEst := &EstimateSwapParam{
						Network:   txData.PAppNetwork,
						Amount:    amtMidStr,
						FromToken: txData.MidToken,
						ToToken:   txData.ToToken,
						Slippage:  txData.Slippage,
					}
					p2Bytes, _ := json.Marshal(addonParamEst)
					fmt.Printf("addonParamEst 2: %s\n", string(p2Bytes))

					est2, err := CallEstimateSwap(addonParamEst, config)
					if err != nil {
						log.Println("Estimate swap for addon tx error", err)
						return err
					}
					// if estimate papp is not available or not meet MinAcceptable, refund MidToken for users
					isMidRefund := false
					pAppAddOn := new(QuoteData)

					pApps := est2.Networks[txData.PAppNetwork]
					if len(pApps) > 0 {
						for _, pApp := range pApps {
							if pApp.CallContract == txData.PAppContract {
								pAppAddOn = &pApp
								break
							}
						}
						if pAppAddOn == nil {
							log.Printf("InterswapID %v Not found trade path for add on tx\n", interswapTxID)
							isMidRefund = true
						}

					} else {
						isMidRefund = true
					}

					if isMidRefund {
						refundTxID, err := createTxRefund(config.ISIncPrivKey, txData.OTARefund, txData.MidToken, amtMidToken, []coin.PlainCoin{}, []uint{})
						if err != nil {
							log.Printf("InterswapID %v create tx refund error %v\n", interswapTxID, err)
							return fmt.Errorf("InterswapID %v create tx refund error %v\n", interswapTxID, err)
						}
						updateInfo := map[string]interface{}{
							"txidrefund": refundTxID,
							"status":     MidRefunding,
							"statusstr":  StatusStr[MidRefunding],
						}
						err = database.DBUpdateInterswapTxInfo(interswapTxID, updateInfo)
						if err != nil {
							log.Printf("InterswapID %v Update info %+v error %v\n", interswapTxID, updateInfo, err)
							return fmt.Errorf("InterswapID %v Update info %+v error %v\n", interswapTxID, updateInfo, err)
						}

						return nil
					}

					// log.Printf("InterswapID %v Not found trade path for add on tx\n", interswapTxID)

					// create addon tx
					data := metadataBridge.BurnForCallRequestData{}
					tx, err := incClient.CreateAndSendBurnForCallRequestTransaction(config.ISIncPrivKey, txData.MidToken, data, []string{}, []uint64{},
						[]coin.PlainCoin{}, []uint64{}, []coin.PlainCoin{}, []uint64{})
					fmt.Printf("tx addon: %v - Err %v", tx, err)

					// // update addon swap info: amountFrom
					// updatedAddonSwapInfo := task.AddOnSwapInfo

					// // re-calculate AmountIn for AddOn tx
					// midTokenAmt := pdexStatus.RespondAmounts[0]
					// amountStrMidToken := convertToWithoutDecStr(pdexStatus.RespondAmounts[0], pdexStatus.RespondTokens[0])

					// updatedAddonSwapInfo.AmountIn = amountStrMidToken
					// updatedAddonSwapInfo.AmountInRaw = pdexStatus.RespondAmounts[0]

					// // check minAcceptedAmount of AddOn tx is still valid or not

				} else if pdexStatus.Status == "refund" {

				} else {

				}

			}
			// status, err := checkBeaconBridgeAggUnshieldStatus(txData.IncTx)
			// if err != nil {
			// 	return err
			// }

			// switch status {
			// case 0:
			// 	err = database.DBUpdatePappTxStatus(txData.IncTx, wcommon.StatusRejected, "")
			// 	if err != nil {
			// 		return err
			// 	}
			// 	go slacknoti.SendSlackNoti(fmt.Sprintf("`[swaptx]` inctx `%v` rejected by beacon 😢\n", txData.IncTx))
			// case 1:
			// 	go slacknoti.SendSlackNoti(fmt.Sprintf("`[swaptx]` inctx `%v` accepted by beacon 👌\n", txData.IncTx))
			// 	err = database.DBUpdatePappTxStatus(txData.IncTx, wcommon.StatusAccepted, "")
			// 	if err != nil {
			// 		return err
			// 	}
			// 	err = database.DBUpdatePappTxSubmitOutchainStatus(txData.IncTx, wcommon.StatusWaiting)
			// 	if err != nil {
			// 		return err
			// 	}
			// 	for _, network := range txData.Networks {
			// 		_, err := SubmitOutChainTx(txData.IncTx, network, txData.IsUnifiedToken, false, wcommon.ExternalTxTypeSwap)
			// 		if err != nil {
			// 			return err
			// 		}
			// 	}
			// default:
			// 	if txData.Status != wcommon.StatusExecuting && txData.Status != wcommon.StatusAccepted {
			// 		err = database.DBUpdatePappTxStatus(txData.IncTx, wcommon.StatusExecuting, "")
			// 		if err != nil {
			// 			return err
			// 		}
			// 	}
			// }

			return nil

		}
	default:
		{

		}
	}
	return nil

}

func createTxRefund(
	senderPrivKey, otaReceiver, tokenID string, amount uint64,
	utxos []coin.PlainCoin, utxoIndices []uint,
) (string, error) {
	return "", nil

}

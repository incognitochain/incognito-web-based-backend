package interswap

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/incognitochain/go-incognito-sdk-v2/coin"
	"github.com/incognitochain/go-incognito-sdk-v2/common"
	metadataBridge "github.com/incognitochain/go-incognito-sdk-v2/metadata/bridge"
	beCommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/slacknoti"
)

// func

func watchInterswapPendingTx(config beCommon.Config) {
	for {
		firstPendingTxs, err := database.DBRetrieveTxsByStatus([]int{FirstPending}, 0, 0)
		if err != nil {
			log.Println("DBRetrieveTxsByStatus err:", err)
		}
		for _, txdata := range firstPendingTxs {
			err := processInterswapPendingFirstTx(txdata, config)
			if err != nil {
				log.Println("processPendingShieldTxs err:", txdata.TxID)
			}
		}

		secondPendingTxs, err := database.DBRetrieveTxsByStatus([]int{SecondPending}, 0, 0)
		if err != nil {
			log.Println("DBRetrieveTxsByStatus err:", err)
		}
		for _, txdata := range secondPendingTxs {
			err := processInterswapPendingFirstTx(txdata, config)
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
		// 		if tx.Status == wbeCommon.StatusAccepted {
		// 			err = database.DBUpdatePappTxSubmitOutchainStatus(txdata.IncTx, wbeCommon.StatusAccepted)
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

func processInterswapPendingFirstTx(txData beCommon.InterSwapTxData, config beCommon.Config) error {
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
			_, pdexStatus, err := CallGetPdexSwapTxStatus(interswapTxID, txData.MidToken, config)
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

					// TODO: 0xkraken  validate

					amtMidToken := pdexStatus.RespondAmounts[0]
					amtMidStr, err := convertToWithoutDecStr(amtMidToken, txData.MidToken, config)
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

					pAppAddOn, isMidRefund, err := callEstimateSwapAndValidation(addonParamEst, txData.FinalMinExpectedAmt, txData.PAppNetwork, txData.PAppContract, txData.MidToken, 0, interswapTxID)
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
					if err != nil {
						return err
					}

					addonSwapAmt := amtMidToken - pAppAddOn.Fee[0].Amount
					addonFeeAmt := pAppAddOn.Fee[0].Amount
					amtTmp, err := convertFloat64ToWithoutDecStr(addonSwapAmt, txData.MidToken, config)
					if err != nil {
						log.Printf("InterswapID %v Convert amount from to string %+v error %v\n", interswapTxID, addonSwapAmt, err)
						return fmt.Errorf("InterswapID %v Convert amount from to string %+v error %v\n", interswapTxID, addonSwapAmt, err)
					}
					addonParamEst.Amount = amtTmp

					// estimate with final addon amount
					pAppAddOn, isMidRefund, err = callEstimateSwapAndValidation(addonParamEst, txData.FinalMinExpectedAmt, txData.PAppNetwork, txData.PAppContract, txData.MidToken, addonFeeAmt, interswapTxID)
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
					if err != nil {
						return err
					}

					// TODO: get child token of unified token
					childTokenIDStr, err := getChildTokenUnified(txData.MidToken, beCommon.GetNetworkID(txData.PAppNetwork), config)
					if err != nil {
						return err
					}
					childTokenID, err := common.Hash{}.NewHashFromStr(childTokenIDStr)
					if err != nil {
						return err
					}
					redepositAddress := new(coin.OTAReceiver)
					err = redepositAddress.FromString(txData.OTAToToken)
					if err != nil {
						log.Printf("InterswapID %v Invalid  OTAToToken %v\n", interswapTxID, err)
						return fmt.Errorf("InterswapID %v Invalid  OTAToToken %v\n", interswapTxID, err)
					}

					// create addon tx
					data := metadataBridge.BurnForCallRequestData{
						BurningAmount:       addonSwapAmt,
						ExternalNetworkID:   uint8(beCommon.GetNetworkID(txData.PAppNetwork)),
						IncTokenID:          *childTokenID,
						ExternalCalldata:    pAppAddOn.Calldata,
						ExternalCallAddress: txData.PAppContract,
						ReceiveToken:        txData.ToToken,
						RedepositReceiver:   *redepositAddress, // user OTA
						WithdrawAddress:     txData.WithdrawAddress,
					}
					txBytes, addOnTxID, err := incClient.CreateBurnForCallRequestTransaction(config.ISIncPrivKey, txData.MidToken, data, []string{}, []uint64{},
						[]coin.PlainCoin{}, []uint64{}, []coin.PlainCoin{}, []uint64{})
					if err != nil {
						log.Printf("InterswapID %v Create papp swap tx error %v\n", interswapTxID, err)
						return fmt.Errorf("InterswapID %v Create papp swap tx error %v\n", interswapTxID, err)
					}

					// submit addon tx to BE
					_, err = CallSubmitPappSwapTx(string(txBytes), addOnTxID, txData.OTARefundFee, config)
					if err != nil {
						log.Printf("InterswapID %v Submit papp swap tx error %v\n", interswapTxID, err)
						return fmt.Errorf("InterswapID %v Submit papp swap tx error %v\n", interswapTxID, err)
					}

					// update db
					updateInfo := map[string]interface{}{
						"addon_txid": addOnTxID,
						"status":     SecondPending,
						"statusstr":  StatusStr[SecondPending],
					}
					err = database.DBUpdateInterswapTxInfo(interswapTxID, updateInfo)
					if err != nil {
						log.Printf("InterswapID %v Update info %+v error %v\n", interswapTxID, updateInfo, err)
						return fmt.Errorf("InterswapID %v Update info %+v error %v\n", interswapTxID, updateInfo, err)
					}

					return nil

				} else if pdexStatus.Status == "refund" {
					err = database.DBUpdateInterswapTxStatus(interswapTxID, FirstRefunded, StatusStr[FirstRefunded], "")
					if err != nil {
						log.Printf("InterswapID %v Update status %+v error %v\n", interswapTxID, FirstRefunded, err)
						return fmt.Errorf("InterswapID %v Update status %+v error %v\n", interswapTxID, FirstRefunded, err)
					}
					return nil
				}
			}
		}
	case PAppToPdex:
		{
			pAppStatus, err := CallGetPappSwapTxStatus(interswapTxID, config)
			if err != nil {
				log.Printf("CallGetPappSwapTxStatus TxID %v error %v ", interswapTxID, err)
				return err
			}

			switch pAppStatus.BurnStatus {
			case beCommon.StatusSubmitFailed:
				{
					// fullnode reject
					// update status to Failed
					return nil
				}
			case beCommon.StatusRejected:
				{
					// beacon rejected
					// TODO: get tx response
					// create refund tx
					return nil

				}
			}

			switch pAppStatus.SwapStatus {
			case "reverted":
				{
					// tx swap outchain revert
					// TODO: get tx response
					// create refund tx
				}
			case "failed":
				{
					// TODO: review this status
					// tx swap outchain revert
					// TODO: get tx response
					// create refund tx
				}
			case "success":
				{
					// update status

					if pAppStatus.IsRedeposit {
						// wait to deposit success
						if pAppStatus.IsRedeposited {
							// create addon tx
							// re-estimate addon tx
							amtMidStr, err := addStrs(pAppStatus.BuyAmount, pAppStatus.Reward)
							if err != nil {
								return err
							}
							amtMidToken, err := convertToDecAmtUint64(amtMidStr, txData.MidToken, config)
							if err != nil {
								return err
							}

							// re-estimate with addon tx with pdex
							addonParamEst := &EstimateSwapParam{
								Network:   IncNetworkStr,
								Amount:    amtMidStr,
								FromToken: txData.MidToken,
								ToToken:   txData.ToToken,
								Slippage:  txData.Slippage,
							}
							p2Bytes, _ := json.Marshal(addonParamEst)
							fmt.Printf("addonParamEst 2: %s\n", string(p2Bytes))

							pDexAddOn, isMidRefund, err := callEstimateSwapAndValidation(addonParamEst, txData.FinalMinExpectedAmt, IncNetworkStr, "", txData.MidToken, 0, interswapTxID)
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
							if err != nil {
								return err
							}

							addonSwapAmt := amtMidToken - pDexAddOn.Fee[0].Amount
							addonFeeAmt := pDexAddOn.Fee[0].Amount
							amtTmp, err := convertFloat64ToWithoutDecStr(addonSwapAmt, txData.MidToken, config)
							if err != nil {
								log.Printf("InterswapID %v Convert amount from to string %+v error %v\n", interswapTxID, addonSwapAmt, err)
								return fmt.Errorf("InterswapID %v Convert amount from to string %+v error %v\n", interswapTxID, addonSwapAmt, err)
							}
							addonParamEst.Amount = amtTmp

							// estimate with final addon amount
							pDexAddOn, isMidRefund, err = callEstimateSwapAndValidation(addonParamEst, txData.FinalMinExpectedAmt, IncNetworkStr, "", txData.MidToken, addonFeeAmt, interswapTxID)
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
							if err != nil {
								return err
							}

							addOnTxID, err := incClient.CreateAndSendPdexv3TradeTransaction(
								config.ISIncPrivKey, pDexAddOn.PoolPairs, txData.MidToken, txData.ToToken, addonSwapAmt, txData.FinalMinExpectedAmt, addonFeeAmt, false)
							if err != nil {
								log.Printf("InterswapID %v Create pdex tx error %v\n", interswapTxID, err)
								return fmt.Errorf("InterswapID %v Create pdex tx error %v\n", interswapTxID, err)
							}

							// update db
							updateInfo := map[string]interface{}{
								"addon_txid": addOnTxID,
								"status":     SecondPending,
								"statusstr":  StatusStr[SecondPending],
							}
							err = database.DBUpdateInterswapTxInfo(interswapTxID, updateInfo)
							if err != nil {
								log.Printf("InterswapID %v Update info %+v error %v\n", interswapTxID, updateInfo, err)
								return fmt.Errorf("InterswapID %v Update info %+v error %v\n", interswapTxID, updateInfo, err)
							}

							return nil
						}

					} else {
						return fmt.Errorf("InterswapID is not redeposit: %v", interswapTxID)
					}
				}
			}
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

func callEstimateSwapAndValidation(
	params *EstimateSwapParam,
	expectedMinAmount uint64,
	expectedNetwork string,
	expectedCallContract string,
	expectedFeeToken string,
	expectedFeeAmount uint64,
	interswapTxID string,
) (pAppAddOn *QuoteData, isMidRefund bool, errRes error) {
	est2, err := CallEstimateSwap(params, config)
	if err != nil {
		log.Printf("InterswapID %v Estimate swap for addon tx error %v", interswapTxID, err)
		return nil, false, err
	}
	// if estimate papp is not available or not meet MinAcceptable, refund MidToken for users
	isMidRefund = false
	pAppAddOn = new(QuoteData)

	pApps := est2.Networks[expectedNetwork]
	if len(pApps) > 0 {
		for _, pApp := range pApps {
			if pApp.CallContract == expectedCallContract {
				pAppAddOn = &pApp
				break
			}
		}
		if pAppAddOn == nil {
			log.Printf("InterswapID %v Not found trade path for add on tx\n", interswapTxID)
			errRes = fmt.Errorf("InterswapID %v Not found trade path for add on tx\n", interswapTxID)
			isMidRefund = true
			return
		}

		minAmountOut, err := strToFloat64(pAppAddOn.AmountOutRaw)
		if err != nil {
			log.Printf("InterswapID %v Addon Estimate swap can not convert AmountOutRaw err %v\n", interswapTxID, err)
			errRes = fmt.Errorf("InterswapID %v Addon Estimate swap can not convert AmountOutRaw err %v\n", interswapTxID, err)
			return
		}

		if uint64(minAmountOut) < expectedMinAmount {
			log.Printf("InterswapID %v Addon Estimate swap %v not valid with FinalMinExpectedAmt\n", interswapTxID, uint64(minAmountOut))
			isMidRefund = true
			errRes = fmt.Errorf("InterswapID %v Addon Estimate swap %v not valid with FinalMinExpectedAmt\n", interswapTxID, uint64(minAmountOut))
			return
		}

		if pAppAddOn.Fee[0].TokenID != expectedFeeToken {
			log.Printf("InterswapID %v Estimate swap fee invalid, expected %v, got %v\n", interswapTxID, expectedFeeToken, pAppAddOn.Fee[0].TokenID)
			errRes = fmt.Errorf("InterswapID %v Estimate swap fee invalid, expected %v, got %v\n", interswapTxID, expectedFeeToken, pAppAddOn.Fee[0].TokenID)
			return
		}

		if expectedFeeAmount > 0 {
			if pAppAddOn.Fee[0].Amount <= expectedFeeAmount {
				log.Printf("InterswapID %v Estimate swap fee invalid, expected not greater than %v, got %v\n", interswapTxID, expectedFeeAmount, pAppAddOn.Fee[0].Amount)
				errRes = fmt.Errorf("InterswapID %v Estimate swap fee invalid, expected not greater than %v, got %v\n", interswapTxID, expectedFeeAmount, pAppAddOn.Fee[0].Amount)
				return
			}
		}
	} else {
		log.Printf("InterswapID %v Not found trade path for add on tx\n", interswapTxID)
		isMidRefund = true
	}

	return pAppAddOn, isMidRefund, nil

}

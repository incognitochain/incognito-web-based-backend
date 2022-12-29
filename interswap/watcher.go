package interswap

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/incognitochain/go-incognito-sdk-v2/coin"
	"github.com/incognitochain/go-incognito-sdk-v2/common"
	"github.com/incognitochain/go-incognito-sdk-v2/common/base58"
	"github.com/incognitochain/go-incognito-sdk-v2/metadata"
	metadataBridge "github.com/incognitochain/go-incognito-sdk-v2/metadata/bridge"
	metadataCommon "github.com/incognitochain/go-incognito-sdk-v2/metadata/common"
	metadataPdexv3 "github.com/incognitochain/go-incognito-sdk-v2/metadata/pdexv3"
	beCommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"go.mongodb.org/mongo-driver/mongo"
)

// watchInterswapPendingTx listen status of interswap and handle the next step
func watchInterswapPendingTx(config beCommon.Config) {
	log.Println("Starting Interswap Watcher")
	for {

		firstPendingTxs, err := database.DBRetrieveInterswapTxsByStatus([]int{FirstPending}, 0, 0)
		if err != nil {
			log.Println("DBRetrieveTxsByStatus err:", err)
		}
		for _, txdata := range firstPendingTxs {
			err := processInterswapPendingFirstTx(txdata, config)
			if err != nil {
				log.Println("processInterswapPendingFirstTx err:", txdata.TxID)
			}
		}

		secondPendingTxs, err := database.DBRetrieveInterswapTxsByStatus([]int{SecondPending}, 0, 0)
		if err != nil {
			log.Println("DBRetrieveTxsByStatus err:", err)
		}
		for _, txdata := range secondPendingTxs {
			err := processInterswapPendingFirstTx(txdata, config)
			if err != nil {
				log.Println("processPendingShieldTxs err:", txdata.TxID)
			}
		}
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
			SendSlackAlert(fmt.Sprintf("`InterswapID %v submit first swaptx failed 😵 `%v` \n", interswapTxID, err.Error()))
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
			return CheckStatusAndHandlePdexTx(&txData, config)
		}
	case PAppToPdex:
		{
			return CheckStatusAndHandlePappTx(&txData, config)
		}
	default:
		{

		}
	}
	return nil

}

type ResponseInfo struct {
	Coin      coin.PlainCoin
	CoinIndex *big.Int
	TxID      string
}

func findResponseUTXOs(privKey string, txReq string, tokenID string, metadataType int, config beCommon.Config) (*ResponseInfo, error) {
	// get UTXOs
	utxos, utxoIndices, err := incClient.GetUnspentOutputCoins(privKey, tokenID, 0)
	if err != nil {
		return nil, fmt.Errorf("Error get utxos %v", err)
	}
	coinPubKeys := []string{}
	for _, u := range utxos {
		coinPubKeys = append(coinPubKeys, base58.Base58Check{}.Encode(u.GetPublicKey().ToBytesS(), common.ZeroByte))
	}

	txResponses, err := CallGetTxsByCoinPubKeys(coinPubKeys, config)
	if err != nil {
		return nil, fmt.Errorf("Error get txs by coin pubkeys %v", err)
	}

	var foundIndex *int
	for i, tx := range txResponses {
		if tx.TxDetail.Metatype != metadataType {
			continue
		}

		if tx.TxDetail.Metadata == "" {
			continue
		}

		metadata, err := metadata.ParseMetadata([]byte(tx.TxDetail.Metadata))
		if err != nil {
			fmt.Printf("Error parse metadata %v\n", err)
			continue
		}

		switch metadataType {
		case metadataCommon.Pdexv3TradeResponseMeta:
			{
				md := metadata.(*metadataPdexv3.TradeResponse)
				if md.RequestTxID.String() != txReq {
					continue
				}
				foundIndex = &i
				break

			}
		case metadataCommon.BurnForCallResponseMeta:
			{
				md := metadata.(*metadataBridge.BurnForCallResponse)
				if md.UnshieldResponse.RequestedTxID.String() != txReq {
					continue
				}
				foundIndex = &i
				break

			}
		case metadataCommon.IssuingReshieldResponseMeta:
			{
				md := metadata.(*metadataBridge.IssuingReshieldResponse)
				if md.RequestedTxID.String() != txReq {
					continue
				}
				foundIndex = &i
				break
			}

		}
	}

	if foundIndex == nil {
		return nil, errors.New("Not found response utxos")
	}
	return &ResponseInfo{
		Coin:      utxos[*foundIndex],
		CoinIndex: utxoIndices[*foundIndex],
		TxID:      txResponses[*foundIndex].TxDetail.Hash,
	}, nil

}

func createTxRefundAndUpdateStatus(
	txData *beCommon.InterSwapTxData,
	amountRefund uint64, tokenRefund string,
	tokenUtxos []coin.PlainCoin, TokenUtxoIndices []uint64,
	updateStatus int,
) error {
	interswapTxID := txData.TxID
	refundTxID, err := createTxTokenWithInputCoins(config.ISIncPrivKey, txData.OTARefund, tokenRefund, amountRefund,
		tokenUtxos, TokenUtxoIndices, nil)
	if err != nil {
		log.Printf("InterswapID %v create tx refund error %v\n", interswapTxID, err)
		return fmt.Errorf("InterswapID %v create tx refund error %v\n", interswapTxID, err)
	}
	updateInfo := map[string]interface{}{
		"txidrefund": refundTxID,
		"status":     updateStatus,
		"statusstr":  StatusStr[updateStatus],
	}
	err = database.DBUpdateInterswapTxInfo(interswapTxID, updateInfo)
	if err != nil {
		log.Printf("InterswapID %v Update info %+v error %v\n", interswapTxID, updateInfo, err)
		return fmt.Errorf("InterswapID %v Update info %+v error %v\n", interswapTxID, updateInfo, err)
	}
	return nil
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

// func Validate
func CheckStatusAndHandlePdexTx(txData *beCommon.InterSwapTxData, config beCommon.Config) error {
	interswapTxID := txData.TxID

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

			responseInfo, err := findResponseUTXOs(config.ISIncPrivKey, interswapTxID, txData.MidToken, metadataCommon.Pdexv3TradeResponseMeta, config)
			if err != nil {
				log.Printf("InterswapID %v findResponseUTXOs error %v\n", interswapTxID, err)
				return err
			}
			responseAmt := responseInfo.Coin.GetValue()

			if responseAmt != pdexStatus.RespondAmounts[0] {
				msg := fmt.Sprintf("InterswapID %v response amount mismatched, expected %v, got %v\n", interswapTxID, pdexStatus.RespondAmounts[0], responseAmt)
				log.Printf(msg)
				SendSlackAlert(msg)
				return errors.New(msg)
			}
			if responseInfo.TxID != pdexStatus.RespondTxs[0] {
				msg := fmt.Sprintf("InterswapID %v response txid mismatched, expected %v, got %v\n", interswapTxID, pdexStatus.RespondTxs[0], responseInfo.TxID)
				log.Printf(msg)
				SendSlackAlert(msg)
				return errors.New(msg)
			}

			amtMidToken := responseAmt
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
				return createTxRefundAndUpdateStatus(txData, amtMidToken, txData.MidToken, []coin.PlainCoin{responseInfo.Coin}, []uint64{responseInfo.CoinIndex.Uint64()}, MidRefunding)
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
				return createTxRefundAndUpdateStatus(txData, amtMidToken, txData.MidToken, []coin.PlainCoin{responseInfo.Coin}, []uint64{responseInfo.CoinIndex.Uint64()}, MidRefunding)
			}
			if err != nil {
				return err
			}

			// get child token of unified token
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

			// create addon tx (papp)
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
			// TODO: 0xkraken
			// createTxTokenWithInputCoins(config.ISIncPrivKey, "", txData.MidToken)

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
	return nil
}

func CheckStatusAndHandlePappTx(txData *beCommon.InterSwapTxData, config beCommon.Config) error {
	interswapTxID := txData.TxID
	data, err := database.DBGetPappTxData(interswapTxID)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Printf("InterswapID %v DBGetPappTxData error %v\n", interswapTxID, err)
			return err
		}
	}

	// Step 1: check inchain tx
	if data.Status == beCommon.StatusSubmitFailed {
		// shard chain reject
		// update status SubmitFailed
		// do nothing
		err = database.DBUpdateInterswapTxStatus(interswapTxID, SubmitFailed, StatusStr[SubmitFailed], "")
		if err != nil {
			log.Printf("InterswapID %v Update status %+v error %v\n", interswapTxID, SubmitFailed, err)
			return fmt.Errorf("InterswapID %v Update status %+v error %v\n", interswapTxID, SubmitFailed, err)
		}
		return nil
	}

	if data.Status == beCommon.StatusRejected {
		// get tx response and corresponding utxo
		amtResponse := uint64(0)
		err := createTxRefundAndUpdateStatus(txData, amtResponse, txData.FromToken, []coin.PlainCoin{}, []uint64{}, FirstRefunding)
		if err != nil {
			return err
		}
		return nil
	}

	if data.Status == beCommon.StatusAccepted {
		// Step 2: check outchain tx
		externalTxStatus, err := database.DBRetrieveExternalTxByIncTxID(data.IncTx)
		if err != nil {
			log.Printf("InterswapID %v DBRetrieveExternalTxByIncTxID error %v\n", interswapTxID, err)
			return fmt.Errorf("InterswapID %v DBRetrieveExternalTxByIncTxID error %v\n", interswapTxID, err)
		}

		if externalTxStatus.Status == beCommon.StatusSubmitFailed {
			msg := fmt.Sprintf("InterswapID %v Tx out chain submit failed. Please check!\n", interswapTxID)
			SendSlackAlert(msg)
			return fmt.Errorf(msg)
		}

		if externalTxStatus.Status == beCommon.StatusAccepted {
			// unmarshal OtherInfo
			var outchainTxResult beCommon.ExternalTxSwapResult
			err = json.Unmarshal([]byte(externalTxStatus.OtherInfo), &outchainTxResult)
			if err != nil {
				log.Printf("InterswapID %v Unmarshal externalTxStatus.OtherInfo error %v\n", interswapTxID, err)
				return fmt.Errorf("InterswapID %v Unmarshal externalTxStatus.OtherInfo error %v\n", interswapTxID, err)
			}

			if outchainTxResult.IsFailed {
				msg := fmt.Sprintf("InterswapID %v Tx out chain status failed. Please check!\n", interswapTxID)
				SendSlackAlert(msg)
				return fmt.Errorf(msg)
			}

			if outchainTxResult.IsReverted {
				// swap failed
				// wait to redeposited, get tx repsonse and utxo create tx refund
				//
				amountRefund := uint64(0)

				err := createTxRefundAndUpdateStatus(txData, amountRefund, txData.MidToken, []coin.PlainCoin{}, []uint64{}, FirstRefunding)
				if err != nil {
					return err
				}
				return nil

			} else {
				// swap sucess
				// wait to redeposit
				if outchainTxResult.IsRedeposit {
					// create the add on tx
					// re-estimate addon tx
					buyAmount := ""
					reward := ""
					amtMidStr, err := addStrs(buyAmount, reward)
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
						err := createTxRefundAndUpdateStatus(txData, amtMidToken, txData.MidToken,
							[]coin.PlainCoin{}, []uint64{}, MidRefunding)
						if err != nil {
							return err
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
						err := createTxRefundAndUpdateStatus(txData, amtMidToken, txData.MidToken,
							[]coin.PlainCoin{}, []uint64{}, MidRefunding)
						if err != nil {
							return err
						}
						return nil
					}
					if err != nil {
						return err
					}

					// create pdex tx
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
			}
		}
	}
	return nil
}

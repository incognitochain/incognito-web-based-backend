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
	metadataBridge "github.com/incognitochain/go-incognito-sdk-v2/metadata/bridge"
	metadataCommon "github.com/incognitochain/go-incognito-sdk-v2/metadata/common"
	metadataPdexv3 "github.com/incognitochain/go-incognito-sdk-v2/metadata/pdexv3"
	beCommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"go.mongodb.org/mongo-driver/mongo"
)

var SkipTxs = []string{
	"82a30987e144aed3f3a3c56c0f1209477fe1c0444d8c38643989674ed24bf339",
	"b8d289152ca4e8819a2cc81cc5cff59ca999a71d80b34c2160df320936cbe669",
	"1cd9e638e1a10182348dae56ca974356c5ee2d876b7ba0fedcb44594a08c8711",
}

// watchInterswapPendingTx listen status of interswap and handle the next step
func watchInterswapPendingTx(config beCommon.Config) {
	log.Println("Starting Interswap Watcher")
	for {
		firstPendingTxs, err := database.DBRetrieveInterswapTxsByStatus([]int{FirstPending}, 0, 0)
		if err != nil {
			log.Println("DBRetrieveTxsByStatus err:", err)
		}
		log.Printf("Found `%v` pending first txs", len(firstPendingTxs))
		for _, txdata := range firstPendingTxs {
			err := processInterswapPendingFirstTx(txdata, config)
			if err != nil {
				log.Printf("processInterswapPendingFirstTx %v err: %v", txdata.TxID, err)
			}
		}

		secondPendingTxs, err := database.DBRetrieveInterswapTxsByStatus([]int{SecondPending}, 0, 0)
		if err != nil {
			log.Println("DBRetrieveTxsByStatus err:", err)
		}
		log.Printf("Found `%v` pending second txs", len(secondPendingTxs))
		for _, txdata := range secondPendingTxs {
			err := processInterswapPendingSecondTx(txdata, config)
			if err != nil {
				log.Println("processInterswapPendingSecondTx err:", txdata.TxID)
			}
		}

		refundingTxs, err := database.DBRetrieveInterswapTxsByStatus([]int{MidRefunding, FirstRefunding}, 0, 0)
		if err != nil {
			log.Println("DBRetrieveTxsByStatus err:", err)
		}
		log.Printf("Found `%v` refunding txs", len(refundingTxs))
		for _, txdata := range refundingTxs {
			err := processInterswapRefundingTx(txdata, config)
			if err != nil {
				log.Println("processInterswapRefundingTx err:", txdata.TxID)
			}
		}
		time.Sleep(10 * time.Second)
	}
}

// TODO: check Lam's BE and watcher use the same IncFullnode or not
func processInterswapPendingFirstTx(txData beCommon.InterSwapTxData, config beCommon.Config) error {
	interswapTxID := txData.TxID
	if IsExist(SkipTxs, interswapTxID) {
		return nil
	}

	// check tx by hash
	isConfirm, isMaxReCheck, _, updatedNumRecheck, err := CheckIncTxIsConfirmed(interswapTxID, txData.NumRecheck)
	database.DBUpdateInterswapTxNumRecheck(interswapTxID, updatedNumRecheck)
	if err != nil {
		if isMaxReCheck {
			SendSlackAlert(fmt.Sprintf("InterswapID %v First tx is rejected by chain. Update status Failed 😵 `%v` \n", interswapTxID, err.Error()))
			err = database.DBUpdateInterswapTxStatus(interswapTxID, SubmitFailed, StatusStr[SubmitFailed], err.Error())
			if err != nil {
				SendSlackAlert(fmt.Sprintf("`InterswapID %v update status `%v` error 😵 `%v` \n", interswapTxID, SubmitFailed, err.Error()))
				log.Printf("`InterswapID %v update status `%v` error 😵 `%v` \n", interswapTxID, SubmitFailed, err.Error())
				return err
			}
			SendSlackSwapInfo(interswapTxID, txData.UserAgent, "failed",
				txData.FromAmount, txData.FromToken, txData.FinalMinExpectedAmt, txData.ToToken, txData.FromAmount, txData.FromToken, config)
		}
		return err
	}
	if !isConfirm {
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
			return errors.New("Invalid path type")
		}
	}
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
		coinPubKey := base58.Base58Check{}.Encode(u.GetPublicKey().ToBytesS(), common.ZeroByte)
		log.Printf("findResponseUTXOs: coinPubKey %v \n", coinPubKey)
		coinPubKeys = append(coinPubKeys, coinPubKey)
	}

	txResponses, err := CallGetTxsByCoinPubKeys2(coinPubKeys, config, incClient)
	if err != nil {
		return nil, fmt.Errorf("Error get txs by coin pubkeys %v", err)
	}
	log.Printf("findResponseUTXOs: txResponses %v %v \n", len(txResponses), txResponses)

	foundCoinPubKey := ""
	foundTxID := ""
	for pubCoin, txs := range txResponses {
		for txID, tx := range txs {
			if tx.GetMetadata() == nil {
				continue
			}
			if tx.GetMetadataType() != metadataType {
				continue
			}

			metadata := tx.GetMetadata()

			switch metadataType {
			case metadataCommon.Pdexv3TradeResponseMeta:
				{
					md := metadata.(*metadataPdexv3.TradeResponse)
					if md.RequestTxID.String() != txReq {
						continue
					}
					foundCoinPubKey = pubCoin
					foundTxID = txID
					break

				}
			case metadataCommon.BurnForCallResponseMeta:
				{
					md := metadata.(*metadataBridge.BurnForCallResponse)
					if md.UnshieldResponse.RequestedTxID.String() != txReq {
						continue
					}
					foundCoinPubKey = pubCoin
					foundTxID = txID
					break

				}
			case metadataCommon.IssuingReshieldResponseMeta:
				{
					md := metadata.(*metadataBridge.IssuingReshieldResponse)
					if md.RequestedTxID.String() != txReq {
						continue
					}
					foundCoinPubKey = pubCoin
					foundTxID = txID
					break
				}
			}
		}
	}

	if foundCoinPubKey == "" || foundTxID == "" {
		return nil, errors.New("Not found response utxos")
	}

	var foundIndex *int
	for i, coin := range coinPubKeys {
		if coin == foundCoinPubKey {
			foundIndex = &i
			break
		}
	}
	if foundIndex == nil {
		return nil, errors.New("Invalid coin pub key and index")
	}
	return &ResponseInfo{
		Coin:      utxos[*foundIndex],
		CoinIndex: utxoIndices[*foundIndex],
		TxID:      foundTxID,
	}, nil

}

func createTxRefundAndUpdateStatus(
	privateKey string,
	txData *beCommon.InterSwapTxData,
	amountRefund uint64, tokenRefund string,
	tokenUtxos []coin.PlainCoin, tokenUtxoIndices []uint64,
	updateStatus int,
) error {
	interswapTxID := txData.TxID
	refundTxID, _, err := createTxTokenWithInputCoins(privateKey, txData.OTARefund, tokenRefund, amountRefund,
		tokenUtxos, tokenUtxoIndices, nil, true)
	if err != nil {
		log.Printf("InterswapID %v create tx refund error %v\n", interswapTxID, err)
		return fmt.Errorf("InterswapID %v create tx refund error %v\n", interswapTxID, err)
	}
	log.Printf("InterswapID %v Create refund txID %v\n", interswapTxID, refundTxID)
	updateInfo := map[string]interface{}{
		"txidrefund":     refundTxID,
		"amountresponse": amountRefund,
		"tokenresponse":  tokenRefund,
		"status":         updateStatus,
		"statusstr":      StatusStr[updateStatus],
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
	est2, err := CallEstimateSwap(params, config, BEEndpoint)
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
			if strings.ToLower(Remove0xPrefix(pApp.CallContract)) == strings.ToLower(Remove0xPrefix(expectedCallContract)) {
				pAppAddOn = &pApp
				break
			}
		}
		if pAppAddOn == nil || pAppAddOn.AmountOut == "" {
			log.Printf("InterswapID %v Not found trade path for add on tx\n", interswapTxID)
			errRes = fmt.Errorf("InterswapID %v Not found trade path for add on tx\n", interswapTxID)
			isMidRefund = true
			return
		}

		minAmountOut, err := convertToDecAmtUint64(pAppAddOn.AmountOut, params.ToToken, config)
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
			if pAppAddOn.Fee[0].Amount > expectedFeeAmount {
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

func CheckStatusAndHandlePdexTx(txData *beCommon.InterSwapTxData, config beCommon.Config) error {
	interswapTxID := txData.TxID
	shardID := fmt.Sprint(txData.ShardID)

	_, pdexStatus, err := CallGetPdexSwapTxStatus(interswapTxID, config)
	if err != nil {
		log.Printf("CallGetPdexSwapTxStatus TxID %v error %v ", interswapTxID, err)
		return err
	}

	if len(pdexStatus.RespondTxs) > 0 {
		if pdexStatus.Status == "accepted" {
			// parse tx response to get received UTXO
			if len(pdexStatus.RespondTxs) != 1 {
				log.Printf("InterswapID %v PDex response txs greater than one tx\n", interswapTxID)
				return err
			}

			responseInfo, err := findResponseUTXOs(config.ISIncPrivKeys[shardID], interswapTxID, txData.MidToken, metadataCommon.Pdexv3TradeResponseMeta, config)
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
				msg := fmt.Sprintf("InterswapID %v convert amount mid token to string error %v\n", interswapTxID, err)
				log.Printf(msg)
				SendSlackAlert(msg)
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
			log.Printf("InterswapID %v addonParamEst 2: %s\n", interswapTxID, string(p2Bytes))

			pAppAddOn, isMidRefund, err := callEstimateSwapAndValidation(addonParamEst, txData.FinalMinExpectedAmt, txData.PAppNetwork, txData.PAppContract, txData.MidToken, 0, interswapTxID)
			if isMidRefund {
				// refund: Estimation for addon tx is not valid
				return createTxRefundAndUpdateStatus(config.ISIncPrivKeys[shardID], txData, amtMidToken, txData.MidToken, []coin.PlainCoin{responseInfo.Coin}, []uint64{responseInfo.CoinIndex.Uint64()}, MidRefunding)
			}
			if err != nil {
				log.Printf("InterswapID %v Estimate addon tx error %v\n", interswapTxID, err)
				return err
			}

			addonFeeAmt := pAppAddOn.Fee[0].Amount
			if addonFeeAmt >= amtMidToken {
				// refund: Fee of addon tx is greater than swap amount
				log.Printf("InterswapID %v Addon tx swap fee is greater than swap amount", interswapTxID)
				return createTxRefundAndUpdateStatus(config.ISIncPrivKeys[shardID], txData, amtMidToken, txData.MidToken, []coin.PlainCoin{responseInfo.Coin}, []uint64{responseInfo.CoinIndex.Uint64()}, MidRefunding)
			}
			addonSwapAmt := amtMidToken - addonFeeAmt

			amtTmp, err := ConvertUint64ToWithoutDecStr(addonSwapAmt, txData.MidToken, config)
			if err != nil {
				log.Printf("InterswapID %v Convert amount from to string %+v error %v\n", interswapTxID, addonSwapAmt, err)
				return fmt.Errorf("InterswapID %v Convert amount from to string %+v error %v\n", interswapTxID, addonSwapAmt, err)
			}
			addonParamEst.Amount = amtTmp

			// estimate with final addon amount
			pAppAddOn, isMidRefund, err = callEstimateSwapAndValidation(addonParamEst, txData.FinalMinExpectedAmt, txData.PAppNetwork, txData.PAppContract, txData.MidToken, addonFeeAmt, interswapTxID)
			if isMidRefund {
				return createTxRefundAndUpdateStatus(config.ISIncPrivKeys[shardID], txData, amtMidToken, txData.MidToken, []coin.PlainCoin{responseInfo.Coin}, []uint64{responseInfo.CoinIndex.Uint64()}, MidRefunding)
			}
			if err != nil {
				log.Printf("InterswapID %v Estimate addon tx error %v\n", interswapTxID, err)
				return err
			}

			// get child token of MidToken (sellToken)
			burnTokenID := txData.MidToken
			networkID := uint8(beCommon.GetNetworkID(txData.PAppNetwork))

			childTokenIDStr, err := GetChildTokenUnified(burnTokenID, int(networkID), config)
			if err != nil {
				log.Printf("InterswapID %v Get child token network %v of midToken error %v\n", interswapTxID, txData.PAppNetwork, err)
				return err
			}
			childTokenID, err := common.Hash{}.NewHashFromStr(childTokenIDStr)
			if err != nil {
				log.Printf("InterswapID %v Invalid child token ID %v %v\n", interswapTxID, childTokenIDStr, err)
				return err
			}
			redepositAddress := new(coin.OTAReceiver)
			err = redepositAddress.FromString(txData.OTAToToken)
			if err != nil {
				// OTAToToken was verified in the submit step, so never meet this error
				log.Printf("InterswapID %v Invalid OTAToToken %v\n", interswapTxID, err)
				return fmt.Errorf("InterswapID %v Invalid OTAToToken %v\n", interswapTxID, err)
			}

			// get receiveToken (contractID of child ToToken)
			childToToken, err := GetChildTokenUnified(txData.ToToken, int(networkID), config)
			if err != nil {
				log.Printf("InterswapID %v Get child token of ToToken error %v\n", interswapTxID, err)
				return fmt.Errorf("InterswapID %v Get child token of ToToken error %v\n", interswapTxID, err)
			}

			childToTokenInfo, err := getTokenInfo(childToToken, config)
			if err != nil {
				log.Printf("InterswapID %v Get child token info of ToToken error %v\n", interswapTxID, err)
				return fmt.Errorf("InterswapID %v Get child token info of ToToken error %v\n", interswapTxID, err)
			}

			receiveTokenContract := Remove0xPrefix(childToTokenInfo.ContractID)
			withdrawAddr := txData.WithdrawAddress
			if withdrawAddr == "" {
				withdrawAddr = "0000000000000000000000000000000000000000"
			}

			// create addon tx (papp)
			data := metadataBridge.BurnForCallRequestData{
				BurningAmount:       addonSwapAmt,
				ExternalNetworkID:   networkID,
				IncTokenID:          *childTokenID,
				ExternalCalldata:    pAppAddOn.Calldata,
				ExternalCallAddress: Remove0xPrefix(txData.PAppContract),
				ReceiveToken:        receiveTokenContract,
				RedepositReceiver:   *redepositAddress, // user OTA
				WithdrawAddress:     txData.WithdrawAddress,
			}

			addOnTxID, txBytes, err := createTxBurnForCallWithInputCoins(config.ISIncPrivKeys[shardID], burnTokenID, data,
				[]string{pAppAddOn.FeeAddress}, []uint64{addonFeeAmt},
				[]coin.PlainCoin{responseInfo.Coin}, []uint64{responseInfo.CoinIndex.Uint64()}, false)
			if err != nil {
				msg := fmt.Sprintf("InterswapID %v. Please check @hiennguyen. Create addon tx papp error %v\n", interswapTxID, err)
				log.Printf(msg)
				SendSlackAlert(msg)
				return errors.New(msg)
			}

			// submit addon tx to BE
			_, err = CallSubmitPappSwapTx(string(txBytes), addOnTxID, txData.OTARefundFee, config, BEEndpoint)
			if err != nil {
				msg := fmt.Sprintf("InterswapID %v. Please check @hiennguyen. Submit addon tx papp error %v\n", interswapTxID, err)
				log.Printf(msg)
				SendSlackAlert(msg)
				return errors.New(msg)
			}

			// update db
			updateInfo := map[string]interface{}{
				"addon_txid": addOnTxID,
				"status":     SecondPending,
				"statusstr":  StatusStr[SecondPending],
			}
			err = database.DBUpdateInterswapTxInfo(interswapTxID, updateInfo)
			if err != nil {
				msg := fmt.Sprintf("InterswapID %v Update info %+v error %v\n", interswapTxID, updateInfo, err)
				log.Printf(msg)
				SendSlackAlert(msg)
				return fmt.Errorf(msg)
			}

			return nil

		} else if pdexStatus.Status == "refund" {
			err = SendSlackSwapInfo(interswapTxID, txData.UserAgent, "was refunded (first tx)",
				txData.FromAmount, txData.FromToken,
				txData.FinalMinExpectedAmt, txData.ToToken,
				txData.FromAmount, txData.FromToken, config)
			if err != nil {
				log.Printf("InterswapID %v send slack swap info error %v", interswapTxID, err)
			}
			err = database.DBUpdateInterswapTxStatus(interswapTxID, FirstRefunded, StatusStr[FirstRefunded], "")
			if err != nil {
				msg := fmt.Sprintf("InterswapID %v Update status %+v error %v\n", interswapTxID, FirstRefunded, err)
				log.Printf(msg)
				SendSlackAlert(msg)
				return fmt.Errorf(msg)
			}
			return nil
		}
	}
	return nil
}

func CheckStatusAndHandlePappTx(txData *beCommon.InterSwapTxData, config beCommon.Config) error {
	interswapTxID := txData.TxID
	shardID := fmt.Sprint(txData.ShardID)
	data, err := database.DBGetPappTxData(interswapTxID)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Printf("InterswapID %v DBGetPappTxData error %v\n", interswapTxID, err)
			return err
		}
	}

	// Step 1: check inchain tx
	// NOTE: Because tx was confirmed in block, the status never is StatusSubmitFailed

	if data.Status == beCommon.StatusRejected {
		// get tx response and corresponding utxo
		responseInfo, err := findResponseUTXOs(config.ISIncPrivKeys[shardID], interswapTxID, txData.FromToken, metadataCommon.BurnForCallResponseMeta, config)
		if err != nil {
			msg := fmt.Sprintf("InterswapID %v find response UTXOs to create refund tx error %v", interswapTxID, err)
			log.Printf(msg)
			return errors.New(msg)
		}
		amtResponse := responseInfo.Coin.GetValue()
		return createTxRefundAndUpdateStatus(config.ISIncPrivKeys[shardID], txData, amtResponse, txData.FromToken, []coin.PlainCoin{responseInfo.Coin}, []uint64{responseInfo.CoinIndex.Uint64()}, FirstRefunding)
	}

	if data.Status == beCommon.StatusAccepted {
		// Step 2: check outchain tx
		externalTxStatus, err := database.DBRetrieveExternalTxByIncTxID(data.IncTx)
		if err != nil || externalTxStatus == nil {
			log.Printf("InterswapID %v DBRetrieveExternalTxByIncTxID error %v\n", interswapTxID, err)
			return fmt.Errorf("InterswapID %v DBRetrieveExternalTxByIncTxID error %v\n", interswapTxID, err)
		}

		if externalTxStatus.Status == beCommon.StatusSubmitFailed {
			msg := fmt.Sprintf("InterswapID %v Tx out chain submit failed. Please check @lam!\n", interswapTxID)
			log.Printf(msg)
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
				msg := fmt.Sprintf("InterswapID %v Tx out chain status failed. Please check @lam!\n", interswapTxID)
				log.Printf(msg)
				SendSlackAlert(msg)
				return fmt.Errorf(msg)
			}

			if outchainTxResult.IsReverted {
				// swap failed
				// wait to redeposited, get tx repsonse and utxo create tx refund
				if !outchainTxResult.IsRedeposit {
					msg := fmt.Sprintf("InterswapID %v Tx out chain is not redeposit. Please check!\n", interswapTxID)
					SendSlackAlert(msg)
					return fmt.Errorf(msg)
				}

				// get txID of redeposit
				redepositInfo, err := database.DBGetShieldTxByExternalTx(externalTxStatus.Txhash, beCommon.GetNetworkID(externalTxStatus.Network))
				if err != nil || redepositInfo.IncTx == "" {
					log.Printf("InterswapID %v DBGetShieldTxByExternalTx %v err %v", interswapTxID, externalTxStatus.Txhash, err)
					return err
				}
				txIDReqRedeposit := redepositInfo.IncTx
				tokenRefund := txData.FromToken
				responseInfo, err := findResponseUTXOs(config.ISIncPrivKeys[shardID], txIDReqRedeposit, tokenRefund, metadataCommon.IssuingReshieldResponseMeta, config)
				if err != nil {
					msg := fmt.Sprintf("InterswapID %v find response UTXOs to create refund tx error %v", interswapTxID, err)
					log.Printf(msg)
					return errors.New(msg)
				}
				amountRefund := responseInfo.Coin.GetValue()
				return createTxRefundAndUpdateStatus(config.ISIncPrivKeys[shardID], txData, amountRefund, tokenRefund, []coin.PlainCoin{responseInfo.Coin}, []uint64{responseInfo.CoinIndex.Uint64()}, FirstRefunding)

			} else {
				// swap sucess
				// wait to redeposit
				if outchainTxResult.IsRedeposit {
					// get txID of redeposit
					redepositInfo, err := database.DBGetShieldTxByExternalTx(externalTxStatus.Txhash, beCommon.GetNetworkID(externalTxStatus.Network))
					if err != nil || redepositInfo.IncTx == "" {
						log.Printf("InterswapID %v DBGetShieldTxByExternalTx %v err %v", interswapTxID, externalTxStatus.Txhash, err)
						return err
					}
					txIDReqRedeposit := redepositInfo.IncTx
					tokenResponse := txData.MidToken
					responseInfo, err := findResponseUTXOs(config.ISIncPrivKeys[shardID], txIDReqRedeposit, tokenResponse, metadataCommon.IssuingReshieldResponseMeta, config)
					if err != nil {
						msg := fmt.Sprintf("InterswapID %v find response UTXOs to create addon tx error %v", interswapTxID, err)
						log.Printf(msg)
						return errors.New(msg)
					}
					amountResponse := responseInfo.Coin.GetValue()

					// create the add on tx
					// re-estimate addon tx
					amtMidToken := amountResponse
					amtMidStr, err := ConvertUint64ToWithoutDecStr(amtMidToken, tokenResponse, config)
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
					fmt.Printf("InterswapID %v Estimate param for addon tx: %s\n", interswapTxID, string(p2Bytes))

					pDexAddOn, isMidRefund, err := callEstimateSwapAndValidation(addonParamEst, txData.FinalMinExpectedAmt, IncNetworkStr, "", txData.MidToken, 0, interswapTxID)
					if isMidRefund {
						return createTxRefundAndUpdateStatus(config.ISIncPrivKeys[shardID], txData, amountResponse, tokenResponse,
							[]coin.PlainCoin{responseInfo.Coin}, []uint64{responseInfo.CoinIndex.Uint64()}, MidRefunding)
					}
					if err != nil {
						return err
					}

					addonFeeAmt := pDexAddOn.Fee[0].Amount
					if addonFeeAmt >= amtMidToken {
						// refund: Fee of addon tx is greater than swap amount
						log.Printf("InterswapID %v Addon tx swap fee is greater than swap amount", interswapTxID)
						return createTxRefundAndUpdateStatus(config.ISIncPrivKeys[shardID], txData, amountResponse, tokenResponse,
							[]coin.PlainCoin{responseInfo.Coin}, []uint64{responseInfo.CoinIndex.Uint64()}, MidRefunding)
					}
					addonSwapAmt := amtMidToken - addonFeeAmt
					amtTmp, err := ConvertUint64ToWithoutDecStr(addonSwapAmt, txData.MidToken, config)
					if err != nil {
						log.Printf("InterswapID %v Convert amount from to string %+v error %v\n", interswapTxID, addonSwapAmt, err)
						return fmt.Errorf("InterswapID %v Convert amount from to string %+v error %v\n", interswapTxID, addonSwapAmt, err)
					}
					addonParamEst.Amount = amtTmp

					// estimate with final addon amount
					pDexAddOn, isMidRefund, err = callEstimateSwapAndValidation(addonParamEst, txData.FinalMinExpectedAmt, IncNetworkStr, "", txData.MidToken, addonFeeAmt, interswapTxID)
					if isMidRefund {
						return createTxRefundAndUpdateStatus(config.ISIncPrivKeys[shardID], txData, amtMidToken, txData.MidToken,
							[]coin.PlainCoin{}, []uint64{}, MidRefunding)
					}
					if err != nil {
						return err
					}

					// create pdex tx
					// user's ota
					otaReceiver := map[string]string{
						txData.MidToken: txData.OTAFromToken,
						txData.ToToken:  txData.OTAToToken,
					}
					addOnTxID, err := incClient.CreateAndSendPdexv3TradeWithOTAReceiversTransaction(
						config.ISIncPrivKeys[shardID], pDexAddOn.PoolPairs, txData.MidToken, txData.ToToken,
						addonSwapAmt, txData.FinalMinExpectedAmt, addonFeeAmt, false, otaReceiver)
					if err != nil {
						msg := fmt.Sprintf("InterswapID %v. Please check @hiennguyen. Create addon tx pdex error %v\n", interswapTxID, err)
						log.Printf(msg)
						SendSlackAlert(msg)
						return errors.New(msg)
					}
					log.Printf("InterswapID %v Create addon pdex txID %v\n", interswapTxID, addOnTxID)

					// update db
					updateInfo := map[string]interface{}{
						"addon_txid": addOnTxID,
						"status":     SecondPending,
						"statusstr":  StatusStr[SecondPending],
					}
					err = database.DBUpdateInterswapTxInfo(interswapTxID, updateInfo)
					if err != nil {
						msg := fmt.Sprintf("InterswapID %v Update info %+v error %v\n", interswapTxID, updateInfo, err)
						log.Printf(msg)
						SendSlackAlert(msg)
						return fmt.Errorf(msg)
					}

					return nil
				}
			}
		}
	}
	return nil
}

var IsUpdate = false

func processInterswapPendingSecondTx(txData beCommon.InterSwapTxData, config beCommon.Config) error {
	interswapTxID := txData.TxID
	addOnTxID := txData.AddOnTxID
	fmt.Printf("InterswapID %v Start processing addOnTxID %v\n", interswapTxID, addOnTxID)
	if IsExist(SkipTxs, interswapTxID) {
		return nil
	}

	// check tx by hash
	isConfirm, isMaxReCheck, shouldRefund, updatedNumRecheck, err := CheckIncTxIsConfirmed(addOnTxID, txData.NumRecheck)
	database.DBUpdateInterswapTxNumRecheck(interswapTxID, updatedNumRecheck)
	if err != nil {
		if isMaxReCheck && shouldRefund {
			SendSlackAlert(fmt.Sprintf("InterswapID %v AddonTxID %v Second tx is rejected by chain. Need to retry 😵 `%v` \n", interswapTxID, addOnTxID, err.Error()))

			// createTxRefundAndUpdateStatus()
			status := FirstPending
			err = database.DBUpdateInterswapTxStatus(interswapTxID, status, StatusStr[status], err.Error())
			if err != nil {
				SendSlackAlert(fmt.Sprintf("`InterswapID %v update status `%v` error 😵 `%v` \n", interswapTxID, status, err.Error()))
				log.Printf("`InterswapID %v update status `%v` error 😵 `%v` \n", interswapTxID, status, err.Error())
				return err
			}
		}
		return err
	}
	if !isConfirm {
		return nil
	}

	switch txData.PathType {
	case PAppToPdex:
		{
			return CheckStatusAndHandlePdexTxSecond(&txData, config)
		}
	case PdexToPApp:
		{
			return CheckStatusAndHandlePappTxSecond(&txData, config)
		}
	default:
		{
			return errors.New("Invalid path type")
		}
	}
}

// CheckStatusAndHandlePdexTxSecond
func CheckStatusAndHandlePdexTxSecond(txData *beCommon.InterSwapTxData, config beCommon.Config) error {
	interswapTxID := txData.TxID
	addOnTxID := txData.AddOnTxID
	// shardID := fmt.Sprint(txData.ShardID)

	_, pdexStatus, err := CallGetPdexSwapTxStatus(addOnTxID, config)
	if err != nil {
		msg := fmt.Sprintf("InterswapID %v CallGetPdexSwapTxStatus addOnTxID %v error %v\n", interswapTxID, addOnTxID, err)
		log.Printf(msg)
		SendSlackAlert(msg)
		return fmt.Errorf(msg)
	}
	fmt.Printf("InterswapID %v processing addOnTxID %v pdexstatus %v", interswapTxID, addOnTxID, pdexStatus)

	if len(pdexStatus.RespondTxs) > 0 {
		if pdexStatus.Status == "accepted" {
			// parse tx response to get received UTXO
			if len(pdexStatus.RespondTxs) != 1 {
				log.Printf("InterswapID %v CallGetPdexSwapTxStatus error %v", interswapTxID, err)
				return err
			}

			err := SendSlackSwapInfo(interswapTxID, txData.UserAgent, "was success",
				txData.FromAmount, txData.FromToken,
				txData.FinalMinExpectedAmt, txData.ToToken,
				pdexStatus.RespondAmounts[0], pdexStatus.RespondTokens[0], config)
			if err != nil {
				log.Printf("InterswapID %v send slack swap info error %v", interswapTxID, err)
			}

			// update db
			updateInfo := map[string]interface{}{
				"txidresponse":   pdexStatus.RespondTxs[0],
				"amountresponse": pdexStatus.RespondAmounts[0],
				"tokenresponse":  pdexStatus.RespondTokens[0],
				"status":         SecondSuccess,
				"statusstr":      StatusStr[SecondSuccess],
			}
			err = database.DBUpdateInterswapTxInfo(interswapTxID, updateInfo)
			if err != nil {
				msg := fmt.Sprintf("InterswapID %v Update info %+v error %v\n", interswapTxID, updateInfo, err)
				log.Printf(msg)
				SendSlackAlert(msg)
				return fmt.Errorf(msg)
			}
			return nil

		} else if pdexStatus.Status == "refund" {
			// parse tx response to get received UTXO
			if len(pdexStatus.RespondTxs) != 1 {
				log.Printf("InterswapID %v CallGetPdexSwapTxStatus error %v", interswapTxID, err)
				return err
			}

			err := SendSlackSwapInfo(interswapTxID, txData.UserAgent, "was refunded (second tx)",
				txData.FromAmount, txData.FromToken,
				txData.FinalMinExpectedAmt, txData.ToToken,
				pdexStatus.RespondAmounts[0], pdexStatus.RespondTokens[0], config)
			if err != nil {
				log.Printf("InterswapID %v send slack swap info error %v", interswapTxID, err)
			}

			// update db
			updateInfo := map[string]interface{}{
				"txidresponse":   pdexStatus.RespondTxs[0],
				"amountresponse": pdexStatus.RespondAmounts[0],
				"tokenresponse":  pdexStatus.RespondTokens[0],
				"status":         SecondRefunded,
				"statusstr":      StatusStr[SecondRefunded],
			}
			err = database.DBUpdateInterswapTxInfo(interswapTxID, updateInfo)
			if err != nil {
				log.Printf("InterswapID %v Update info %+v error %v\n", interswapTxID, updateInfo, err)
				return fmt.Errorf("InterswapID %v Update info %+v error %v\n", interswapTxID, updateInfo, err)
			}
			return nil
		}
	}
	return nil
}

func CheckStatusAndHandlePappTxSecond(txData *beCommon.InterSwapTxData, config beCommon.Config) error {
	interswapTxID := txData.TxID
	addOnTxID := txData.AddOnTxID
	data, err := database.DBGetPappTxData(addOnTxID)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			msg := fmt.Sprintf("InterswapID %v DBGetPappTxData addOnTxID %v error %v\n", interswapTxID, addOnTxID, err)
			log.Printf(msg)
			SendSlackAlert(msg)
			return fmt.Errorf(msg)
		}
	}

	// Step 1: check inchain tx
	// NOTE: Because tx was confirmed in block, the status never is StatusSubmitFailed

	if data.Status == beCommon.StatusRejected {
		err := SendSlackSwapInfo(interswapTxID, txData.UserAgent, "was refunded (second tx)",
			txData.FromAmount, txData.FromToken,
			txData.FinalMinExpectedAmt, txData.ToToken,
			data.BurntAmount, data.BurntToken, config)
		if err != nil {
			log.Printf("InterswapID %v send slack swap info error %v", interswapTxID, err)
		}

		// update db
		updateInfo := map[string]interface{}{
			"txidresponse":   "", // NOTE: currently, cannot get tx response from tx request
			"amountresponse": data.BurntAmount,
			"tokenresponse":  data.BurntToken,
			"status":         SecondRefunded,
			"statusstr":      StatusStr[SecondRefunded],
		}
		err = database.DBUpdateInterswapTxInfo(interswapTxID, updateInfo)
		if err != nil {
			log.Printf("InterswapID %v Update info %+v error %v\n", interswapTxID, updateInfo, err)
			return fmt.Errorf("InterswapID %v Update info %+v error %v\n", interswapTxID, updateInfo, err)
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
				if !outchainTxResult.IsRedeposit {
					msg := fmt.Sprintf("InterswapID %v Tx out chain is not redeposit. Please check!\n", interswapTxID)
					SendSlackAlert(msg)
					return fmt.Errorf(msg)
				}

				// get txID of redeposit
				redepositInfo, err := database.DBGetShieldTxByExternalTx(externalTxStatus.Txhash, beCommon.GetNetworkID(externalTxStatus.Network))
				if err != nil || redepositInfo.IncTx == "" {
					log.Printf("InterswapID %v AddOnTxID %v DBGetShieldTxByExternalTx err %v", interswapTxID, addOnTxID, err)
					return err
				}
				txIDReqRedeposit := redepositInfo.IncTx

				err = SendSlackSwapInfo(interswapTxID, txData.UserAgent, "was refunded (second tx)",
					txData.FromAmount, txData.FromToken,
					txData.FinalMinExpectedAmt, txData.ToToken,
					data.BurntAmount, data.BurntToken, config)
				if err != nil {
					log.Printf("InterswapID %v send slack swap info error %v", interswapTxID, err)
				}

				// update database
				updateInfo := map[string]interface{}{
					"txidresponse":   txIDReqRedeposit, // NOTE: currently, cannot get tx response from tx request
					"amountresponse": data.BurntAmount,
					"tokenresponse":  data.BurntToken,
					"status":         SecondRefunded,
					"statusstr":      StatusStr[SecondRefunded],
				}
				err = database.DBUpdateInterswapTxInfo(interswapTxID, updateInfo)
				if err != nil {
					log.Printf("InterswapID %v Update info %+v error %v\n", interswapTxID, updateInfo, err)
					return fmt.Errorf("InterswapID %v Update info %+v error %v\n", interswapTxID, updateInfo, err)
				}
				return nil
			} else {
				// swap sucess
				// wait to redeposit
				if outchainTxResult.IsRedeposit {
					// get txID of redeposit
					redepositInfo, err := database.DBGetShieldTxByExternalTx(externalTxStatus.Txhash, beCommon.GetNetworkID(externalTxStatus.Network))
					if err != nil || redepositInfo.IncTx == "" {
						log.Printf("InterswapID %v DBGetShieldTxByExternalTx err %v", interswapTxID, err)
						return err
					}
					txIDReqRedeposit := redepositInfo.IncTx

					// update database
					amtResponse, err := convertAmtExtDecToAmtPDec(outchainTxResult.Amount, redepositInfo.TokenID, config)
					if err != nil {
						log.Printf("InterswapID %v Calculate the final response amount error %v\n", interswapTxID, err)
						return fmt.Errorf("InterswapID %v Calculate the final response amount error %v\n", interswapTxID, err)
					}

					err = SendSlackSwapInfo(interswapTxID, txData.UserAgent, "was success",
						txData.FromAmount, txData.FromToken,
						txData.FinalMinExpectedAmt, txData.ToToken,
						amtResponse, redepositInfo.UTokenID, config)
					if err != nil {
						log.Printf("InterswapID %v send slack swap info error %v", interswapTxID, err)
					}

					updateInfo := map[string]interface{}{
						"txidresponse":   txIDReqRedeposit, // NOTE: currently, cannot get tx response from tx request
						"amountresponse": amtResponse,
						"tokenresponse":  redepositInfo.UTokenID,
						"status":         SecondSuccess,
						"statusstr":      StatusStr[SecondSuccess],
					}
					err = database.DBUpdateInterswapTxInfo(interswapTxID, updateInfo)
					if err != nil {
						log.Printf("InterswapID %v Update info %+v error %v\n", interswapTxID, updateInfo, err)
						return fmt.Errorf("InterswapID %v Update info %+v error %v\n", interswapTxID, updateInfo, err)
					}
					return nil
				} else {
					// update database
					amtResponse, err := convertAmtExtDecToAmtPDec(outchainTxResult.Amount, txData.ToToken, config)
					if err != nil {
						log.Printf("InterswapID %v Calculate the final response amount error %v\n", interswapTxID, err)
						return fmt.Errorf("InterswapID %v Calculate the final response amount error %v\n", interswapTxID, err)
					}
					err = SendSlackSwapInfo(interswapTxID, txData.UserAgent, "was success",
						txData.FromAmount, txData.FromToken,
						txData.FinalMinExpectedAmt, txData.ToToken,
						amtResponse, txData.ToToken, config)
					if err != nil {
						log.Printf("InterswapID %v send slack swap info error %v", interswapTxID, err)
					}
					updateInfo := map[string]interface{}{
						"txidoutchain":   externalTxStatus.Txhash,
						"amountresponse": amtResponse,
						"tokenresponse":  txData.ToToken,
						"status":         SecondSuccess,
						"statusstr":      StatusStr[SecondSuccess],
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

func processInterswapRefundingTx(txData beCommon.InterSwapTxData, config beCommon.Config) error {
	interswapTxID := txData.TxID
	// addOnTxID := txData.AddOnTxID
	refundTxID := txData.TxIDRefund
	fmt.Printf("InterswapID %v Start processing refundTxID %v\n", interswapTxID, refundTxID)

	// check tx by hash
	txDetail, err := incClient.GetTxDetail(refundTxID)
	if err != nil {
		if strings.Contains(err.Error(), "RPC returns an error:") {
			// TODO: 0xkraken recheck tx detail 3 times
			// if still error, retry to create refund tx
			SendSlackAlert(fmt.Sprintf("InterswapID %v RefundTxID %v check tx is in block error 😵 `%v` \n", interswapTxID, refundTxID, err.Error()))
		}
		return err
	}
	if txDetail.IsInBlock {
		curStatus := txData.Status
		newStatus := MidRefunded
		statusMsg := "was refunded (from mid)"
		if curStatus == FirstRefunding {
			newStatus = FirstRefunded
			statusMsg = "was refunded (first tx)"
		}

		err := database.DBUpdateInterswapTxStatus(interswapTxID, newStatus, StatusStr[newStatus], "")
		err2 := SendSlackSwapInfo(interswapTxID, txData.UserAgent, statusMsg,
			txData.FromAmount, txData.FromToken,
			txData.FinalMinExpectedAmt, txData.ToToken,
			txData.AmountResponse, txData.TokenResponse, config)
		if err2 != nil {
			log.Printf("InterswapID %v SendSlackSwapInfo err %v\n", interswapTxID, err)
		}
		if err != nil {
			msg := fmt.Sprintf("InterswapID %v Update status %+v error %v\n", interswapTxID, newStatus, err)
			log.Printf(msg)
			SendSlackAlert(fmt.Sprintf(msg))
			return errors.New(msg)
		}
	}

	return nil
}

func CheckIncTxIsConfirmed(txID string, numRecheck uint) (isConfirm bool, isMaxRecheck bool, shouldRefund bool, numRecheckRes uint, err error) {
	isConfirm = false
	isMaxRecheck = false
	shouldRefund = false
	numRecheckRes = numRecheck
	// check tx by hash
	txDetail, err := incClient.GetTxDetail(txID)
	if err != nil {
		if strings.Contains(err.Error(), "RPC returns an error:") {
			numRecheckRes++
			if numRecheckRes >= MaxNumRecheck {
				isMaxRecheck = true
				shouldRefund = true
				numRecheckRes = 0 // reset num recheck
				// SendSlackAlert(fmt.Sprintf("InterswapID %v check tx is in block error 😵 `%v` \n", txID, err.Error()))
			}
			// TODO: 0xkraken recheck tx detail 3 times
			// if still error, retry to create refund tx
		}
		return
	}

	isConfirm = txDetail.IsInBlock
	numRecheckRes = 0 // reset num recheck
	return
}

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/incognitochain/go-incognito-sdk-v2/coin"
	"github.com/incognitochain/go-incognito-sdk-v2/common"
	"github.com/incognitochain/go-incognito-sdk-v2/common/base58"
	"github.com/incognitochain/go-incognito-sdk-v2/crypto"
	"github.com/incognitochain/go-incognito-sdk-v2/metadata"
	"github.com/incognitochain/go-incognito-sdk-v2/metadata/bridge"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/submitproof"
)

func APIGetUnshieldTxStatus(c *gin.Context) {
	var req SubmitTxListRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	result := make(map[string]interface{})

	var wg sync.WaitGroup
	var lock sync.Mutex
	for _, txHash := range req.TxList {
		wg.Add(1)
		go func(txh string) {
			statusResult := checkUnshieldTxStatus(txh)
			lock.Lock()
			if len(statusResult) == 0 {
				statusResult["error"] = "tx not found"
				result[txh] = statusResult
			} else {
				result[txh] = statusResult
			}
			lock.Unlock()
			wg.Done()
		}(txHash)
	}
	wg.Wait()
	c.JSON(200, gin.H{"Result": result})
}

func APIUnshieldFeeEstimate(c *gin.Context) {
	network := c.Query("network")
	tokenid := c.Query("tokenid")
	amount := c.Query("amount")
	amountUint, err := strconv.ParseUint(amount, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	networkFees, err := database.DBRetrieveFeeTable()
	if err != nil {
		fmt.Println("DBRetrieveFeeTable", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	burnTokenInfo, err := getTokenInfo(tokenid)
	if err != nil {
		fmt.Println("getTokenInfo", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("not supported token")})
		return
	}
	spTkList := getSupportedTokenList()
	feeAmount, err := estimateUnshieldFee(amountUint, burnTokenInfo, network, networkFees, spTkList)
	if err != nil {
		fmt.Println("estimateUnshieldFee", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"Result": feeAmount})
}

func APISubmitUnshieldTxNew(c *gin.Context) {
	userAgent := c.Request.UserAgent()
	var req SubmitSwapTxRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"Error": "ShouldBindJSON " + err.Error()})
		return
	}

	rawTxBytes, _, err := base58.Base58Check{}.Decode(req.TxRaw)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"Error": errors.New("invalid txhash").Error()})
		return
	}

	mdRaw, isPRVTx, outCoins, txHash, err := extractDataFromRawTx(rawTxBytes)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"Error": "extractDataFromRawTx " + err.Error()})
		return
	}
	var mdUnified *bridge.UnshieldRequest
	var md *bridge.BurningRequest

	md, ok := mdRaw.(*bridge.BurningRequest)
	if !ok {
		mdUnified, ok = mdRaw.(*bridge.UnshieldRequest)
		if !ok {
			var md2 bridge.BurningRequest
			mdRawJson, _ := json.Marshal(mdRaw)
			err = json.Unmarshal(mdRawJson, &md2)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"Error": "invalid metadata type"})
				return
			}
			md = &md2
		}
	}
	var burnTokenInfo *wcommon.TokenInfo
	var unshieldToken *wcommon.TokenInfo
	var burntAmount uint64
	isUnifiedToken := false
	networkList := []string{}
	tokenID := ""
	uTokenID := ""
	if md == nil {
		//unshield unified
		burnTokenInfo, err = getTokenInfo(mdUnified.UnifiedTokenID.String())
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"Error": errors.New("not supported token").Error()})
			return
		}
		burntAmount = mdUnified.Data[0].BurningAmount

		unshieldToken, err = getTokenInfo(mdUnified.Data[0].IncTokenID.String())
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"Error": errors.New("not supported token").Error()})
			return
		}

		tokenID = unshieldToken.TokenID
		uTokenID = burnTokenInfo.TokenID

		isUnifiedToken = true
	} else {
		burnTokenInfo, err = getTokenInfo(md.TokenID.String())
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"Error": errors.New("not supported token").Error()})
			return
		}
		tokenID = burnTokenInfo.TokenID
		uTokenID = burnTokenInfo.TokenID
		burntAmount = md.BurningAmount

	}

	valid, externalAddr, network, feeToken, feeAmount, pfeeAmount, feeDiff, err := checkValidUnshield(md, mdUnified, burnTokenInfo, unshieldToken, outCoins)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"Error": "invalid tx err: " + err.Error()})
		return
	}
	networkList = append(networkList, network)

	if valid {
		status, err := submitproof.SubmitUnshieldTx(txHash, []byte(req.TxRaw), isPRVTx, feeToken, feeAmount, pfeeAmount, tokenID, uTokenID, burntAmount, isUnifiedToken, externalAddr, networkList, req.FeeRefundOTA, req.FeeRefundAddress, userAgent)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"Error": "SubmitUnshieldTx " + err.Error()})
			return
		}
		c.JSON(200, gin.H{"Result": map[string]interface{}{"inc_request_tx_status": status}, "feeDiff": feeDiff})
		return
	}
}

func estimateUnshieldFee(amount uint64, burnTokenInfo *wcommon.TokenInfo, network string, networkFees *wcommon.ExternalNetworksFeeData, spTkList []PappSupportedTokenData) (*UnshieldNetworkFee, error) {
	feeTokenWhiteList, err := retrieveFeeTokenWhiteList()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	isFeeWhitelist := false
	if _, ok := feeTokenWhiteList[burnTokenInfo.TokenID]; ok {
		isFeeWhitelist = true
	}
	networkID := wcommon.GetNetworkID(network)

	if _, ok := networkFees.GasPrice[network]; !ok {
		return nil, errors.New("network gasPrice not found")
	}
	gasPrice := networkFees.GasPrice[network]

	nativeCurrentType := wcommon.GetNativeNetworkCurrencyType(network)
	nativeToken, err := getNativeTokenData(spTkList, nativeCurrentType)
	if err != nil {
		return nil, err
	}

	isUnifiedNativeToken := false

	if burnTokenInfo.CurrencyType == nativeCurrentType {
		isUnifiedNativeToken = true
	}
	if burnTokenInfo.CurrencyType == wcommon.UnifiedCurrencyType {
		for _, v := range burnTokenInfo.ListUnifiedToken {
			if v.CurrencyType == nativeCurrentType {
				isUnifiedNativeToken = true
				break
			}
		}
	}
	gasFee := (UNSHIELD_GAS_LIMIT * gasPrice)
	amountInBig0 := new(big.Float).SetUint64(amount)

	additionalTokenInFee := amountInBig0.Mul(amountInBig0, new(big.Float).SetFloat64(0.003))
	additionalTokenInFee = additionalTokenInFee.Mul(additionalTokenInFee, new(big.Float).SetFloat64(math.Pow10(-burnTokenInfo.PDecimals)))
	fees := getFee(isFeeWhitelist, isUnifiedNativeToken, nativeToken, new(big.Float).SetInt64(1), gasFee, burnTokenInfo.TokenID, burnTokenInfo, &PappSupportedTokenData{
		CurrencyType: burnTokenInfo.CurrencyType,
	}, new(big.Float).SetInt64(1), additionalTokenInFee)
	if len(fees) == 0 {
		return nil, errors.New("can't get fee")
	}
	burntAmount := uint64(0)
	protocolFee := uint64(0)
	if burnTokenInfo.CurrencyType == wcommon.UnifiedCurrencyType {
		var tokenID string

		for _, childToken := range burnTokenInfo.ListUnifiedToken {
			childNetID, err := wcommon.GetNetworkIDFromCurrencyType(childToken.CurrencyType)
			if err != nil {
				return nil, err
			}
			if childNetID == networkID {
				tokenID = childToken.TokenID
				break
			}
		}
		burntAmount, protocolFee, err = getUnifiedUnshieldFee(tokenID, burnTokenInfo.TokenID, amount)
		if err != nil {
			return nil, err
		}
	} else {
		burntAmount = amount
	}

	feeAddress := ""
	feeAddressShardID := byte(0)
	if incFeeKeySet != nil {
		feeAddress, err = incFeeKeySet.GetPaymentAddress()
		if err != nil {
			return nil, err
		}
		_, feeAddressShardID = common.GetShardIDsFromPublicKey(incFeeKeySet.KeySet.PaymentAddress.Pk)
	}
	result := UnshieldNetworkFee{
		FeeAddress:        feeAddress,
		FeeAddressShardID: int(feeAddressShardID),
		ExpectedReceive:   amount,
		BurntAmount:       burntAmount,
		TokenID:           fees[0].TokenID,
		Amount:            fees[0].Amount,
		PrivacyFee:        fees[0].PrivacyFee,
		ProtocolFee:       protocolFee,
		FeeInUSD:          fees[0].FeeInUSD,
	}

	return &result, nil
}

func checkValidUnshield(md *bridge.BurningRequest, mdUnified *bridge.UnshieldRequest, burnTokenInfo, unshieldToken *wcommon.TokenInfo, outCoins []coin.Coin) (bool, string, string, string, uint64, uint64, int64, error) {
	var feeAmount uint64
	var pfeeAmount uint64
	var feeToken string

	var requireFee uint64
	var requireFeeToken string
	var externalAddress string

	var burnAmount uint64

	var result bool
	feeDiff := int64(-1)
	callNetwork := ""
	// networkInfo, err := getBridgeNetworkInfos()
	// if err != nil {
	// 	return result, callNetwork, feeToken, feeAmount, pfeeAmount, feeDiff, err
	// }
	networkFees, err := database.DBRetrieveFeeTable()
	if err != nil {
		return result, externalAddress, callNetwork, feeToken, feeAmount, pfeeAmount, feeDiff, err
	}

	spTkList := getSupportedTokenList()
	// burnTokenAssetTag := crypto.HashToPoint(md.BurnTokenID[:])
	for _, c := range outCoins {
		feeCoin, rK := c.DoesCoinBelongToKeySet(&incFeeKeySet.KeySet)
		if feeCoin {
			if c.GetAssetTag() == nil {
				feeToken = common.PRVCoinID.String()
			} else {
				assetTag := c.GetAssetTag()
				blinder, err := coin.ComputeAssetTagBlinder(rK)
				if err != nil {
					return result, externalAddress, callNetwork, feeToken, feeAmount, pfeeAmount, feeDiff, err
				}
				rawAssetTag := new(crypto.Point).Sub(
					assetTag,
					new(crypto.Point).ScalarMult(crypto.PedCom.G[coin.PedersenRandomnessIndex], blinder),
				)
				_ = rawAssetTag
				feeToken = burnTokenInfo.TokenID
			}

			coin, err := c.Decrypt(&incFeeKeySet.KeySet)
			if err != nil {
				return result, externalAddress, callNetwork, feeToken, feeAmount, pfeeAmount, feeDiff, err
			}
			feeAmount = coin.GetValue()
		}
	}
	if feeAmount == 0 {
		return result, externalAddress, callNetwork, feeToken, feeAmount, pfeeAmount, feeDiff, errors.New("you need to paid fee")
	}
	if md != nil {
		callNetworkID, err := wcommon.GetNetworkIDFromCurrencyType(burnTokenInfo.CurrencyType)
		if err != nil {
			return result, externalAddress, callNetwork, feeToken, feeAmount, pfeeAmount, feeDiff, err
		}
		callNetwork = wcommon.GetNetworkName(callNetworkID)
		if md.Type == metadata.BurningPRVBEP20RequestMeta {
			callNetwork = wcommon.NETWORK_BSC
		}
		if md.Type == metadata.BurningPRVERC20RequestMeta {
			callNetwork = wcommon.NETWORK_ETH
		}
		burnAmount = md.BurningAmount
		externalAddress = md.RemoteAddress
	}
	if mdUnified != nil {
		callNetworkID, err := wcommon.GetNetworkIDFromCurrencyType(unshieldToken.CurrencyType)
		if err != nil {
			return result, externalAddress, callNetwork, feeToken, feeAmount, pfeeAmount, feeDiff, err
		}
		callNetwork = wcommon.GetNetworkName(callNetworkID)
		burnAmount = mdUnified.Data[0].BurningAmount
		externalAddress = mdUnified.Data[0].RemoteAddress
	}

	feeUnshield, err := estimateUnshieldFee(burnAmount, burnTokenInfo, callNetwork, networkFees, spTkList)
	if err != nil {
		return result, externalAddress, callNetwork, feeToken, feeAmount, pfeeAmount, feeDiff, err
	}

	requireFee = feeUnshield.Amount
	requireFeeToken = feeUnshield.TokenID
	if feeToken != requireFeeToken {
		return result, externalAddress, callNetwork, feeToken, feeAmount, pfeeAmount, feeDiff, fmt.Errorf("invalid fee token, fee token can't be %v must be %v ", feeToken, requireFeeToken)
	}
	feeDiff = int64(feeAmount) - int64(feeUnshield.Amount)
	if feeDiff < 0 {
		feeDiffFloat := math.Abs(float64(feeDiff))
		diffPercent := feeDiffFloat / float64(feeUnshield.Amount) * 100
		if diffPercent > wcommon.PercentFeeDiff {
			return result, externalAddress, callNetwork, feeToken, feeAmount, pfeeAmount, feeDiff, fmt.Errorf("invalid fee amount, fee amount must be at least: %v not %v", requireFee, feeAmount)
		}
	}
	pfeeAmount = feeUnshield.PrivacyFee

	if requireFeeToken == "" {
		return result, externalAddress, callNetwork, feeToken, feeAmount, pfeeAmount, feeDiff, errors.New("invalid ExternalCallAddress")
	}
	// all pass
	result = true

	return result, externalAddress, callNetwork, feeToken, feeAmount, pfeeAmount, feeDiff, nil
}

func getUnifiedUnshieldFee(tokenID, uTokenID string, amount uint64) (uint64, uint64, error) {

	methodRPC := "bridgeaggEstimateFeeByExpectedAmount"

	reqRPC := genRPCBody(methodRPC, []interface{}{
		map[string]interface{}{
			"UnifiedTokenID": uTokenID,
			"TokenID":        tokenID,
			"ExpectedAmount": amount,
		},
	})

	var responseBodyData struct {
		Result struct {
			BurntAmount       uint64
			Fee               uint64
			MaxFee            uint64
			MinReceivedAmount uint64
			ReceivedAmount    uint64
		}
		Error interface{}
	}
	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).SetBody(reqRPC).
		Post(config.FullnodeURL)
	if err != nil {
		return 0, 0, err
	}

	if responseBodyData.Error != nil {
		return 0, 0, err
	}
	return responseBodyData.Result.BurntAmount, responseBodyData.Result.Fee, err
}

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"math"
	"math/big"
	"strconv"

	"github.com/incognitochain/go-incognito-sdk-v2/coin"
	"github.com/incognitochain/go-incognito-sdk-v2/common/base58"
	"github.com/incognitochain/go-incognito-sdk-v2/crypto"
	"github.com/incognitochain/go-incognito-sdk-v2/metadata/bridge"
	metadataCommon "github.com/incognitochain/go-incognito-sdk-v2/metadata/common"
	"github.com/incognitochain/go-incognito-sdk-v2/transaction"
	"go.mongodb.org/mongo-driver/mongo"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/incognitochain/go-incognito-sdk-v2/common"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/papps"
	"github.com/incognitochain/incognito-web-based-backend/submitproof"
)

func APISubmitSwapTx(c *gin.Context) {
	var req SubmitSwapTxRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	if req.FeeRefundOTA != "" && req.FeeRefundAddress != "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "FeeRefundOTA & FeeRefundAddress can't be used as the same time"})
		return
	}

	var md *bridge.BurnForCallRequest
	var mdRaw metadataCommon.Metadata
	var isPRVTx bool
	var isUnifiedToken bool
	var outCoins []coin.Coin
	var txHash string
	var rawTxBytes []byte

	if req.FeeRefundOTA == "" && req.FeeRefundAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "FeeRefundOTA/FeeRefundAddress need to be provided one of these values"})
		return
	}

	var ok bool
	// if req.TxHash != "" {
	// 	txHash = req.TxHash

	// 	statusResult := checkPappTxSwapStatus(txHash)
	// 	if len(statusResult) > 0 {
	// 		if er, ok := statusResult["error"]; ok {
	// 			if er != "not found" {
	// 				c.JSON(200, gin.H{"Result": statusResult})
	// 				return
	// 			}
	// 		} else {
	// 			c.JSON(200, gin.H{"Result": statusResult})
	// 			return
	// 		}
	// 	}

	// 	txDetail, err := incClient.GetTx(req.TxHash)
	// 	if err != nil {
	// 		c.JSON(http.StatusBadRequest, gin.H{"Error":  err.Error()})
	// 		return
	// 	}
	// 	mdRaw = txDetail.GetMetadata()
	// 	txType := txDetail.GetType()
	// 	switch txType {
	// 	case common.TxCustomTokenPrivacyType:
	// 		isPRVTx = false
	// 		txToken := txDetail.(tx_generic.TransactionToken)
	// 		outCoins = append(outCoins, txToken.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()...)
	// 		// outCoins = append(outCoins, txDetail.GetProof().GetOutputCoins()...)
	// 	case common.TxNormalType:
	// 		isPRVTx = true
	// 		// feeToken = common.PRVCoinID.String()
	// 		outCoins = append(outCoins, txDetail.GetProof().GetOutputCoins()...)
	// 	}
	// } else {
	rawTxBytes, _, err = base58.Base58Check{}.Decode(req.TxRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid txhash")})
		return
	}

	mdRaw, isPRVTx, outCoins, txHash, err = extractDataFromRawTx(rawTxBytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	// }

	statusResult := checkPappTxSwapStatus(txHash)
	if len(statusResult) > 0 {
		if er, ok := statusResult["error"]; ok {
			if er != "not found" {
				c.JSON(200, gin.H{"Result": statusResult})
				return
			}
		} else {
			c.JSON(200, gin.H{"Result": statusResult})
			return
		}
	}

	if mdRaw == nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid tx metadata type")})
		return
	}
	md, ok = mdRaw.(*bridge.BurnForCallRequest)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid tx metadata type")})
		return
	}

	burnTokenInfo, err := getTokenInfo(md.BurnTokenID.String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid tx metadata type")})
		return
	}
	if burnTokenInfo.CurrencyType == wcommon.UnifiedCurrencyType {
		isUnifiedToken = true
	}

	tokenList, err := retrieveTokenList()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	pappTokens, err := getPappSupportedTokenList(tokenList)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	valid, networkList, feeToken, feeAmount, feeDiff, receiveToken, receiveAmount, err := checkValidTxSwap(md, outCoins, pappTokens)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid tx err:" + err.Error()})
		return
	}
	// valid = true

	burntAmount, _ := md.TotalBurningAmount()
	if valid {
		status, err := submitproof.SubmitPappTx(txHash, []byte(req.TxRaw), isPRVTx, feeToken, feeAmount, md.BurnTokenID.String(), burntAmount, receiveToken, receiveAmount, isUnifiedToken, networkList, req.FeeRefundOTA, req.FeeRefundAddress)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"Result": map[string]interface{}{"inc_request_tx_status": status}, "feeDiff": feeDiff})
		return
	}
}

func APIGetVaultState(c *gin.Context) {
	var responseBodyData APIRespond
	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).
		Get(config.CoinserviceURL + "/bridge/aggregatestate")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(200, responseBodyData)
}

func APIEstimateSwapFee(c *gin.Context) {
	var req EstimateSwapRequest
	err := c.MustBindWith(&req, binding.JSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	switch req.Network {
	case "inc", "eth", "bsc", "plg":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"Error": "unsupported network"})
		return
	}

	_, ok := new(big.Float).SetString(req.Amount)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Amount isn't a valid number"})
		return
	}

	_, ok = new(big.Float).SetString(req.Slippage)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Slippage isn't a valid number"})
		return
	}

	slippage, err := verifySlippage(req.Slippage)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	networkID := wcommon.GetNetworkID(req.Network)
	if networkID == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid network"})
		return
	}

	var response struct {
		Result interface{}
		Error  interface{}
	}
	var result EstimateSwapRespond
	result.Networks = make(map[string]interface{})
	result.NetworksError = make(map[string]interface{})

	tkFromInfo, err := getTokenInfo(req.FromToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	tkToInfo, err := getTokenInfo(req.ToToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	tkToNetworkID := 0
	if tkToInfo.CurrencyType != wcommon.UnifiedCurrencyType {
		tkToNetworkID, _ = getNetworkIDFromCurrencyType(tkToInfo.CurrencyType)
	}

	networkErr := make(map[string]interface{})
	var pdexEstimate []QuoteDataResp

	if req.Network == "inc" {
		// pdexresult := estimateSwapFeeWithPdex(req.FromToken, req.ToToken, req.Amount, slippage, tkFromInfo)
		// if pdexresult != nil {
		// 	pdexEstimate = append(pdexEstimate, *pdexresult)
		// }
	}

	supportedNetworks := []int{}

	outofVaultNetworks := []int{}
	if tkFromInfo.CurrencyType == wcommon.UnifiedCurrencyType {
		amount := new(big.Float)
		amount, errBool := amount.SetString(req.Amount)
		if !errBool {
			c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid amount")})
			return
		}
		dm := new(big.Float)
		dm.SetFloat64(math.Pow10(tkFromInfo.PDecimals))
		amountUint64, _ := amount.Mul(amount, dm).Uint64()

		supportedOutNetworks := []int{}
		for _, v := range tkFromInfo.ListUnifiedToken {
			if networkID == 0 {
				//check all vaults
				isEnoughVault, err := checkEnoughVault(req.FromToken, v.TokenID, amountUint64)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
					return
				}
				if isEnoughVault {
					supportedOutNetworks = append(supportedOutNetworks, v.NetworkID)
				} else {
					outofVaultNetworks = append(outofVaultNetworks, v.NetworkID)
					networkErr[wcommon.GetNetworkName(v.NetworkID)] = "not enough token in vault"
				}
			} else {
				//check 1 vault only
				if networkID == v.NetworkID {
					isEnoughVault, err := checkEnoughVault(req.FromToken, v.TokenID, amountUint64)
					if err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
						return
					}
					if isEnoughVault {
						supportedOutNetworks = append(supportedOutNetworks, v.NetworkID)
					} else {
						outofVaultNetworks = append(outofVaultNetworks, v.NetworkID)
						networkErr[wcommon.GetNetworkName(v.NetworkID)] = "not enough token in vault"
					}
				}
			}
		}
		if len(supportedOutNetworks) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "The amount exceeds the swap limit. Please retry with another token or switch to other pApps"})
			return
		}
		fmt.Println("pass check vault", "supportedOutNetworks", supportedOutNetworks)

		// check supported to token
		for _, spNetID := range supportedOutNetworks {
			if tkToInfo.CurrencyType == wcommon.UnifiedCurrencyType {
				for _, v := range tkToInfo.ListUnifiedToken {
					if v.NetworkID == spNetID {
						supportedNetworks = append(supportedNetworks, spNetID)
					}
				}
			} else {
				if tkToNetworkID == spNetID {
					supportedNetworks = append(supportedNetworks, spNetID)
				}
			}
		}
	} else {
		supportedOutNetworks := []int{}
		tkFromNetworkID, _ := getNetworkIDFromCurrencyType(tkFromInfo.CurrencyType)
		if tkFromNetworkID > 0 {
			if networkID == tkFromNetworkID {
				supportedOutNetworks = append(supportedOutNetworks, tkFromNetworkID)
			} else {
				if networkID == 0 {
					supportedOutNetworks = append(supportedOutNetworks, tkFromNetworkID)
				} else {
					c.JSON(http.StatusBadRequest, gin.H{"Error": "No supported networks found"})
					return
				}
			}

			// check supported to token
			for _, spNetID := range supportedOutNetworks {
				if tkToInfo.CurrencyType == wcommon.UnifiedCurrencyType {
					for _, v := range tkToInfo.ListUnifiedToken {
						if v.NetworkID == spNetID {
							supportedNetworks = append(supportedNetworks, spNetID)
						}
					}
				} else {
					if tkToNetworkID == spNetID {
						supportedNetworks = append(supportedNetworks, spNetID)
					}
				}
			}
		}
	}
	if len(supportedNetworks) == 0 {
		for net, v := range networkErr {
			result.NetworksError[net] = v
		}
		if req.Network == "inc" && len(pdexEstimate) != 0 {
			result.Networks["inc"] = pdexEstimate
			response.Result = result
			c.JSON(200, response)
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"Error": NotTradeable.Error()})
		return
	}

	networksInfo, err := getBridgeNetworkInfos()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	tokenList, err := retrieveTokenList()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	spTkList, err := getPappSupportedTokenList(tokenList)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	networkFees, err := database.DBRetrieveFeeTable()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	for _, network := range supportedNetworks {
		data, err := estimateSwapFee(req.FromToken, req.ToToken, req.Amount, network, spTkList, networksInfo, networkFees, tkFromInfo, slippage)
		if err != nil {
			networkErr[wcommon.GetNetworkName(network)] = err.Error()
		} else {
			result.Networks[wcommon.GetNetworkName(network)] = data
		}
	}

	for net, v := range networkErr {
		result.NetworksError[net] = v
	}

	// if req.Network != "inc" && len(networkErr) == 1 {
	// 	response.Result = result
	// 	response.Error = NotTradeable.Error()
	// 	c.JSON(200, response)
	// 	return
	// }
	// if req.Network == "inc" && len(networkErr) == len(supportedNetworks) && pdexEstimate == nil {
	// 	response.Result = result
	// 	response.Error = NotTradeable.Error()
	// 	c.JSON(200, response)
	// 	return
	// }
	if len(pdexEstimate) != 0 {
		result.Networks["inc"] = pdexEstimate
	}
	if len(result.Networks) == 0 && len(pdexEstimate) == 0 {
		response.Error = NotTradeable.Error()
	}

	response.Result = result

	c.JSON(200, response)

}

func estimateSwapFeeWithPdex(fromToken, toToken, amount string, slippage *big.Float, tkFromInfo *wcommon.TokenInfo) *QuoteDataResp {
	type APIRespond struct {
		Result map[string]PdexEstimateRespond
		Error  *string
	}

	amountBig, _ := new(big.Float).SetString(amount)
	amountBig = amountBig.Mul(amountBig, new(big.Float).SetFloat64(math.Pow10(tkFromInfo.PDecimals)))

	amountRaw, _ := amountBig.Uint64()
	var responseBodyData APIRespond

	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).
		Get(config.CoinserviceURL + "/pdex/v3/estimatetrade?" + fmt.Sprintf("selltoken=%v&buytoken=%v&sellamount=%v", fromToken, toToken, amountRaw))
	if err != nil {
		log.Println("estimateSwapFeeWithPdex", err)
		return nil
	}
	if responseBodyData.Error != nil {
		log.Println("estimateSwapFeeWithPdex", errors.New(*responseBodyData.Error))
		return nil
	}

	pdexResult := PdexEstimateRespond{}
	v, ok := responseBodyData.Result["FeeToken"]
	if !ok {
		return nil
	}

	amountOutBigFloat := new(big.Float).SetFloat64(v.MaxGet)
	if slippage != nil {
		sl := new(big.Float).SetFloat64(0.01)
		sl = sl.Mul(sl, slippage)
		sl = new(big.Float).Sub(new(big.Float).SetFloat64(1), sl)
		amountOutBigFloat = amountOutBigFloat.Mul(amountOutBigFloat, sl)
		amountOutFloat, _ := amountOutBigFloat.Float64()
		v.MaxGet = math.Floor(amountOutFloat)
	}
	pdexResult = v

	tkToInfo, _ := getTokenInfo(toToken)
	amountOutBig := new(big.Float).SetFloat64(pdexResult.MaxGet)
	amountOutBig = amountOutBig.Mul(amountOutBig, new(big.Float).SetFloat64(math.Pow10(-tkToInfo.PDecimals)))
	amountOut := amountOutBig.String()

	result := QuoteDataResp{
		AppName:      "pdex",
		AmountIn:     amount,
		AmountInRaw:  fmt.Sprintf("%v", amountRaw),
		AmountOut:    fmt.Sprintf("%v", amountOut),
		AmountOutRaw: fmt.Sprintf("%f", pdexResult.MaxGet),
		Paths:        pdexResult.TokenRoute,
		PoolPairs:    pdexResult.Route,
		ImpactAmount: fmt.Sprintf("%f", pdexResult.ImpactAmount),
		Fee:          []PappNetworkFee{{TokenID: fromToken, Amount: pdexResult.Fee}},
	}

	return &result
}

func estimateSwapFee(fromToken, toToken, amount string, networkID int, spTkList []PappSupportedTokenData, networksInfo []wcommon.BridgeNetworkData, networkFees *wcommon.ExternalNetworksFeeData, fromTokenInfo *wcommon.TokenInfo, slippage *big.Float) ([]QuoteDataResp, error) {

	feeTokenWhiteList, err := retrieveFeeTokenWhiteList()
	if err != nil {
		log.Println(err)
	}
	result := []QuoteDataResp{}
	feeAddress := ""
	feeAddressShardID := byte(0)
	if incFeeKeySet != nil {
		feeAddress, err = incFeeKeySet.GetPaymentAddress()
		if err != nil {
			return nil, err
		}
		feeAddressShardID, _ = common.GetShardIDsFromPublicKey(incFeeKeySet.KeySet.PaymentAddress.Pk)
	}
	log.Println("estimateSwapFee for", fromToken, toToken, amount, networkID)
	networkName := wcommon.GetNetworkName(networkID)
	pappList, err := database.DBRetrievePAppsByNetwork(networkName)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("no supported papps found")
		}
		fmt.Println("DBRetrievePAppsByNetwork", err)
		return nil, err
	}

	vaultData, err := database.DBGetPappVaultData(networkName, wcommon.PappTypeSwap)
	if err != nil {
		fmt.Println("DBGetPappVaultData", err)
		return nil, err
	}

	pTokenContract1, err := getpTokenContractID(fromToken, networkID, spTkList)
	if err != nil {
		log.Println("err get pTokenContract1")
		if len(fromTokenInfo.ListUnifiedToken) > 0 {
			for _, cTk := range fromTokenInfo.ListUnifiedToken {
				pTokenContract1, _ = getpTokenContractID(cTk.TokenID, networkID, spTkList)
				if pTokenContract1 != nil {
					break
				}
			}
			if pTokenContract1 == nil {
				return nil, errors.New("can't find contractID for token " + fromToken)
			}
		} else {
			return nil, err
		}
	}
	toTokenInfo, err := getTokenInfo(toToken)
	if err != nil {
		return nil, err
	}

	pTokenContract2, err := getpTokenContractID(toToken, networkID, spTkList)
	if err != nil {
		log.Println("err get pTokenContract2")
		if toTokenInfo.MovedUnifiedToken {
			utokenInfo, err := getParentUToken(toToken)
			if err != nil {
				return nil, err
			}
			pTokenContract2, err = getpTokenContractID(utokenInfo.TokenID, networkID, spTkList)
			if err != nil {
				return nil, err
			}
		} else {
			for _, cTk := range toTokenInfo.ListUnifiedToken {
				pTokenContract2, _ = getpTokenContractID(cTk.TokenID, networkID, spTkList)
				if pTokenContract2 != nil {
					break
				}
			}
			if pTokenContract2 == nil {
				return nil, errors.New("can't find contractID for token " + toToken)
			}
		}
	}

	log.Println("done get pTokenContract1")
	networkChainId := ""
	for _, v := range networksInfo {
		if networkName == v.Network {
			networkChainId = v.ChainID
			break
		}
	}

	if _, ok := networkFees.GasPrice[networkName]; !ok {
		return nil, errors.New("network gasPrice not found")
	}
	gasPrice := networkFees.GasPrice[networkName]

	nativeCurrentType := wcommon.GetNativeNetworkCurrencyType(networkName)
	nativeToken, err := getNativeTokenData(spTkList, nativeCurrentType)
	if err != nil {
		return nil, err
	}
	isUnifiedNativeToken := false
	// if fromTokenInfo
	if pTokenContract1.CurrencyType == nativeCurrentType {
		isUnifiedNativeToken = true
	}
	if pTokenContract1.CurrencyType == wcommon.UnifiedCurrencyType {
		for _, v := range fromTokenInfo.ListUnifiedToken {
			if v.CurrencyType == nativeCurrentType {
				isUnifiedNativeToken = true
			}
		}
	}

	log.Println("len(pappList.ExchangeApps)", len(pappList.ExchangeApps), pappList.ExchangeApps)

	if len(pappList.ExchangeApps) == 0 {
		log.Println("len(pappList.ExchangeApps) == 0")
		return nil, errors.New("no ExchangeApps found")
	}

	amountFloat := new(big.Float)
	amountFloat, ok := amountFloat.SetString(amount)
	if !ok {
		return nil, fmt.Errorf("amount is not a number")
	}
	amountBigFloat := ConvertToNanoIncognitoToken(amountFloat, int64(pTokenContract1.Decimals)) //amount *big.Float, decimal int64, return *big.Float
	amountInBig, _ := amountBigFloat.Int(nil)

	amountInBig0 := new(big.Int).Set(amountInBig)

	additionalTokenInFee := amountInBig0.Div(amountInBig0, new(big.Int).SetUint64(1000))

	toTokenDecimal := big.NewFloat(math.Pow10(-pTokenContract2.Decimals))

	isFeeWhitelist := false
	if _, ok := feeTokenWhiteList[fromToken]; ok {
		isFeeWhitelist = true
	}

	for appName, endpoint := range pappList.ExchangeApps {
		switch appName {
		case "uniswap":
			fmt.Println("uniswap", networkID, pTokenContract1.ContractID, pTokenContract2.ContractID)
			realAmountIn := amountFloat
			realAmountIn = realAmountIn.Mul(realAmountIn, new(big.Float).SetFloat64(0.997))
			realAmountInFloat, _ := realAmountIn.Float64()
			realAmountInStr := fmt.Sprintf("%f", realAmountInFloat)

			data, err := papps.UniswapQuote(pTokenContract1.ContractID, pTokenContract2.ContractID, realAmountInStr, networkChainId, true, endpoint)
			if err != nil {
				log.Println(err)
				continue
			}
			quote, feePaths, err := uniswapDataExtractor(data)
			if err != nil {
				log.Println(err)
				continue
			}
			// fees := []PappNetworkFee{}

			estGasUsedStr := quote.Data.EstimatedGasUsed
			estGasUsed, err := strconv.ParseUint(estGasUsedStr, 10, 64)
			if err != nil {
				log.Println(err)
				continue
			}
			estGasUsed += 200000
			estGasUsed = estGasUsed + estGasUsed/3*2

			amountOutBigFloat0, _ := new(big.Float).SetString(quote.Data.AmountOutRaw)
			rate := new(big.Float).Quo(amountOutBigFloat0, new(big.Float).Set(amountBigFloat))
			gasFee := (estGasUsed * gasPrice)

			fees := getFee(isFeeWhitelist, isUnifiedNativeToken, nativeToken, rate, gasFee, fromToken, fromTokenInfo, pTokenContract1, toTokenDecimal, additionalTokenInFee)
			// if isUnifiedNativeToken {
			// 	gasFeeFloat := new(big.Float).SetUint64(gasFee)
			// 	gasFeeFloat = gasFeeFloat.Mul(gasFeeFloat, rate)
			// 	gasFeeInt, _ := gasFeeFloat.Uint64()
			// 	fees = append(fees, PappNetworkFee{
			// 		Amount:           ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", gasFee+additionalTokenInFee.Uint64()), int64(nativeToken.Decimals), int64(nativeToken.PDecimals)),
			// 		TokenID:          fromToken,
			// 		AmountInBuyToken: gasFeeInt,
			// 	})
			// } else {
			// 	gasFeePRV := float64(gasFee+additionalTokenInFee.Uint64()) * nativeToken.PricePrv
			// 	gasFeeFromToken := gasFeePRV / fromTokenInfo.PricePrv
			// 	gasFeeFloat := new(big.Float).SetFloat64(gasFeeFromToken)
			// 	gasFeeFloat = gasFeeFloat.Mul(gasFeeFloat, rate)
			// 	gasFeeIntToToken, _ := gasFeeFloat.Uint64()

			// 	if pTokenContract1.CurrencyType == wcommon.UnifiedCurrencyType {
			// 		fees = append(fees, PappNetworkFee{
			// 			Amount:           ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", uint64(gasFeePRV/fromTokenInfo.PricePrv)), int64(nativeToken.Decimals), int64(nativeToken.PDecimals)),
			// 			TokenID:          fromToken,
			// 			AmountInBuyToken: gasFeeIntToToken,
			// 		})
			// 	} else {
			// 		fees = append(fees, PappNetworkFee{
			// 			Amount:           ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", uint64(gasFeePRV)), int64(nativeToken.Decimals), int64(nativeToken.PDecimals)),
			// 			TokenID:          common.PRVCoinID.String(),
			// 			AmountInBuyToken: gasFeeIntToToken,
			// 		})
			// 	}
			// }
			if estGasUsed > wcommon.EVMGasLimitETH {
				return nil, errors.New("estimated used gas exceed gas limit")
			}
			// fees = append(fees, PappNetworkFee{
			// 	Amount:  estGasUsed,
			// 	TokenID: "estGasUsed",
			// })
			// fees = append(fees, PappNetworkFee{
			// 	Amount:  gasPrice,
			// 	TokenID: "gasPrice",
			// })

			vaultAddress := ethcommon.Address{}
			err = vaultAddress.UnmarshalText([]byte(vaultData.ContractAddress))
			if err != nil {
				return nil, err
			}

			amountOutBigFloat, _ := new(big.Float).SetString(quote.Data.AmountOutRaw)
			if slippage != nil {
				sl := new(big.Float).SetFloat64(0.01)
				sl = sl.Mul(sl, slippage)
				sl = new(big.Float).Sub(new(big.Float).SetFloat64(1), sl)
				amountOutBigFloat = amountOutBigFloat.Mul(amountOutBigFloat, sl)
			}
			amountOutBig, _ := amountOutBigFloat.Int(nil)

			paths := []ethcommon.Address{}
			traversedTk := make(map[string]struct{})

			for _, route := range quote.Data.Route[0] {
				tokenAddress := ethcommon.Address{}
				err = tokenAddress.UnmarshalText([]byte(route.TokenIn.Address))
				if err != nil {
					return nil, err
				}
				paths = append(paths, tokenAddress)
				traversedTk[route.TokenIn.Address] = struct{}{}

				tokenAddress2 := ethcommon.Address{}
				err = tokenAddress2.UnmarshalText([]byte(route.TokenOut.Address))
				if err != nil {
					return nil, err
				}
				if _, ok := traversedTk[route.TokenOut.Address]; !ok {
					paths = append(paths, tokenAddress2)
				}
				traversedTk[route.TokenOut.Address] = struct{}{}

			}

			tokenOutAddress := ethcommon.Address{}
			err = tokenOutAddress.UnmarshalText([]byte(pTokenContract2.ContractID))
			if err != nil {
				return nil, err
			}
			// paths = append(paths, tokenOutAddress)

			contract, ok := pappList.AppContracts[appName]
			if !ok {
				return nil, errors.New("contract not found " + appName)
			}

			uniswapProxy := ethcommon.HexToAddress(contract)
			recipient := vaultAddress
			isNativeTokenOut := false
			if wcommon.CheckIsWrappedNativeToken(tokenOutAddress.Hex(), networkID) {
				isNativeTokenOut = true
				recipient = uniswapProxy
			}

			calldata, err := papps.BuildCallDataUniswap(paths, recipient, feePaths[0], amountInBig, amountOutBig, isNativeTokenOut)
			if err != nil {
				log.Println("Error building call data: ", err)
				calldata = err.Error()
			}

			pTokenAmount := new(big.Float).Mul(amountOutBigFloat, toTokenDecimal)

			pTkAmountFloatStr := pTokenAmount.Text('f', -1)

			pathsList := []string{}
			for _, v := range paths {
				pathsList = append(pathsList, v.String())
			}
			result = append(result, QuoteDataResp{
				AppName:           appName,
				AmountIn:          amount,
				AmountInRaw:       quote.Data.AmountIn,
				AmountOut:         pTkAmountFloatStr,
				AmountOutRaw:      amountOutBig.String(),
				Paths:             pathsList,
				Fee:               fees,
				Calldata:          calldata,
				CallContract:      contract,
				FeeAddress:        feeAddress,
				FeeAddressShardID: int(feeAddressShardID),
				RouteDebug:        quote.Data.Route,
			})
			log.Println("done estimate uniswap")
		case "pancake":
			fmt.Println("pancake", networkID, pTokenContract1.ContractID, pTokenContract2.ContractID)
			realAmountIn := amountFloat
			if strings.Contains(config.NetworkID, "testnet") {
				realAmountIn = realAmountIn.Mul(realAmountIn, new(big.Float).SetFloat64(0.998))
			} else {
				realAmountIn = realAmountIn.Mul(realAmountIn, new(big.Float).SetFloat64(0.9975))
			}
			realAmountInFloat, _ := realAmountIn.Float64()
			realAmountInStr := fmt.Sprintf("%f", realAmountInFloat)

			tokenMap, err := buildPancakeTokenMap(spTkList)
			if err != nil {
				log.Println(err)
				continue
			}

			tokenMapBytes, err := json.Marshal(tokenMap)
			if err != nil {
				log.Println(err)
				continue
			}

			log.Println("tokenMapBytes", string(tokenMapBytes))
			data, err := papps.PancakeQuote(pTokenContract1.ContractID, pTokenContract2.ContractID, realAmountInStr, networkChainId, pTokenContract1.Symbol, pTokenContract2.Symbol, pTokenContract1.Decimals, pTokenContract2.Decimals, false, endpoint, string(tokenMapBytes))
			if err != nil {
				log.Println(err)
				continue
			}
			quote, err := pancakeDataExtractor(data)
			if err != nil {
				log.Println(err)
				continue
			}
			// fees := []PappNetworkFee{}
			// estGasUsedStr := quote.Data.EstimatedGasUsed
			// estGasUsed, err := strconv.ParseUint(estGasUsedStr, 10, 64)
			// if err != nil {
			// 	return nil, err
			// }
			estGasUsed := uint64(wcommon.EVMGasLimitPancake)

			amountOutBigFloat0, _ := new(big.Float).SetString(quote.Data.Outputs[len(quote.Data.Outputs)-1])
			rate := new(big.Float).Quo(amountOutBigFloat0, new(big.Float).Set(amountBigFloat))
			gasFee := (estGasUsed * gasPrice)

			fees := getFee(isFeeWhitelist, isUnifiedNativeToken, nativeToken, rate, gasFee, fromToken, fromTokenInfo, pTokenContract1, toTokenDecimal, additionalTokenInFee)
			// if isUnifiedNativeToken {
			// 	fees = append(fees, PappNetworkFee{
			// 		Amount:  ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", estGasUsed*gasPrice), int64(nativeToken.Decimals), int64(nativeToken.PDecimals)),
			// 		TokenID: fromToken,
			// 	})
			// } else {
			// 	if pTokenContract1.CurrencyType == wcommon.UnifiedCurrencyType {
			// 		fees = append(fees, PappNetworkFee{
			// 			Amount:  ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", uint64(float64(estGasUsed*gasPrice)*nativeToken.PricePrv/fromTokenInfo.PricePrv)), int64(nativeToken.Decimals), int64(nativeToken.PDecimals)),
			// 			TokenID: fromToken,
			// 		})
			// 	} else {
			// 		fees = append(fees, PappNetworkFee{
			// 			Amount:  ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", uint64(float64(estGasUsed*gasPrice)*nativeToken.PricePrv)), int64(nativeToken.Decimals), int64(nativeToken.PDecimals)),
			// 			TokenID: common.PRVCoinID.String(),
			// 		})
			// 	}
			// }

			log.Println("len(quote.Data.Outputs)", len(quote.Data.Outputs), quote.Data.Outputs, quote.Data.Outputs[len(quote.Data.Outputs)-1])

			amountOutBigFloat, _ := new(big.Float).SetString(quote.Data.Outputs[len(quote.Data.Outputs)-1])
			if slippage != nil {
				sl := new(big.Float).SetFloat64(0.01)
				sl = sl.Mul(sl, slippage)
				sl = new(big.Float).Sub(new(big.Float).SetFloat64(1), sl)
				amountOutBigFloat = amountOutBigFloat.Mul(amountOutBigFloat, sl)
			}
			amountOutBig, _ := amountOutBigFloat.Int(nil)

			log.Println("outInt64", amountOutBigFloat.String())

			paths := []ethcommon.Address{}

			for _, token := range quote.Data.Route {
				tokenAddress := ethcommon.Address{}
				err = tokenAddress.UnmarshalText([]byte(token))
				if err != nil {
					return nil, err
				}
				paths = append(paths, tokenAddress)
			}

			calldata, err := papps.BuildCallDataPancake(paths, amountInBig, amountOutBig, isUnifiedNativeToken)
			if err != nil {
				log.Println("Error building call data: ", err)
				err1 := errors.New("Error building call data: " + err.Error())
				return nil, err1
			}

			amountOut, ok := new(big.Float).SetString(amountOutBig.String())
			if !ok {
				err = errors.New("Error building call data: amountout out of range")
				log.Println(err.Error())
				return nil, err

			}
			pTokenAmount := new(big.Float).Mul(amountOut, toTokenDecimal)
			pTkAmountFloatStr := pTokenAmount.Text('f', -1)
			contract, ok := pappList.AppContracts[appName]
			if !ok {
				return nil, errors.New("contract not found " + appName)
			}

			result = append(result, QuoteDataResp{
				AppName:           appName,
				AmountIn:          amount,
				AmountOut:         pTkAmountFloatStr,
				AmountOutRaw:      amountOutBig.String(),
				Paths:             quote.Data.Route,
				Fee:               fees,
				Calldata:          calldata,
				CallContract:      contract,
				FeeAddress:        feeAddress,
				FeeAddressShardID: int(feeAddressShardID),
				ImpactAmount:      fmt.Sprintf("%.2f", quote.Data.Impact),
			})

			log.Println("done estimate pancake")
		case "curve":
			fmt.Println("curve", networkID, pTokenContract1.ContractID, pTokenContract2.ContractID)
			poolList, err := getCurvePoolIndex(config.ShieldService)
			if err != nil {
				log.Println("curve", err)
				continue
			}
			token1PoolIndex, curvePoolAddress1, err := getTokenCurvePoolIndex(pTokenContract1.ContractID, poolList)
			if err != nil {
				log.Println("curve", err)
				continue
			}
			token2PoolIndex, _, err := getTokenCurvePoolIndex(pTokenContract2.ContractID, poolList)
			if err != nil {
				log.Println("curve", err)
				continue
			}

			amountFloat := new(big.Float)
			amountFloat, ok := amountFloat.SetString(amount)
			if !ok {
				return nil, fmt.Errorf("amount is not a number")
			}
			amountBigFloat := ConvertToNanoIncognitoToken(amountFloat, int64(pTokenContract1.Decimals)) //amount *big.Float, decimal int64, return *big.Float
			log.Println("amountBigFloat: ", amountBigFloat.String())

			//fee 0.04%
			realAmountIn := amountBigFloat
			realAmountIn = realAmountIn.Mul(realAmountIn, new(big.Float).SetFloat64(0.9996))

			// convert float to bigin:
			amountBigInt, _ := realAmountIn.Int(nil)

			if amountBigInt == nil {
				return nil, errors.New("invalid amount")
			}

			i := big.NewInt(int64(token1PoolIndex))
			j := big.NewInt(int64(token2PoolIndex))

			curvePool := ethcommon.HexToAddress(curvePoolAddress1)

			networkInfo, err := database.DBGetBridgeNetworkInfo(networkName)
			if err != nil {
				log.Println(err)
				continue
			}
			var amountOut *big.Int
			var calldata string

			for _, endpoint := range networkInfo.Endpoints {
				evmClient, err := ethclient.Dial(endpoint)
				if err != nil {
					log.Println(err)
					continue
				}
				amountOut, err = papps.CurveQuote(evmClient, amountBigInt, i, j, curvePool)
				if err != nil {
					log.Println(err)
					continue
				} else {
					amountOutBigFloat := new(big.Float).SetInt(amountOut)
					if slippage != nil {
						sl := new(big.Float).SetFloat64(0.01)
						sl = sl.Mul(sl, slippage)
						sl = new(big.Float).Sub(new(big.Float).SetFloat64(1), sl)
						amountOutBigFloat = amountOutBigFloat.Mul(amountOutBigFloat, sl)
					}
					amountOutBigFloat.Int(amountOut)
					calldata, err = papps.BuildCurveCallData(amountBigInt, amountOut, i, j, curvePool)
					if err != nil {
						log.Println("Error building call data: ", err)
						err1 := errors.New("Error building call data: " + err.Error())
						return nil, err1
					}
					break
				}
			}

			amountOutFloat := new(big.Float)
			amountOutFloat, _ = amountOutFloat.SetString(amountOut.String())
			pTokenAmount := new(big.Float).Mul(amountOutFloat, toTokenDecimal)

			estGasUsed := uint64(wcommon.EVMGasLimit)

			amountOutBigFloat0 := new(big.Float).SetInt(amountOut)
			rate := new(big.Float).Quo(amountOutBigFloat0, new(big.Float).Set(amountBigFloat))
			gasFee := (estGasUsed * gasPrice)

			fees := getFee(isFeeWhitelist, isUnifiedNativeToken, nativeToken, rate, gasFee, fromToken, fromTokenInfo, pTokenContract1, toTokenDecimal, additionalTokenInFee)
			// if isUnifiedNativeToken {
			// 	fees = append(fees, PappNetworkFee{
			// 		Amount:  ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", estGasUsed*gasPrice), int64(nativeToken.Decimals), int64(nativeToken.PDecimals)),
			// 		TokenID: fromToken,
			// 	})
			// } else {
			// 	if pTokenContract1.CurrencyType == wcommon.UnifiedCurrencyType {
			// 		fees = append(fees, PappNetworkFee{
			// 			Amount:  ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", uint64(float64(estGasUsed*gasPrice)*nativeToken.PricePrv/fromTokenInfo.PricePrv)), int64(nativeToken.Decimals), int64(nativeToken.PDecimals)),
			// 			TokenID: fromToken,
			// 		})
			// 	} else {
			// 		fees = append(fees, PappNetworkFee{
			// 			Amount:  ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", uint64(float64(estGasUsed*gasPrice)*nativeToken.PricePrv)), int64(nativeToken.Decimals), int64(nativeToken.PDecimals)),
			// 			TokenID: common.PRVCoinID.String(),
			// 		})
			// 	}
			// }
			contract, ok := pappList.AppContracts[appName]
			if !ok {
				return nil, errors.New("contract not found " + appName)
			}
			result = append(result, QuoteDataResp{
				AppName:           appName,
				AmountIn:          amount,
				AmountOut:         pTokenAmount.String(),
				AmountOutRaw:      amountOut.String(),
				Fee:               fees,
				CallContract:      contract,
				Calldata:          calldata,
				FeeAddress:        feeAddress,
				Paths:             []string{pTokenContract1.ContractID, pTokenContract2.ContractID},
				FeeAddressShardID: int(feeAddressShardID),
			})
			log.Println("done estimate curve")
		}
	}
	if len(result) == 0 {
		return nil, errors.New("no data")
	}
	return result, nil
}

func cacheVaultState() {
	for {
		var responseBodyData APIRespond
		_, err := restyClient.R().
			EnableTrace().
			SetHeader("Content-Type", "application/json").
			SetResult(&responseBodyData).
			Get(config.CoinserviceURL + "/bridge/aggregatestate")
		if err != nil {
			log.Println("cacheVaultState", err.Error())
			continue
		}

		err = cacheStoreCustom(cacheVaultStateKey, responseBodyData, 30*time.Second)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(15 * time.Second)
	}
}

func cacheBridgeNetworkInfos() {
	for {
		networkInfo, err := database.DBGetBridgeNetworkInfos()
		if err != nil {
			log.Println("cacheBridgeNetworkInfos", err.Error())
			continue
		} else {
			err = cacheStoreCustom(cacheNetworkInfosKey, networkInfo, 30*time.Second)
			if err != nil {
				log.Println(err)
			}
		}
		time.Sleep(15 * time.Second)
	}
}

func cacheSupportedPappsTokens() {
	for {
		var responseBodyData APIRespond
		_, err := restyClient.R().
			EnableTrace().
			SetHeader("Content-Type", "application/json").
			SetResult(&responseBodyData).
			Get(config.ShieldService + "/trade/supported-tokens")
		if err != nil {
			log.Println("cacheSupportedPappsTokens", err.Error())
			continue
		} else {
			err = cacheStoreCustom(cacheSupportedPappsTokensKey, responseBodyData, 30*time.Second)
			if err != nil {
				log.Println(err)
			}
		}

		time.Sleep(15 * time.Second)
	}
}

func cacheTokenList() {
	for {
		type APIRespond struct {
			Result []wcommon.TokenInfo
			Error  *string
		}

		var responseBodyData APIRespond
		_, err := restyClient.R().
			EnableTrace().
			SetHeader("Content-Type", "application/json").
			SetResult(&responseBodyData).
			Get(config.CoinserviceURL + "/coins/tokenlist")
		if err != nil {
			log.Println("cacheTokenList", err.Error())
			continue
		} else {
			err = cacheStoreCustom(cacheTokenListKey, responseBodyData, 30*time.Second)
			if err != nil {
				log.Println(err)
			}
		}
		time.Sleep(15 * time.Second)
	}
}

func retrieveTokenList() ([]wcommon.TokenInfo, error) {
	type APIRespond struct {
		Result []wcommon.TokenInfo
		Error  *string
	}

	var responseBodyData APIRespond

	err := cacheGet(cacheTokenListKey, &responseBodyData)
	if err != nil {
		_, err := restyClient.R().
			EnableTrace().
			SetHeader("Content-Type", "application/json").
			SetResult(&responseBodyData).
			Get(config.CoinserviceURL + "/coins/tokenlist")
		if err != nil {
			return nil, err
		}
		if responseBodyData.Error != nil {
			return nil, errors.New(*responseBodyData.Error)
		}
		return responseBodyData.Result, nil
	}

	return responseBodyData.Result, nil
}

func getPappSupportedTokenList(tokenList []wcommon.TokenInfo) ([]PappSupportedTokenData, error) {

	var responseBodyData struct {
		Result []PappSupportedTokenData
		Error  *struct {
			Code    int
			Message string
		} `json:"Error"`
	}

	re, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		Get(config.ShieldService + "/trade/supported-tokens")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(re.Body(), &responseBodyData)
	if err != nil {
		return nil, err
	}

	if responseBodyData.Error != nil {
		return nil, errors.New(responseBodyData.Error.Message)
	}
	result := []PappSupportedTokenData{}

	dblist, err := database.DBGetPappSupportedToken()
	if err != nil {
		return nil, err
	}

	result = append(result, responseBodyData.Result...)
	result = append(result, transformShieldServicePappSupportedToken(dblist, tokenList)...)

	return result, nil
}

func getBridgeNetworkInfos() ([]wcommon.BridgeNetworkData, error) {
	var result []wcommon.BridgeNetworkData
	err := cacheGet(cacheNetworkInfosKey, &result)
	if err != nil {
		networkInfo, err := database.DBGetBridgeNetworkInfos()
		if err != nil {
			return nil, err
		}
		return networkInfo, nil
	}
	return result, nil
}

func checkValidTxSwap(md *bridge.BurnForCallRequest, outCoins []coin.Coin, spTkList []PappSupportedTokenData) (bool, []string, string, uint64, int64, string, uint64, error) {
	var feeAmount uint64
	var feeToken string

	var requireFee uint64
	var requireFeeToken string

	var receiveToken string
	var receiveAmount uint64

	var result bool
	feeDiff := int64(-1)
	callNetworkList := []string{}
	networkInfo, err := getBridgeNetworkInfos()
	if err != nil {
		return result, callNetworkList, feeToken, feeAmount, feeDiff, receiveToken, receiveAmount, err
	}
	networkFees, err := database.DBRetrieveFeeTable()
	if err != nil {
		return result, callNetworkList, feeToken, feeAmount, feeDiff, receiveToken, receiveAmount, err
	}
	tokenInfo, err := getTokenInfo(md.BurnTokenID.String())
	if err != nil {
		return result, callNetworkList, feeToken, feeAmount, feeDiff, receiveToken, receiveAmount, err
	}

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
					return result, callNetworkList, feeToken, feeAmount, feeDiff, receiveToken, receiveAmount, err
				}
				rawAssetTag := new(crypto.Point).Sub(
					assetTag,
					new(crypto.Point).ScalarMult(crypto.PedCom.G[coin.PedersenRandomnessIndex], blinder),
				)
				_ = rawAssetTag
				feeToken = md.BurnTokenID.String()
			}

			coin, err := c.Decrypt(&incFeeKeySet.KeySet)
			if err != nil {
				return result, callNetworkList, feeToken, feeAmount, feeDiff, receiveToken, receiveAmount, err
			}
			feeAmount = coin.GetValue()
		}
	}
	if feeAmount == 0 {
		return result, callNetworkList, feeToken, feeAmount, feeDiff, receiveToken, receiveAmount, errors.New("you need to paid fee")
	}

	for _, v := range md.Data {
		callNetworkList = append(callNetworkList, wcommon.GetNetworkName(int(v.ExternalNetworkID)))
		receiveToken, err = getTokenIDByContractID(v.ReceiveToken, int(v.ExternalNetworkID), spTkList)
		if err != nil {
			return result, callNetworkList, feeToken, feeAmount, feeDiff, receiveToken, receiveAmount, err
		}
		burnAmountFloat := float64(v.BurningAmount) / math.Pow10(tokenInfo.PDecimals)
		burnAmountStr := fmt.Sprintf("%f", burnAmountFloat)
		quoteDatas, err := estimateSwapFee(md.BurnTokenID.String(), receiveToken, burnAmountStr, int(v.ExternalNetworkID), spTkList, networkInfo, networkFees, tokenInfo, nil)
		if err != nil {
			log.Println("estimateSwapFee error", err)
			return result, callNetworkList, feeToken, feeAmount, feeDiff, receiveToken, receiveAmount, errors.New("can't validate fee at the moment, please try again later")
		}

		for _, quote := range quoteDatas {
			if strings.EqualFold(quote.CallContract, "0x"+v.ExternalCallAddress) {
				requireFeeToken = quote.Fee[0].TokenID
				requireFee = quote.Fee[0].Amount

				switch quote.AppName {
				case "curve":
				case "uniswap":
				case "pancake":

				}

				if feeToken != requireFeeToken {
					return result, callNetworkList, feeToken, feeAmount, feeDiff, receiveToken, receiveAmount, errors.New(fmt.Sprintf("invalid fee token, fee token can't be %v must be %v ", feeToken, requireFeeToken))
				}
				for _, fee := range quote.Fee {
					if fee.TokenID == feeToken {
						feeDiff = int64(feeAmount) - int64(fee.Amount)
						if feeDiff < 0 {
							feeDiffFloat := math.Abs(float64(feeDiff))
							diffPercent := feeDiffFloat / float64(fee.Amount) * 100
							if diffPercent > wcommon.PercentFeeDiff {
								return result, callNetworkList, feeToken, feeAmount, feeDiff, receiveToken, receiveAmount, errors.New("invalid fee amount, fee amount must be at least: " + fmt.Sprintf("%v", requireFee))
							}
						}
					}
				}
			}
		}
	}
	if requireFeeToken == "" {
		return result, callNetworkList, feeToken, feeAmount, feeDiff, receiveToken, receiveAmount, errors.New("invalid ExternalCallAddress")
	}

	result = true

	return result, callNetworkList, feeToken, feeAmount, feeDiff, receiveToken, receiveAmount, nil
}

func buildPancakeTokenMap(tokenList []PappSupportedTokenData) (map[string]PancakeTokenMapItem, error) {
	result := make(map[string]PancakeTokenMapItem)

	for _, token := range tokenList {
		if token.Protocol == "pancake" && token.Verify {
			contractID := strings.ToLower(token.ContractID)
			// if contractID == strings.ToLower("0xae13d989dac2f0debff460ac112a837c89baa7cd") || contractID == strings.ToLower("0x7ef95a0fee0dd31b22626fa2e10ee6a223f8a684") || contractID == strings.ToLower("0x8babbb98678facc7342735486c851abd7a0d17ca") || contractID == strings.ToLower("0x78867BbEeF44f2326bF8DDd1941a4439382EF2A7") || contractID == strings.ToLower("0x8a9424745056Eb399FD19a0EC26A14316684e274") || contractID == strings.ToLower("0xDAcbdeCc2992a63390d108e8507B98c7E2B5584a") {
			// 	result[contractID] = PancakeTokenMapItem{
			// 		Decimals: token.Decimals,
			// 		Symbol:   token.Symbol,
			// 	}
			// }

			result[contractID] = PancakeTokenMapItem{
				Decimals: token.Decimals,
				Symbol:   token.Symbol,
			}
		}
	}

	return result, nil
}

func cacheCurvePoolIndex() {
	for {
		var responseBodyData struct {
			Result []CurvePoolIndex
			Error  *struct {
				Code    int
				Message string
			} `json:"Error"`
		}
		_, err := restyClient.R().
			EnableTrace().
			SetHeader("Content-Type", "application/json").SetResult(&responseBodyData).
			Get(config.ShieldService + "/trade/supported-tokens")
		if err != nil {
			log.Println("cacheCurvePoolIndex", err.Error())
			continue
		}

		err = cacheStoreCustom(cacheCurvePoolIndexKey, responseBodyData, 30*time.Second)
		if err != nil {
			log.Println(err)
		}

		time.Sleep(60 * time.Second)
	}
}

func getCurvePoolIndex(endpoint string) ([]CurvePoolIndex, error) {
	var responseBodyData struct {
		Result []CurvePoolIndex
		Error  *struct {
			Code    int
			Message string
		} `json:"Error"`
	}
	re, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		Get(endpoint + "/trade/curve-pool-indices")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(re.Body(), &responseBodyData)
	if err != nil {
		return nil, err
	}

	if responseBodyData.Error != nil {
		return nil, errors.New(responseBodyData.Error.Message)
	}
	return responseBodyData.Result, nil

}

func getTokenCurvePoolIndex(contractID string, poolList []CurvePoolIndex) (int, string, error) {
	for _, v := range poolList {
		if v.DappTokenAddress == contractID {
			return v.CurveTokenIndex, v.CurvePoolAddress, nil
		}
	}
	return -1, "", errors.New("pool not found")
}

func extractDataFromRawTx(txraw []byte) (metadataCommon.Metadata, bool, []coin.Coin, string, error) {
	var mdRaw metadataCommon.Metadata
	var isPRVTx bool
	var outCoins []coin.Coin
	var txHash string

	// Unmarshal from json data to object tx))
	tx, err := transaction.DeserializeTransactionJSON(txraw)
	if err != nil {
		return mdRaw, isPRVTx, outCoins, txHash, err
	}
	if tx.TokenVersion2 != nil {
		isPRVTx = false
		txHash = tx.TokenVersion2.Hash().String()
		mdRaw = tx.TokenVersion2.GetMetadata()
		outCoins = append(outCoins, tx.TokenVersion2.Tx.Proof.GetOutputCoins()...)
		outCoins = append(outCoins, tx.TokenVersion2.TokenData.Proof.GetOutputCoins()...)
	}
	if tx.Version2 != nil {
		isPRVTx = true
		txHash = tx.TokenVersion2.Hash().String()
		mdRaw = tx.Version2.GetMetadata()
		outCoins = tx.Version2.GetProof().GetOutputCoins()
	}

	return mdRaw, isPRVTx, outCoins, txHash, nil
}

func getFee(isFeeWhitelist, isUnifiedNativeToken bool, nativeToken *PappSupportedTokenData, rate *big.Float, gasFee uint64, fromToken string, fromTokenInfo *wcommon.TokenInfo, pTokenContract1 *PappSupportedTokenData, toTokenDecimal *big.Float, additionalTokenInFee *big.Int) []PappNetworkFee {
	var fees []PappNetworkFee

	if isUnifiedNativeToken {
		feeAmount := ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", gasFee+additionalTokenInFee.Uint64()), int64(nativeToken.Decimals), int64(nativeToken.PDecimals))

		gasFeeFloat := new(big.Float).SetUint64(feeAmount)
		gasFeeFloat = gasFeeFloat.Mul(gasFeeFloat, rate)
		gasFeeFloat = gasFeeFloat.Mul(gasFeeFloat, toTokenDecimal)

		nativeDecimal := math.Pow10(-nativeToken.PDecimals)

		tdecimal, _ := toTokenDecimal.Float64()

		dcrate := new(big.Float).SetFloat64(nativeDecimal / tdecimal)
		gasFeeFloat = gasFeeFloat.Mul(gasFeeFloat, dcrate)

		gasFeeInt := gasFeeFloat.String()
		fees = append(fees, PappNetworkFee{
			Amount:           feeAmount,
			TokenID:          fromToken,
			AmountInBuyToken: gasFeeInt,
		})
	} else {

		gasFeePRV := float64(gasFee) * nativeToken.PricePrv
		gasFeeFromToken := (gasFeePRV / fromTokenInfo.PricePrv) + float64(additionalTokenInFee.Uint64())
		gasFeeFromTokenToPrv := gasFeeFromToken * fromTokenInfo.PricePrv

		feeAmount := ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", uint64(gasFeeFromToken)), int64(nativeToken.Decimals), int64(nativeToken.PDecimals))

		gasFeeFloat := new(big.Float).SetUint64(feeAmount)
		gasFeeFloat = gasFeeFloat.Mul(gasFeeFloat, rate)
		gasFeeFloat = gasFeeFloat.Mul(gasFeeFloat, toTokenDecimal)
		nativeDecimal := math.Pow10(-nativeToken.PDecimals)

		tdecimal, _ := toTokenDecimal.Float64()

		dcrate := new(big.Float).SetFloat64(nativeDecimal / tdecimal)
		gasFeeFloat = gasFeeFloat.Mul(gasFeeFloat, dcrate)
		gasFeeIntToToken := gasFeeFloat.String()
		if pTokenContract1.CurrencyType == wcommon.UnifiedCurrencyType || isFeeWhitelist {
			fees = append(fees, PappNetworkFee{
				Amount:           feeAmount,
				TokenID:          fromToken,
				AmountInBuyToken: gasFeeIntToToken,
			})
		} else {
			fees = append(fees, PappNetworkFee{
				Amount:           ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", uint64(gasFeeFromTokenToPrv)), int64(nativeToken.Decimals), int64(nativeToken.PDecimals)),
				TokenID:          common.PRVCoinID.String(),
				AmountInBuyToken: gasFeeIntToToken,
			})
		}
	}

	return fees
}

func APIRetrySwapTx(c *gin.Context) {

	type Request struct {
		TxList []string
	}

	var req Request
	err := c.MustBindWith(&req, binding.JSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

}

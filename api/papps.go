package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"math"
	"math/big"
	"strconv"

	"github.com/incognitochain/go-incognito-sdk-v2/coin"
	"github.com/incognitochain/go-incognito-sdk-v2/common/base58"
	"github.com/incognitochain/go-incognito-sdk-v2/crypto"
	"github.com/incognitochain/go-incognito-sdk-v2/key"
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
	"github.com/incognitochain/incognito-web-based-backend/interswap"
	"github.com/incognitochain/incognito-web-based-backend/papps"
	"github.com/incognitochain/incognito-web-based-backend/submitproof"
)

func APISubmitSwapTx(c *gin.Context) {
	var req SubmitSwapTxRequest
	userAgent := c.Request.UserAgent()
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

	spTkList := getSupportedTokenList()

	statusResult := checkPappTxSwapStatus(txHash, spTkList)
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

	valid, networkList, feeToken, feeAmount, pfeeAmount, feeDiff, swapInfo, err := checkValidTxSwap(md, outCoins, spTkList, wcommon.ExternalTxTypeSwap)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid tx err:" + err.Error()})
		return
	}
	// valid = true

	burntAmount, _ := md.TotalBurningAmount()
	if valid {
		status, err := submitproof.SubmitPappTx(txHash, []byte(req.TxRaw), isPRVTx, feeToken, feeAmount, pfeeAmount, md.BurnTokenID.String(), burntAmount, swapInfo, isUnifiedToken, networkList, req.FeeRefundOTA, req.FeeRefundAddress, userAgent, wcommon.ExternalTxTypeSwap)
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

	amount := new(big.Float)
	amount, errBool := amount.SetString(req.Amount)
	if !errBool {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid amount")})
		return
	}

	switch req.Network {
	case "inc", "pdex", "eth", "bsc", "plg", "ftm", "aurora", "avax":
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

	// estimate with Interswap
	if !req.IsFromInterswap {
		fmt.Println("Starting estimate interswap")
		interSwapParams := &interswap.EstimateSwapParam{
			Network:   req.Network,
			Amount:    req.Amount,
			Slippage:  req.Slippage,
			FromToken: req.FromToken,
			ToToken:   req.ToToken,
			ShardID:   req.ShardID,
		}

		interSwapRes, err := interswap.EstimateSwap(interSwapParams, config)
		if err != nil {
			result.NetworksError[interswap.InterSwapStr] = err.Error()
			fmt.Println("Estimate interswap with err", err)
		} else {
			for k, v := range interSwapRes {
				result.Networks[k] = v
			}
		}
	}
	fmt.Println("Finish estimate interswap")

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

	// Only estimate with pdex
	if req.Network == wcommon.NETWORK_PDEX {
		pdexresult := estimateSwapFeeWithPdex(req.FromToken, req.ToToken, req.Amount, slippage, tkFromInfo)
		if pdexresult != nil {
			pdexEstimate = append(pdexEstimate, *pdexresult)
		}
		if len(pdexEstimate) != 0 {
			result.Networks["inc"] = pdexEstimate
		}
		if len(result.Networks) == 0 && len(pdexEstimate) == 0 {
			response.Error = NotTradeable.Error()
		}

		response.Result = result
		c.PureJSON(200, response)
		return
	}

	if req.Network == "inc" {
		pdexresult := estimateSwapFeeWithPdex(req.FromToken, req.ToToken, req.Amount, slippage, tkFromInfo)
		if pdexresult != nil {
			pdexEstimate = append(pdexEstimate, *pdexresult)
		}
	}

	var resultLock sync.Mutex
	var wg sync.WaitGroup

	supportedNetworks := []int{}
	outofVaultNetworks := []int{}
	supportedOutNetworks := []int{}
	if tkFromInfo.CurrencyType == wcommon.UnifiedCurrencyType {
		dm := new(big.Float)
		dm.SetFloat64(math.Pow10(tkFromInfo.PDecimals))
		amountUint64, _ := amount.Mul(amount, dm).Uint64()

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
			c.JSON(http.StatusBadRequest, gin.H{"Error": "The amount exceeds the swap limit. Please retry with smaller amount."})
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
		}
		if len(result.Networks) > 0 {
			response.Result = result
			c.JSON(200, response)
			return
		}
		response.Error = NotTradeable.Error()
		c.JSON(http.StatusBadRequest, response)
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
		wg.Add(1)
		go func(net int) {
			data, err := estimateSwapFee(req.FromToken, req.ToToken, req.Amount, net, spTkList, networksInfo, networkFees, tkFromInfo, slippage)
			resultLock.Lock()
			if err != nil {
				networkErr[wcommon.GetNetworkName(net)] = err.Error()
			} else {
				result.Networks[wcommon.GetNetworkName(net)] = data
			}
			resultLock.Unlock()
			wg.Done()
		}(network)
	}
	wg.Wait()

	for net, v := range networkErr {
		result.NetworksError[net] = v
	}

	if len(pdexEstimate) != 0 {
		result.Networks["inc"] = pdexEstimate
	}
	if len(result.Networks) == 0 && len(pdexEstimate) == 0 {
		response.Error = NotTradeable.Error()
	}

	response.Result = result

	c.PureJSON(200, response)
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

	amountOutBigFloatPreSlippage := new(big.Float).SetFloat64(v.MaxGet)
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

	amountOutBigFloatPreSlippage = amountOutBigFloatPreSlippage.Mul(amountOutBigFloatPreSlippage, new(big.Float).SetFloat64(math.Pow10(-tkToInfo.PDecimals)))
	amountOutPreSlippage := amountOutBigFloatPreSlippage.String()

	amountInFloat, _ := new(big.Float).SetString(amount)
	rate := new(big.Float).Quo(amountOutBigFloatPreSlippage, new(big.Float).Set(amountInFloat))

	result := QuoteDataResp{
		AppName:              "pdex",
		AmountIn:             amount,
		AmountInRaw:          fmt.Sprintf("%v", amountRaw),
		AmountOut:            fmt.Sprintf("%v", amountOut),
		AmountOutRaw:         fmt.Sprintf("%f", pdexResult.MaxGet),
		AmountOutPreSlippage: amountOutPreSlippage,
		Rate:                 rate.Text('f', -1),
		Paths:                pdexResult.TokenRoute,
		PoolPairs:            pdexResult.Route,
		ImpactAmount:         fmt.Sprintf("%f", pdexResult.ImpactAmount),
		Fee:                  []PappNetworkFee{{TokenID: fromToken, Amount: pdexResult.Fee}},
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
		_, feeAddressShardID = common.GetShardIDsFromPublicKey(incFeeKeySet.KeySet.PaymentAddress.Pk)
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

	vaultData, err := database.DBGetPappVaultData(networkName, wcommon.ExternalTxTypeSwap)
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

	toTokenUnifiedID := ""
	toTokenChildID := ""

	if toTokenInfo.MovedUnifiedToken {
		toTokenChildID = toToken
		toTokenUnifiedID, err = getUnifiedTokenFromChildToken(toTokenChildID)
		if err != nil {
			return nil, err
		}
	} else {
		if toTokenInfo.CurrencyType == wcommon.UnifiedCurrencyType {
			toTokenUnifiedID = toToken
			for _, ctk := range toTokenInfo.ListUnifiedToken {
				netID, err := wcommon.GetNetworkIDFromCurrencyType(ctk.CurrencyType)
				if err != nil {
					return nil, err
				}
				if netID == networkID {
					toTokenChildID = ctk.TokenID
					break
				}
			}
		}
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
	isToTokenNative := false

	if pTokenContract2.CurrencyType == nativeCurrentType {
		isToTokenNative = true
	}
	if pTokenContract2.CurrencyType == wcommon.UnifiedCurrencyType {
		for _, v := range toTokenInfo.ListUnifiedToken {
			if v.CurrencyType == nativeCurrentType {
				isToTokenNative = true
			}
		}
	}
	if pTokenContract1.CurrencyType == nativeCurrentType {
		isUnifiedNativeToken = true
	}
	if pTokenContract1.CurrencyType == wcommon.UnifiedCurrencyType {
		for _, v := range fromTokenInfo.ListUnifiedToken {
			if v.CurrencyType == nativeCurrentType {
				isUnifiedNativeToken = true
				break
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

	amountInBig0 := new(big.Float).Set(amountFloat)

	additionalTokenInFee := amountInBig0.Mul(amountInBig0, new(big.Float).SetFloat64(0.003))

	toTokenDecimal := big.NewFloat(math.Pow10(-pTokenContract2.Decimals))
	isFeeWhitelist := false
	if _, ok := feeTokenWhiteList[fromToken]; ok {
		isFeeWhitelist = true
	}

	for appName, endpoint := range pappList.ExchangeApps {
		switch appName {
		case "uniswap":
			fmt.Println("uniswap", networkID, pTokenContract1.ContractID, pTokenContract2.ContractID)
			realAmountIn := new(big.Float).Set(amountFloat)
			// realAmountIn = realAmountIn.Mul(realAmountIn, new(big.Float).SetFloat64(0.997))
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

			// amountOutBigFloat0, _ := new(big.Float).SetString(quote.Data.AmountOut)
			// rate := new(big.Float).Quo(amountOutBigFloat0, new(big.Float).Set(amountFloat))
			// gasFee := (estGasUsed * gasPrice)

			// fees := getFee(isFeeWhitelist, isUnifiedNativeToken, nativeToken, rate, gasFee, fromToken, fromTokenInfo, pTokenContract1, toTokenDecimal, additionalTokenInFee)
			// if estGasUsed > wcommon.EVMGasLimitETH {
			// 	return nil, errors.New("estimated used gas exceed gas limit")
			// }
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
				if _, ok := traversedTk[route.TokenIn.Address]; !ok {
					paths = append(paths, tokenAddress)
				}
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

			amountOutBigFloatPreSlippage, _ := new(big.Float).SetString(quote.Data.AmountOutRaw)
			pTokenAmountPreSlippage := new(big.Float).Mul(amountOutBigFloatPreSlippage, toTokenDecimal)
			pTkAmountPreSlippageFloatStr := pTokenAmountPreSlippage.Text('f', -1)

			pathsList := []string{}
			for _, v := range paths {
				pathsList = append(pathsList, v.String())
			}

			outAmountInc := ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(amountOutBig.String(), int64(pTokenContract2.Decimals), int64(pTokenContract2.PDecimals))
			reDepositReward, _ := getShieldRewardEstimate(toTokenUnifiedID, toTokenChildID, outAmountInc)
			reDepositRewardBig := new(big.Float).SetUint64(reDepositReward)
			reDepositRewardStr := new(big.Float).Mul(reDepositRewardBig, toTokenDecimal).Text('f', -1)

			gasFee := (estGasUsed * gasPrice)
			outFloat, _ := pTokenAmountPreSlippage.Float64()
			inFloat, _ := amountFloat.Float64()
			rate := new(big.Float).SetFloat64(outFloat / inFloat)
			fees := getFee(isFeeWhitelist, isUnifiedNativeToken, nativeToken, rate, gasFee, fromToken, fromTokenInfo, pTokenContract1, toTokenDecimal, additionalTokenInFee, false)

			if amountOutBig.String() == "0" || amountOutBig.String() == "1" {
				return nil, errors.New("amount out is too small")
			}
			result = append(result, QuoteDataResp{
				AppName:              appName,
				AmountIn:             amount,
				AmountInRaw:          quote.Data.AmountIn,
				AmountOut:            pTkAmountFloatStr,
				AmountOutRaw:         amountOutBig.String(),
				AmountOutPreSlippage: pTkAmountPreSlippageFloatStr,
				RedepositReward:      reDepositRewardStr,
				Paths:                renderTokenFromTradePaths(pathsList, networkID),
				PathsContract:        pathsList,
				Rate:                 rate.Text('f', -1),
				Fee:                  fees,
				Calldata:             calldata,
				CallContract:         contract,
				FeeAddress:           feeAddress,
				FeeAddressShardID:    int(feeAddressShardID),
				RouteDebug:           quote.Data.Route,
			})
			log.Println("done estimate uniswap")
		case "pancake", "spooky", "joe", "trisolaris":
			fmt.Println(appName, networkID, pTokenContract1.ContractID, pTokenContract2.ContractID)
			realAmountIn := new(big.Float).Set(amountFloat)
			// if strings.Contains(config.NetworkID, "testnet") {
			// 	realAmountIn = realAmountIn.Mul(realAmountIn, new(big.Float).SetFloat64(0.998))
			// } else {
			// 	realAmountIn = realAmountIn.Mul(realAmountIn, new(big.Float).SetFloat64(0.9975))
			// }
			realAmountInFloat, _ := realAmountIn.Float64()
			realAmountInStr := fmt.Sprintf("%f", realAmountInFloat)

			tokenMap := make(map[string]PancakeTokenMapItem)

			switch appName {
			case "pancake":
				tokenMap, err = buildPancakeTokenMap(spTkList)
				if err != nil {
					log.Println(err)
					continue
				}
			case "spooky":
				tokenMap, err = buildSpookyTokenMap(spTkList)
				if err != nil {
					log.Println(err)
					continue
				}
			case "joe":
				tokenMap, err = buildJoeTokenMap(spTkList)
				if err != nil {
					log.Println(err)
					continue
				}
			case "trisolaris":
				tokenMap, err = buildTrisolarisTokenMap(spTkList)
				if err != nil {
					log.Println(err)
					continue
				}
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

			// amountOutBigFloat0, _ := new(big.Float).SetString(quote.Data.Outputs[len(quote.Data.Outputs)-1])
			// dcrate := new(big.Float).SetInt64(int64(math.Pow10(pTokenContract1.Decimals - pTokenContract2.Decimals)))
			// rate := new(big.Float).Quo(amountOutBigFloat0, new(big.Float).Set(amountBigFloat))
			// rate.Mul(rate, dcrate)

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

			calldata, err := papps.BuildCallDataPancake(paths, amountInBig, amountOutBig, isToTokenNative)
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

			amountOutBigFloatPreSlippage, _ := new(big.Float).SetString(quote.Data.Outputs[len(quote.Data.Outputs)-1])
			pTokenAmountPreSlippage := new(big.Float).Mul(amountOutBigFloatPreSlippage, toTokenDecimal)
			pTkAmountPreSlippageFloatStr := pTokenAmountPreSlippage.Text('f', -1)

			outAmountInc := ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(amountOutBig.String(), int64(pTokenContract2.Decimals), int64(pTokenContract2.PDecimals))
			reDepositReward, _ := getShieldRewardEstimate(toTokenUnifiedID, toTokenChildID, outAmountInc)
			reDepositRewardBig := new(big.Float).SetUint64(reDepositReward)
			reDepositRewardStr := new(big.Float).Mul(reDepositRewardBig, toTokenDecimal).Text('f', -1)

			gasFee := (estGasUsed * gasPrice)
			outFloat, _ := pTokenAmountPreSlippage.Float64()
			inFloat, _ := amountFloat.Float64()
			rate := new(big.Float).SetFloat64(outFloat / inFloat)
			fees := getFee(isFeeWhitelist, isUnifiedNativeToken, nativeToken, rate, gasFee, fromToken, fromTokenInfo, pTokenContract1, toTokenDecimal, additionalTokenInFee, false)

			log.Println("len(quote.Data.Outputs)", len(quote.Data.Outputs), quote.Data.Outputs, quote.Data.Outputs[len(quote.Data.Outputs)-1])

			contract, ok := pappList.AppContracts[appName]
			if !ok {
				return nil, errors.New("contract not found " + appName)
			}
			if amountOutBig.String() == "0" || amountOutBig.String() == "1" {
				return nil, errors.New("amount out is too small")
			}
			result = append(result, QuoteDataResp{
				AppName:              appName,
				AmountIn:             amount,
				AmountOut:            pTkAmountFloatStr,
				AmountOutRaw:         amountOutBig.String(),
				AmountOutPreSlippage: pTkAmountPreSlippageFloatStr,
				RedepositReward:      reDepositRewardStr,
				Rate:                 rate.Text('f', -1),
				Paths:                renderTokenFromTradePaths(quote.Data.Route, networkID),
				PathsContract:        quote.Data.Route,
				Fee:                  fees,
				Calldata:             calldata,
				CallContract:         contract,
				FeeAddress:           feeAddress,
				FeeAddressShardID:    int(feeAddressShardID),
				ImpactAmount:         fmt.Sprintf("%.2f", quote.Data.Impact),
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
			realAmountIn := new(big.Float).Set(amountBigFloat)
			// realAmountIn = realAmountIn.Mul(realAmountIn, new(big.Float).SetFloat64(0.9996))

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
			var pTkAmountPreSlippageFloatStr string
			var pTokenAmountPreSlippage *big.Float
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
					amountOutBigFloatPreSlippage := new(big.Float).SetInt(amountOut)
					pTokenAmountPreSlippage = new(big.Float).Mul(amountOutBigFloatPreSlippage, toTokenDecimal)
					pTkAmountPreSlippageFloatStr = pTokenAmountPreSlippage.Text('f', -1)
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

			if amountOut == nil {
				log.Println(errors.New("cant estimate curve"))
				continue
			}

			amountOutFloat := new(big.Float)
			amountOutFloat, _ = amountOutFloat.SetString(amountOut.String())
			pTokenAmount := new(big.Float).Mul(amountOutFloat, toTokenDecimal)

			estGasUsed := uint64(wcommon.EVMGasLimit)

			// amountOutBigFloat0, _ := new(big.Float).SetString(pTkAmountPreSlippageFloatStr)

			// dcrate := new(big.Float).SetInt64(int64(math.Pow10(pTokenContract1.Decimals - pTokenContract2.Decimals)))
			// amountInFloat, _ := new(big.Float).SetString(amount)
			// rate := new(big.Float).Quo(amountOutBigFloat0, amountInFloat)
			// gasFee := (estGasUsed * gasPrice)
			// fees := getFee(isFeeWhitelist, isUnifiedNativeToken, nativeToken, rate, gasFee, fromToken, fromTokenInfo, pTokenContract1, toTokenDecimal, additionalTokenInFee)
			contract, ok := pappList.AppContracts[appName]
			if !ok {
				return nil, errors.New("contract not found " + appName)
			}

			outAmountInc := ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(amountOut.String(), int64(pTokenContract2.Decimals), int64(pTokenContract2.PDecimals))
			reDepositReward, _ := getShieldRewardEstimate(toTokenUnifiedID, toTokenChildID, outAmountInc)
			reDepositRewardBig := new(big.Float).SetUint64(reDepositReward)
			reDepositRewardStr := new(big.Float).Mul(reDepositRewardBig, toTokenDecimal).Text('f', -1)

			gasFee := (estGasUsed * gasPrice)
			outFloat, _ := pTokenAmountPreSlippage.Float64()
			inFloat, _ := amountFloat.Float64()
			rate := new(big.Float).SetFloat64(outFloat / inFloat)
			fees := getFee(isFeeWhitelist, isUnifiedNativeToken, nativeToken, rate, gasFee, fromToken, fromTokenInfo, pTokenContract1, toTokenDecimal, additionalTokenInFee, false)
			if amountOut.String() == "0" || amountOut.String() == "1" {
				return nil, errors.New("amount out is too small")
			}
			result = append(result, QuoteDataResp{
				AppName:              appName,
				AmountIn:             amount,
				AmountOut:            pTokenAmount.String(),
				AmountOutRaw:         amountOut.String(),
				AmountOutPreSlippage: pTkAmountPreSlippageFloatStr,
				RedepositReward:      reDepositRewardStr,
				Rate:                 rate.Text('f', -1),
				Fee:                  fees,
				CallContract:         contract,
				Calldata:             calldata,
				FeeAddress:           feeAddress,
				Paths:                renderTokenFromTradePaths([]string{pTokenContract1.ContractID, pTokenContract2.ContractID}, networkID),
				PathsContract:        []string{pTokenContract1.ContractID, pTokenContract2.ContractID},
				FeeAddressShardID:    int(feeAddressShardID),
			})
			log.Println("done estimate curve")
		}
	}
	if len(result) == 0 {
		return nil, errors.New("not tradeable")
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

func checkReceiveOutCoinAmount(outCoins []coin.Coin, keySet *key.KeySet, burnToken string) (string, uint64, error) {
	var outToken string
	var outAmount uint64 = 0

	// burnTokenAssetTag := crypto.HashToPoint(md.BurnTokenID[:])
	for _, c := range outCoins {
		feeCoin, rK := c.DoesCoinBelongToKeySet(&incFeeKeySet.KeySet)
		if feeCoin {
			if c.GetAssetTag() == nil {
				outToken = common.PRVCoinID.String()
			} else {
				assetTag := c.GetAssetTag()
				blinder, err := coin.ComputeAssetTagBlinder(rK)
				if err != nil {
					return "", 0, err
				}
				rawAssetTag := new(crypto.Point).Sub(
					assetTag,
					new(crypto.Point).ScalarMult(crypto.PedCom.G[coin.PedersenRandomnessIndex], blinder),
				)
				_ = rawAssetTag //TODO: verify more

				outToken = burnToken
			}

			coin, err := c.Decrypt(&incFeeKeySet.KeySet)
			if err != nil {
				return "", 0, err
			}
			outAmount = coin.GetValue()
		}
	}

	return outToken, outAmount, nil
}

func checkValidTxSwap(md *bridge.BurnForCallRequest, outCoins []coin.Coin, spTkList []PappSupportedTokenData, pappType int) (bool, []string, string, uint64, uint64, int64, *wcommon.PappSwapInfo, error) {
	var feeAmount uint64
	var pfeeAmount uint64
	var feeToken string

	var requireFee uint64
	var requireFeeToken string

	var receiveToken string

	var result bool
	feeDiff := int64(-1)
	callNetworkList := []string{}
	networkInfo, err := getBridgeNetworkInfos()
	if err != nil {
		return result, callNetworkList, feeToken, feeAmount, pfeeAmount, feeDiff, nil, err
	}
	networkFees, err := database.DBRetrieveFeeTable()
	if err != nil {
		return result, callNetworkList, feeToken, feeAmount, pfeeAmount, feeDiff, nil, err
	}
	tokenInfo, err := getTokenInfo(md.BurnTokenID.String())
	if err != nil {
		return result, callNetworkList, feeToken, feeAmount, pfeeAmount, feeDiff, nil, err
	}

	// for _, c := range outCoins {
	// 	feeCoin, rK := c.DoesCoinBelongToKeySet(&incFeeKeySet.KeySet)
	// 	if feeCoin {
	// 		if c.GetAssetTag() == nil {
	// 			feeToken = common.PRVCoinID.String()
	// 		} else {
	// 			assetTag := c.GetAssetTag()
	// 			blinder, err := coin.ComputeAssetTagBlinder(rK)
	// 			if err != nil {
	// 				return result, callNetworkList, feeToken, feeAmount, pfeeAmount, feeDiff, nil, err
	// 			}
	// 			rawAssetTag := new(crypto.Point).Sub(
	// 				assetTag,
	// 				new(crypto.Point).ScalarMult(crypto.PedCom.G[coin.PedersenRandomnessIndex], blinder),
	// 			)
	// 			_ = rawAssetTag
	// 			feeToken = md.BurnTokenID.String()
	// 		}

	// 		coin, err := c.Decrypt(&incFeeKeySet.KeySet)
	// 		if err != nil {
	// 			return result, callNetworkList, feeToken, feeAmount, pfeeAmount, feeDiff, nil, err
	// 		}
	// 		feeAmount = coin.GetValue()
	// 	}
	// }

	feeToken, feeAmount, err = checkReceiveOutCoinAmount(outCoins, &incFeeKeySet.KeySet, md.BurnTokenID.String())
	if err != nil {
		return result, callNetworkList, feeToken, feeAmount, pfeeAmount, feeDiff, nil, err
	}

	if feeAmount == 0 {
		return result, callNetworkList, feeToken, feeAmount, pfeeAmount, feeDiff, nil, errors.New("you need to paid fee")
	}
	var swapInfo *wcommon.PappSwapInfo

	for _, v := range md.Data {
		callNetworkList = append(callNetworkList, wcommon.GetNetworkName(int(v.ExternalNetworkID)))
		if pappType == wcommon.ExternalTxTypeSwap {
			receiveToken, _, err = getTokenIDByContractID(v.ReceiveToken, int(v.ExternalNetworkID), spTkList, true)
			if err != nil {
				return result, callNetworkList, feeToken, feeAmount, pfeeAmount, feeDiff, nil, err
			}
		}
		burnAmountFloat := float64(v.BurningAmount) / math.Pow10(tokenInfo.PDecimals)
		burnAmountStr := fmt.Sprintf("%f", burnAmountFloat)
		var quoteDatas []QuoteDataResp
		if pappType == wcommon.ExternalTxTypeSwap {
			quoteDatas, err = estimateSwapFee(md.BurnTokenID.String(), receiveToken, burnAmountStr, int(v.ExternalNetworkID), spTkList, networkInfo, networkFees, tokenInfo, nil)
			if err != nil {
				log.Println("estimateSwapFee error", err)
				return result, callNetworkList, feeToken, feeAmount, pfeeAmount, feeDiff, nil, errors.New("can't validate fee at the moment, please try again later")
			}
		} else {
			//TODO: opensea
			openseaFee, err := estimateOpenSeaFee(v.BurningAmount, tokenInfo, callNetworkList[0], networkFees, spTkList)
			if err != nil {
				log.Println("estimateSwapFee error", err)
				return result, callNetworkList, feeToken, feeAmount, pfeeAmount, feeDiff, nil, errors.New("can't validate fee at the moment, please try again later")
			}
			pappList, err := database.DBRetrievePAppsByNetwork(callNetworkList[0])
			if err != nil {
				return result, callNetworkList, feeToken, feeAmount, pfeeAmount, feeDiff, nil, errors.New("can't validate fee at the moment, please try again later")

			}
			contract, exist := pappList.AppContracts["opensea"]
			if !exist {
				return result, callNetworkList, feeToken, feeAmount, pfeeAmount, feeDiff, nil, errors.New("can't validate fee at the moment, please try again later")

			}
			quoteDatas = append(quoteDatas, QuoteDataResp{
				AppName:      "opensea",
				CallContract: contract,
				Fee: []PappNetworkFee{{
					TokenID: openseaFee.TokenID,
					Amount:  openseaFee.Amount,
				}},
			})
		}

		for _, quote := range quoteDatas {
			if strings.EqualFold(quote.CallContract, "0x"+v.ExternalCallAddress) {
				requireFeeToken = quote.Fee[0].TokenID
				requireFee = quote.Fee[0].Amount
				dappSwapInfo := wcommon.PappSwapInfo{
					DappName: quote.AppName,
					TokenIn:  md.BurnTokenID.String(),
					TokenOut: receiveToken,
				}
				switch quote.AppName {
				case "opensea":
					//TODO: opensea
				case "curve":
					data, err := papps.DecodeCurveCalldata(v.ExternalCalldata)
					if err != nil {
						return result, callNetworkList, feeToken, feeAmount, pfeeAmount, feeDiff, nil, errors.New("can't decode curve calldata")
					}
					dappSwapInfo.MinOutAmount = data.MinAmount
					dappSwapInfo.TokenInAmount = data.Amount
				case "uniswap":
					data, err := papps.DecodeUniswapCalldata(v.ExternalCalldata)
					if err != nil {
						return result, callNetworkList, feeToken, feeAmount, pfeeAmount, feeDiff, nil, errors.New("can't decode uniswap calldata")
					}
					dappSwapInfo.MinOutAmount = data.AmountOutMinimum
					dappSwapInfo.TokenInAmount = data.AmountIn
				case "pancake", "spooky", "joe", "trisolaris":
					data, err := papps.DecodePancakeCalldata(v.ExternalCalldata)
					if err != nil {
						return result, callNetworkList, feeToken, feeAmount, pfeeAmount, feeDiff, nil, errors.New("can't decode pancake/spooky calldata")
					}
					dappSwapInfo.MinOutAmount = data.AmountOutMin
					dappSwapInfo.TokenInAmount = data.SrcQty
				}
				if feeToken != requireFeeToken {
					return result, callNetworkList, feeToken, feeAmount, pfeeAmount, feeDiff, nil, errors.New(fmt.Sprintf("invalid fee token, fee token can't be %v must be %v ", feeToken, requireFeeToken))
				}
				for _, fee := range quote.Fee {
					if fee.TokenID == feeToken {
						feeDiff = int64(feeAmount) - int64(fee.Amount)
						if feeDiff < 0 {
							feeDiffFloat := math.Abs(float64(feeDiff))
							diffPercent := feeDiffFloat / float64(fee.Amount) * 100
							if diffPercent > wcommon.PercentFeeDiff {
								return result, callNetworkList, feeToken, feeAmount, pfeeAmount, feeDiff, nil, errors.New("invalid fee amount, fee amount must be at least: " + fmt.Sprintf("%v", requireFee))
							}
						}
						pfeeAmount = fee.PrivacyFee
					}
				}
				swapInfo = &dappSwapInfo
				break
			}
		}
	}
	if requireFeeToken == "" {
		return result, callNetworkList, feeToken, feeAmount, pfeeAmount, feeDiff, nil, errors.New("invalid ExternalCallAddress")
	}
	// all pass
	result = true

	return result, callNetworkList, feeToken, feeAmount, pfeeAmount, feeDiff, swapInfo, nil
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

func buildSpookyTokenMap(tokenList []PappSupportedTokenData) (map[string]PancakeTokenMapItem, error) {
	result := make(map[string]PancakeTokenMapItem)

	for _, token := range tokenList {
		if (token.CurrencyType == wcommon.FTM || token.CurrencyType == wcommon.FTM_ERC20) && token.Verify {
			contractID := strings.ToLower(token.ContractID)

			result[contractID] = PancakeTokenMapItem{
				Decimals: token.Decimals,
				Symbol:   token.Symbol,
			}
		}
	}

	return result, nil
}

func buildJoeTokenMap(tokenList []PappSupportedTokenData) (map[string]PancakeTokenMapItem, error) {
	result := make(map[string]PancakeTokenMapItem)

	for _, token := range tokenList {
		if (token.CurrencyType == wcommon.AVAX || token.CurrencyType == wcommon.AVAX_ERC20) && token.Verify {
			contractID := strings.ToLower(token.ContractID)

			result[contractID] = PancakeTokenMapItem{
				Decimals: token.Decimals,
				Symbol:   token.Symbol,
			}
		}
	}

	return result, nil
}

func buildTrisolarisTokenMap(tokenList []PappSupportedTokenData) (map[string]PancakeTokenMapItem, error) {
	result := make(map[string]PancakeTokenMapItem)

	for _, token := range tokenList {
		if (token.CurrencyType == wcommon.AURORA_ETH || token.CurrencyType == wcommon.AURORA_ERC20) && token.Verify {
			contractID := strings.ToLower(token.ContractID)

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
		txHash = tx.Version2.Hash().String()
		mdRaw = tx.Version2.GetMetadata()
		outCoins = tx.Version2.GetProof().GetOutputCoins()
	}

	return mdRaw, isPRVTx, outCoins, txHash, nil
}

func getFee(isFeeWhitelist, isUnifiedNativeToken bool, nativeToken *wcommon.TokenInfo, rate *big.Float, gasFee uint64, fromToken string, fromTokenInfo *wcommon.TokenInfo, pTokenContract1 *PappSupportedTokenData, toTokenDecimal *big.Float, additionalTokenInFee *big.Float, isUnshield bool) []PappNetworkFee {
	var fees []PappNetworkFee

	// additionalTokenInFeeInUSD, _ := new(big.Float).Mul(additionalTokenInFee, new(big.Float).SetFloat64(fromTokenInfo.ExternalPriceUSD)).Uint64()
	// max_pfee := MAX_PFEE_PAPP
	// if isUnshield {
	// 	max_pfee = MAX_PFEE_UNSHIELD
	// }
	// if additionalTokenInFeeInUSD > max_pfee {
	// 	additionalTokenInFee = new(big.Float).SetFloat64(float64(max_pfee) / fromTokenInfo.ExternalPriceUSD)
	// }

	if isUnifiedNativeToken {
		// additionalTokenInFeeUint, _ := additionalTokenInFee.Uint64()
		additionalTokenInFeeUint, _ := new(big.Float).Mul(additionalTokenInFee, new(big.Float).SetFloat64(math.Pow10(int(nativeToken.Decimals)))).Uint64()
		feeAmount := ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", gasFee+additionalTokenInFeeUint), int64(nativeToken.Decimals), int64(nativeToken.PDecimals))
		gasFeeFloat := new(big.Float).SetUint64(feeAmount)
		gasFeeFloat = gasFeeFloat.Mul(gasFeeFloat, rate)
		gasFeeFloat = gasFeeFloat.Mul(gasFeeFloat, toTokenDecimal)

		nativeDecimal := math.Pow10(-nativeToken.PDecimals)

		tdecimal, _ := toTokenDecimal.Float64()

		dcrate := new(big.Float).SetFloat64(nativeDecimal / tdecimal)
		gasFeeFloat = gasFeeFloat.Mul(gasFeeFloat, dcrate)

		gasFeeInt := gasFeeFloat.String()

		feeInUSD := float64(feeAmount) / math.Pow10(nativeToken.PDecimals)
		feeInUSD = feeInUSD * fromTokenInfo.PriceUsd

		fees = append(fees, PappNetworkFee{
			Amount:           feeAmount,
			TokenID:          fromToken,
			AmountInBuyToken: gasFeeInt,
			PrivacyFee:       ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", additionalTokenInFeeUint), int64(nativeToken.Decimals), int64(nativeToken.PDecimals)),
			FeeInUSD:         feeInUSD,
		})
	} else {
		gasFeeUSD := float64(gasFee) * nativeToken.ExternalPriceUSD
		prvInfo, err := getTokenInfo(common.PRVCoinID.String())
		if err != nil {
			log.Println("getTokenInfo prv err:", err)
		}
		gasFeePRV := float64(gasFeeUSD) / prvInfo.PriceUsd
		gasFeeFromToken := (gasFeeUSD / fromTokenInfo.ExternalPriceUSD)

		feeAmount := ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", uint64(gasFeeFromToken)), int64(nativeToken.Decimals), int64(nativeToken.PDecimals))

		additionalTokenInFee1, _ := new(big.Float).Mul(additionalTokenInFee, new(big.Float).SetFloat64(math.Pow10(fromTokenInfo.PDecimals))).Uint64()

		feeAmount2 := feeAmount + additionalTokenInFee1

		additionalTokenInFeeFloat64, _ := additionalTokenInFee.Float64()

		additionalFeeInPRV := additionalTokenInFeeFloat64 * fromTokenInfo.PricePrv

		additionalTokenInFee2, _ := new(big.Float).Mul(new(big.Float).SetFloat64(additionalFeeInPRV), new(big.Float).SetFloat64(math.Pow10(9))).Float64()

		gasFeeFromTokenToPrv := ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", uint64(gasFeePRV)), int64(nativeToken.Decimals), int64(nativeToken.PDecimals))

		gasFeeFromTokenToPrv = gasFeeFromTokenToPrv + uint64(additionalTokenInFee2)

		gasFeeFloat := new(big.Float).SetUint64(feeAmount2)
		gasFeeFloat = gasFeeFloat.Mul(gasFeeFloat, rate)
		gasFeeFloat = gasFeeFloat.Mul(gasFeeFloat, toTokenDecimal)
		nativeDecimal := math.Pow10(-nativeToken.PDecimals)

		tdecimal, _ := toTokenDecimal.Float64()
		// fdecimal := math.Pow10(-fromTokenInfo.PDecimals)

		dcrate := new(big.Float).SetFloat64(nativeDecimal / tdecimal)

		dcrate2 := new(big.Float).SetFloat64(math.Pow10(fromTokenInfo.PDecimals - nativeToken.PDecimals))

		gasFeeFloatFromToken := new(big.Float).SetUint64(feeAmount2)
		gasFeeFloatFromToken = gasFeeFloatFromToken.Mul(gasFeeFloatFromToken, dcrate2)
		feeInFromToken, _ := gasFeeFloatFromToken.Uint64()

		gasFeeFloat = gasFeeFloat.Mul(gasFeeFloat, dcrate)
		gasFeeIntToToken := gasFeeFloat.String()

		if pTokenContract1.CurrencyType == wcommon.UnifiedCurrencyType || isFeeWhitelist {

			feeInUSD := float64(feeInFromToken) / math.Pow10(fromTokenInfo.PDecimals)
			feeInUSD = feeInUSD * fromTokenInfo.ExternalPriceUSD

			fees = append(fees, PappNetworkFee{
				Amount:           feeInFromToken,
				TokenID:          fromToken,
				AmountInBuyToken: gasFeeIntToToken,
				PrivacyFee:       additionalTokenInFee1,
				FeeInUSD:         feeInUSD,
			})
		} else {

			feeInUSD := float64(gasFeeFromTokenToPrv) / math.Pow10(9)
			feeInUSD = feeInUSD * prvInfo.PriceUsd

			fees = append(fees, PappNetworkFee{
				Amount:           gasFeeFromTokenToPrv,
				TokenID:          common.PRVCoinID.String(),
				AmountInBuyToken: gasFeeIntToToken,
				PrivacyFee:       uint64(additionalTokenInFee2),
				FeeInUSD:         feeInUSD,
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

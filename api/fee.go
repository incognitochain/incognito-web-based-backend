package api

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/papps"
)

func APIEstimateSwapFee(c *gin.Context) {
	var req EstimateSwapRequest
	err := c.MustBindWith(&req, binding.JSON)
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}

	networkID := 0
	nativeCurrentType := 0
	switch req.Network {
	case "eth":
		networkID = 1
		nativeCurrentType = common.NativeCurrencyTypeETH
	case "bsc":
		networkID = 2
		nativeCurrentType = common.NativeCurrencyTypeBSC
	case "plg":
		networkID = 3
		nativeCurrentType = common.NativeCurrencyTypePLG
	case "ftm":
		networkID = 4
		nativeCurrentType = common.NativeCurrencyTypeFTM
	}

	if networkID == 0 {
		c.JSON(200, gin.H{"Error": errors.New("unsupport network")})
		return
	}
	var result EstimateSwapRespond
	var response struct {
		Result interface{}
		Error  interface{}
	}

	supportedNetwork, err := database.DBGetBridgeNetworkInfo()
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}
	networkChainId := ""
	for _, v := range supportedNetwork {
		if req.Network == v.Network {
			networkChainId = v.ChainID
			break
		}
	}

	pappList, err := database.DBRetrievePAppsByNetwork(req.Network)
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}

	fmt.Println("APIEstimateSwapFee", req.Network, pappList)

	spTkList, err := getPappSupportedTokenList()
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}

	pTokenContract1, err := getpTokenContractID(req.FromToken, networkID, spTkList)
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}

	pTokenContract2, err := getpTokenContractID(req.ToToken, networkID, spTkList)
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}
	log.Println("UniswapQuote", pTokenContract1.ContractID, pTokenContract2.ContractID, fmt.Sprintf("%v", req.Amount), pappList.ExchangeApps["uniswap"])

	networkFees, err := database.DBRetrieveFeeTable()
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}

	if _, ok := networkFees.Fees[req.Network]; !ok {
		c.JSON(200, gin.H{"Error": "network gasPrice not found"})
		return
	}
	gasPrice := networkFees.Fees[req.Network]
	nativeToken, err := getNativeTokenData(spTkList, nativeCurrentType)
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}
	for appName, endpoint := range pappList.ExchangeApps {
		switch appName {
		case "uniswap":
			data, err := papps.UniswapQuote(pTokenContract1.ContractID, pTokenContract2.ContractID, req.Amount, networkChainId, true, endpoint)
			if err != nil {
				c.JSON(200, gin.H{"Error": err.Error()})
				return
			}
			quote, err := uniswapDataExtractor(data)
			if err != nil {
				c.JSON(200, gin.H{"Error": err.Error()})
				return
			}
			feeMap := make(map[string]uint64)

			estGasUsedStr := quote.Data.EstimatedGasUsed
			estGasUsed, err := strconv.ParseUint(estGasUsedStr, 10, 64)
			if err != nil {
				c.JSON(200, gin.H{"Error": err.Error()})
				return
			}
			feeMap[nativeToken.ID] = estGasUsed * gasPrice / uint64(math.Pow10(nativeToken.Decimals))
			result.Papps[appName] = QuoteDataResp{
				AmountIn:     fmt.Sprintf("%v", req.Amount),
				AmountOut:    quote.Data.AmountOut,
				AmountOutRaw: quote.Data.AmountOutRaw,
				Route:        quote.Data.Route,
				Fee:          feeMap,
			}
		}

	}

	response.Result = result

	// ConvertNanoIncogTokenToOutChainToken

	// feeInc := ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString
	c.JSON(200, response)

}

func APIEstimateReward(c *gin.Context) {
	var req EstimateRewardRequest
	err := c.MustBindWith(&req, binding.JSON)
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}

	reqRPC := genRPCBody("bridgeaggEstimateReward", []interface{}{
		map[string]interface{}{
			"UnifiedTokenID": req.UnifiedTokenID,
			"TokenID":        req.TokenID,
			"Amount":         req.Amount,
		},
	})

	var responseBodyData APIRespond
	_, err = restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).SetBody(reqRPC).
		Post(config.FullnodeURL)
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(200, responseBodyData)
}

func APIEstimateUnshield(c *gin.Context) {
	var req EstimateUnshieldRequest
	err := c.MustBindWith(&req, binding.JSON)
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}

	if req.ExpectedAmount > 0 && req.BurntAmount > 0 {
		c.JSON(200, gin.H{"Error": errors.New("either ExpectedAmount or BurntAmount can > 0, not both")})
		return
	}

	methodRPC := "bridgeaggEstimateFeeByExpectedAmount"
	if req.BurntAmount > 0 {
		methodRPC = "bridgeaggEstimateFeeByBurntAmount"
	}

	reqRPC := genRPCBody(methodRPC, []interface{}{
		map[string]interface{}{
			"UnifiedTokenID": req.UnifiedTokenID,
			"TokenID":        req.TokenID,
			"ExpectedAmount": req.ExpectedAmount,
			"BurntAmount":    req.BurntAmount,
		},
	})

	var responseBodyData struct {
		Result interface{}
		Error  interface{}
	}
	_, err = restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).SetBody(reqRPC).
		Post(config.FullnodeURL)
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(200, responseBodyData)
}

func getQuote() {

}

// https://api.uniswap.org/v1/quote?protocols=v2,v3&tokenInAddress=0x9c3C9283D3e44854697Cd22D3Faa240Cfb032889&tokenInChainId=80001&tokenOutAddress=0xA6FA4fB5f76172d178d61B04b0ecd319C5d1C0aa&tokenOutChainId=80001&amount=1000000000000000000&type=exactIn

// 0xa6fa4fb5f76172d178d61b04b0ecd319c5d1c0aa

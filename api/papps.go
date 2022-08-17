package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"math"
	"math/big"
	"strconv"

	"github.com/incognitochain/go-incognito-sdk-v2/common/base58"
	"github.com/incognitochain/go-incognito-sdk-v2/metadata"
	"github.com/incognitochain/go-incognito-sdk-v2/metadata/bridge"
	"github.com/incognitochain/go-incognito-sdk-v2/transaction"
	"github.com/mongodb/mongo-tools/common/json"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/incognitochain/go-incognito-sdk-v2/common"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/papps"
)

func APISubmitSwapTx(c *gin.Context) {
	var req SubmitSwapTxRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if req.TxRaw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid txhash")})
		return
	}
	rawTxBytes, _, err := base58.Base58Check{}.Decode(req.TxRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid txhash")})
		return
	}

	// Unmarshal from json data to object tx))
	tx, err := transaction.DeserializeTransactionJSON(rawTxBytes)
	// var tx transaction.Tx
	// err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		tx, err = transaction.DeserializeTransactionJSON(rawTxBytes)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
	}
	if tx.TokenVersion2 != nil {
		if tx.TokenVersion2.GetMetadataType() != metadata.BurnForCallRequestMeta {
			md := tx.TokenVersion2.GetMetadata().(*bridge.BurnForCallRequest)
			_ = md
		}
	}
	if tx.Version2 != nil {
		if tx.Version2.GetMetadataType() != metadata.BurnForCallRequestMeta {
			md := tx.Version2.GetMetadata().(*bridge.BurnForCallRequest)
			_ = md
		}
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
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(200, responseBodyData)
}

func APIGetSupportedToken(c *gin.Context) {
	pappTokens, err := getPappSupportedTokenList()
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}

	tokenList, err := retrieveTokenList()
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}
	var result []wcommon.TokenInfo
	dupChecker := make(map[string]struct{})

	for _, tk := range pappTokens {
		for _, v := range tokenList {
			if tk.ID == v.TokenID {
				if _, exist := dupChecker[v.TokenID]; !exist {
					result = append(result, v)
					dupChecker[v.TokenID] = struct{}{}
				}
			}
		}
	}

	var response struct {
		Result interface{}
		Error  interface{}
	}
	response.Result = result

	c.JSON(200, response)
}

func APIEstimateSwapFee(c *gin.Context) {
	var req EstimateSwapRequest
	err := c.MustBindWith(&req, binding.JSON)
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}
	switch req.Network {
	case "inc", "eth", "bsc", "plg", "ftm":
	default:
		c.JSON(200, gin.H{"Error": errors.New("unsupport network")})
		return
	}

	networkID := wcommon.GetNetworkID(req.Network)

	if networkID == -1 {
		c.JSON(200, gin.H{"error": "invalid network"})
		return
	}

	var result EstimateSwapRespond
	result.Networks = make(map[string]interface{})

	tkFromInfo, err := getTokenInfo(req.FromToken)
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}

	tkToInfo, err := getTokenInfo(req.ToToken)
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}
	tkToNetworkID := 0
	if tkToInfo.CurrencyType != wcommon.UnifiedCurrencyType {
		tkToNetworkID, err = getNetworkIDFromCurrencyType(tkToInfo.CurrencyType)
		if err != nil {
			c.JSON(200, gin.H{"Error": err.Error()})
			return
		}
	}

	supportedNetworks := []int{}

	if tkFromInfo.CurrencyType == wcommon.UnifiedCurrencyType {
		amount := new(big.Float)
		amount, errBool := amount.SetString(req.Amount)
		if !errBool {
			c.JSON(200, gin.H{"Error": errors.New("invalid amount")})
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
					c.JSON(200, gin.H{"Error": err.Error()})
					return
				}
				if isEnoughVault {
					supportedOutNetworks = append(supportedOutNetworks, v.NetworkID)
				}
			} else {
				//check 1 vault only
				if networkID == v.NetworkID {
					isEnoughVault, err := checkEnoughVault(req.FromToken, v.TokenID, amountUint64)
					if err != nil {
						c.JSON(200, gin.H{"Error": err.Error()})
						return
					}
					if isEnoughVault {
						supportedOutNetworks = append(supportedOutNetworks, v.NetworkID)
					}
				}
			}
		}
		if len(supportedOutNetworks) == 0 {
			c.JSON(200, gin.H{"Error": "The amount exceeds the swap limit. Please retry with another token or switch to other pApps"})
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
		tkFromNetworkID, err := getNetworkIDFromCurrencyType(tkFromInfo.CurrencyType)
		if err != nil {
			c.JSON(200, gin.H{"Error": err.Error()})
			return
		}
		if networkID == tkFromNetworkID {
			supportedOutNetworks = append(supportedOutNetworks, tkFromNetworkID)
		} else {
			if networkID == 0 {
				supportedOutNetworks = append(supportedOutNetworks, tkFromNetworkID)
			} else {
				c.JSON(200, gin.H{"Error": "No supported networks found"})
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
	if len(supportedNetworks) == 0 {
		c.JSON(200, gin.H{"Error": "No compatible network found"})
		return
	}

	networksInfo, err := database.DBGetBridgeNetworkInfos()
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}

	spTkList, err := getPappSupportedTokenList()
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}

	networkFees, err := database.DBRetrieveFeeTable()
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}

	for _, network := range supportedNetworks {
		data, err := estimateSwapFee(req.FromToken, req.ToToken, req.Amount, network, spTkList, networksInfo, networkFees, tkFromInfo)
		if err != nil {
			result.Networks[wcommon.GetNetworkName(network)] = err.Error()
		} else {
			result.Networks[wcommon.GetNetworkName(network)] = data
		}
	}

	var response struct {
		Result interface{}
		Error  interface{}
	}
	response.Result = result

	c.JSON(200, response)

}

func estimateSwapFee(fromToken, toToken, amount string, networkID int, spTkList []PappSupportedTokenData, networksInfo []wcommon.BridgeNetworkData, networkFees *wcommon.ExternalNetworksFeeData, fromTokenInfo *wcommon.TokenInfo) (interface{}, error) {
	result := []QuoteDataResp{}

	log.Println("estimateSwapFee for", fromToken, toToken, amount, networkID)
	networkName := wcommon.GetNetworkName(networkID)
	pappList, err := database.DBRetrievePAppsByNetwork(networkName)
	if err != nil {
		fmt.Println("DBRetrievePAppsByNetwork", err)
		return nil, err
	}

	pTokenContract1, err := getpTokenContractID(fromToken, networkID, spTkList)
	if err != nil {
		log.Println("err get pTokenContract1")
		return nil, err
	}

	pTokenContract2, err := getpTokenContractID(toToken, networkID, spTkList)
	if err != nil {
		log.Println("err get pTokenContract2")
		return nil, err
	}

	log.Println("done get pTokenContract1")
	networkChainId := ""
	for _, v := range networksInfo {
		if networkName == v.Network {
			networkChainId = v.ChainID
			break
		}
	}

	if _, ok := networkFees.Fees[networkName]; !ok {
		return nil, errors.New("network gasPrice not found")
	}
	gasPrice := networkFees.Fees[networkName]

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

	for appName, endpoint := range pappList.ExchangeApps {
		switch appName {
		case "uniswap":
			fmt.Println("uniswap", networkID, pTokenContract1.ContractID, pTokenContract2.ContractID)
			data, err := papps.UniswapQuote(pTokenContract1.ContractID, pTokenContract2.ContractID, amount, networkChainId, true, endpoint)
			if err != nil {
				return nil, err
			}
			quote, err := uniswapDataExtractor(data)
			if err != nil {
				return nil, err
			}
			fees := []PappNetworkFee{}

			estGasUsedStr := quote.Data.EstimatedGasUsed
			estGasUsed, err := strconv.ParseUint(estGasUsedStr, 10, 64)
			if err != nil {
				return nil, err
			}
			estGasUsed += 200000
			if isUnifiedNativeToken {
				fees = append(fees, PappNetworkFee{
					Amount:     ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", estGasUsed*gasPrice), int64(nativeToken.Decimals), int64(nativeToken.PDecimals)),
					FeeAddress: nativeToken.ID,
				})
			} else {
				if pTokenContract1.CurrencyType == wcommon.UnifiedCurrencyType || pTokenContract1.MovedUnifiedToken {
					fees = append(fees, PappNetworkFee{
						Amount:     ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", uint64(float64(estGasUsed*gasPrice)*nativeToken.PricePrv/pTokenContract1.PricePrv)), int64(nativeToken.Decimals), int64(nativeToken.PDecimals)),
						FeeAddress: pTokenContract1.ID,
					})
				} else {
					fees = append(fees, PappNetworkFee{
						Amount:     ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", uint64(float64(estGasUsed*gasPrice)*nativeToken.PricePrv)), int64(nativeToken.Decimals), int64(nativeToken.PDecimals)),
						FeeAddress: common.PRVCoinID.String(),
					})
				}

			}

			result = append(result, QuoteDataResp{
				AppName:      appName,
				AmountIn:     amount,
				AmountOut:    quote.Data.AmountOut,
				AmountOutRaw: quote.Data.AmountOutRaw,
				Route:        quote.Data.Route,
				Fee:          fees,
			})
		case "pancake":
			fmt.Println("pancake", networkID, pTokenContract1.ContractID, pTokenContract2.ContractID)
			data, err := papps.PancakeQuote(pTokenContract1.ContractID, pTokenContract2.ContractID, amount, networkChainId, pTokenContract1.Symbol, pTokenContract2.Symbol, pTokenContract1.Decimals, pTokenContract2.Decimals, true, endpoint)
			if err != nil {
				return nil, err
			}
			quote, err := pancakeDataExtractor(data)
			if err != nil {
				return nil, err
			}
			fees := []PappNetworkFee{}
			// estGasUsedStr := quote.Data.EstimatedGasUsed
			// estGasUsed, err := strconv.ParseUint(estGasUsedStr, 10, 64)
			// if err != nil {
			// 	return nil, err
			// }
			estGasUsed := uint64(600000)
			if isUnifiedNativeToken {
				fees = append(fees, PappNetworkFee{
					Amount:     ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", estGasUsed*gasPrice), int64(nativeToken.Decimals), int64(nativeToken.PDecimals)),
					FeeAddress: nativeToken.ID,
				})
			} else {
				if pTokenContract1.CurrencyType == wcommon.UnifiedCurrencyType || pTokenContract1.MovedUnifiedToken {
					fees = append(fees, PappNetworkFee{
						Amount:     ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", uint64(float64(estGasUsed*gasPrice)*nativeToken.PricePrv/pTokenContract1.PricePrv)), int64(nativeToken.Decimals), int64(nativeToken.PDecimals)),
						FeeAddress: pTokenContract1.ID,
					})
				} else {
					fees = append(fees, PappNetworkFee{
						Amount:     ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", uint64(float64(estGasUsed*gasPrice)*nativeToken.PricePrv)), int64(nativeToken.Decimals), int64(nativeToken.PDecimals)),
						FeeAddress: common.PRVCoinID.String(),
					})
				}
			}

			amountOut := ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(quote.Data.Outputs[1], int64(pTokenContract2.Decimals), int64(pTokenContract2.PDecimals))
			result = append(result, QuoteDataResp{
				AppName:      appName,
				AmountIn:     amount,
				AmountOut:    fmt.Sprintf("%v", amountOut),
				AmountOutRaw: quote.Data.Outputs[1],
				Route:        quote.Data.Route,
				Fee:          fees,
			})
		case "curve":
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
			return
		}

		err = cacheStoreCustom(cacheVaultStateKey, responseBodyData, 30*time.Second)
		if err != nil {
			log.Println(err)
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
			return
		}

		err = cacheStoreCustom(cacheSupportedPappsTokensKey, responseBodyData, 30*time.Second)
		if err != nil {
			log.Println(err)
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
			return
		}
		err = cacheStoreCustom(cacheTokenListKey, responseBodyData, 30*time.Second)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(10 * time.Second)
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

func getPappSupportedTokenList() ([]PappSupportedTokenData, error) {

	var responseBodyData struct {
		Result []PappSupportedTokenData
		Error  *struct {
			Code    int
			Message string
		} `json:"Error"`
	}

	err := cacheGet(cacheSupportedPappsTokensKey, &responseBodyData)
	if err != nil {
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
		return responseBodyData.Result, nil
	}

	return responseBodyData.Result, nil
}

func checkValidTxSwap(md *bridge.BurnForCallRequest) {}

func sendSwapTxAndStoreDB(txhash string, txRaw string, isTokenTx bool) error {
	if isTokenTx {
		err := incClient.SendRawTokenTx([]byte(txRaw))
		if err != nil {
			return err
		}
	} else {
		err := incClient.SendRawTx([]byte(txRaw))
		if err != nil {
			return err
		}
	}
	return nil
}

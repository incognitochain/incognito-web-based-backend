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
		c.JSON(400, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(200, responseBodyData)
}

func APIGetSupportedToken(c *gin.Context) {
	var responseBodyData APIRespond
	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).
		Get(config.ShieldService + "/trade/supported-tokens")
	if err != nil {
		c.JSON(400, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(200, responseBodyData)
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

	var result EstimateSwapRespond
	result.Networks = make(map[string]map[string]QuoteDataResp)

	tkInfo, err := getTokenInfo(req.FromToken)
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}

	supportedNetworks := []int{}

	if tkInfo.CurrencyType == wcommon.UnifiedCurrencyType {
		amount := new(big.Float)
		amount, errBool := amount.SetString(req.Amount)
		if !errBool {
			c.JSON(200, gin.H{"Error": errors.New("invalid amount")})
			return
		}
		dm := new(big.Float)
		dm.SetFloat64(math.Pow10(tkInfo.PDecimals))
		amountUint64, _ := amount.Mul(amount, dm).Uint64()
		tkInfo, err := getTokenInfo(req.FromToken)
		if err != nil {
			c.JSON(200, gin.H{"Error": err.Error()})
			return
		}
		for _, v := range tkInfo.ListUnifiedToken {
			if networkID == 0 {
				isEnoughVault, err := checkEnoughVault(req.FromToken, v.TokenID, amountUint64)
				if err != nil {
					c.JSON(200, gin.H{"Error": err.Error()})
					return
				}
				if isEnoughVault {
					supportedNetworks = append(supportedNetworks, v.NetworkID)
				}
			} else {
				if networkID == v.NetworkID {
					isEnoughVault, err := checkEnoughVault(req.FromToken, v.TokenID, amountUint64)
					if err != nil {
						c.JSON(200, gin.H{"Error": err.Error()})
						return
					}
					if isEnoughVault {
						supportedNetworks = append(supportedNetworks, v.NetworkID)
					}
				}
			}
		}
		if len(supportedNetworks) == 0 {
			c.JSON(200, gin.H{"Error": "No supported networks found"})
			return
		}
	} else {
		if networkID == tkInfo.NetworkID {
			supportedNetworks = append(supportedNetworks, tkInfo.NetworkID)
		} else {
			c.JSON(200, gin.H{"Error": "No supported networks found"})
			return
		}
	}

	fmt.Println("pass check vault")
	networksInfo, err := database.DBGetBridgeNetworkInfo()
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
		data, err := estimateSwapFee(req.FromToken, req.ToToken, req.Amount, network, spTkList, networksInfo, networkFees)
		if err != nil {
			result.Networks[wcommon.GetNetworkName(network)] = nil
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

func estimateSwapFee(fromToken, toToken, amount string, networkID int, spTkList []PappSupportedTokenData, networksInfo []wcommon.BridgeNetworkData, networkFees *wcommon.ExternalNetworksFeeData) (map[string]QuoteDataResp, error) {
	result := make(map[string]QuoteDataResp)

	networkName := wcommon.GetNetworkName(networkID)
	pappList, err := database.DBRetrievePAppsByNetwork(networkName)
	if err != nil {
		return nil, err
	}

	pTokenContract1, err := getpTokenContractID(fromToken, networkID, spTkList)
	if err != nil {
		return nil, err
	}

	pTokenContract2, err := getpTokenContractID(toToken, networkID, spTkList)
	if err != nil {
		return nil, err
	}

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

	for appName, endpoint := range pappList.ExchangeApps {
		switch appName {
		case "uniswap":
			data, err := papps.UniswapQuote(pTokenContract1.ContractID, pTokenContract2.ContractID, amount, networkChainId, true, endpoint)
			if err != nil {
				return nil, err
			}
			quote, err := uniswapDataExtractor(data)
			if err != nil {
				return nil, err
			}
			feeMap := make(map[string]uint64)

			estGasUsedStr := quote.Data.EstimatedGasUsed
			estGasUsed, err := strconv.ParseUint(estGasUsedStr, 10, 64)
			if err != nil {
				return nil, err
			}
			estGasUsed += 200000
			if pTokenContract1.CurrencyType == nativeCurrentType {
				feeMap[nativeToken.ID] = ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(estGasUsed*gasPrice, int64(nativeToken.Decimals), int64(nativeToken.PDecimals))
			} else {
				feeMap[common.PRVCoinID.String()] = ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(uint64(float64(estGasUsed*gasPrice)*pTokenContract1.PricePrv), int64(nativeToken.Decimals), int64(nativeToken.PDecimals))
			}
			feeMap["estGasUsed"] = estGasUsed
			feeMap["gasPrice"] = gasPrice
			feeMap["nativeToken.Decimals"] = uint64(nativeToken.Decimals)
			feeMap["nativeToken.PDecimals"] = uint64(nativeToken.PDecimals)

			result[appName] = QuoteDataResp{
				AmountIn:     amount,
				AmountOut:    quote.Data.AmountOut,
				AmountOutRaw: quote.Data.AmountOutRaw,
				Route:        quote.Data.Route,
				Fee:          feeMap,
			}
		case "curve":
		case "pancake":
		}
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

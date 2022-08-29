package api

import (
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
	"github.com/incognitochain/go-incognito-sdk-v2/metadata/bridge"
	metadataCommon "github.com/incognitochain/go-incognito-sdk-v2/metadata/common"
	"github.com/incognitochain/go-incognito-sdk-v2/transaction"
	"github.com/mongodb/mongo-tools/common/json"
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
		c.JSON(http.StatusOK, gin.H{"Error": err.Error()})
		return
	}

	var md *bridge.BurnForCallRequest
	var mdRaw metadataCommon.Metadata
	var isPRVTx bool
	var isUnifiedToken bool
	var feeToken string
	var feeAmount uint64
	var outCoins []coin.Coin
	var txHash string
	var rawTxBytes []byte

	var ok bool
	if req.TxHash != "" {
		txHash = req.TxHash
		txDetail, err := incClient.GetTx(req.TxHash)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"Error": err.Error()})
			return
		}
		mdRaw = txDetail.GetMetadata()
		txType := txDetail.GetType()
		switch txType {
		case common.TxCustomTokenPrivacyType:
			isPRVTx = false
		case common.TxNormalType:
			isPRVTx = true
			feeToken = common.PRVCoinID.String()
		}
	} else {
		rawTxBytes, _, err = base58.Base58Check{}.Decode(req.TxRaw)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"Error": errors.New("invalid txhash")})
			return
		}

		// Unmarshal from json data to object tx))
		tx, err := transaction.DeserializeTransactionJSON(rawTxBytes)
		// var tx transaction.Tx
		// err = json.Unmarshal(rawTxBytes, &tx)
		if err != nil {
			tx, err = transaction.DeserializeTransactionJSON(rawTxBytes)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"Error": err.Error()})
				return
			}
		}
		if tx.TokenVersion2 != nil {
			isPRVTx = false
			txHash = tx.TokenVersion2.Hash().String()
			mdRaw = tx.TokenVersion2.GetMetadata()
			outCoins = append(outCoins, tx.TokenVersion2.GetTxTokenData().TxNormal.GetProof().GetOutputCoins()...)
			outCoins = append(outCoins, tx.TokenVersion2.Tx.Proof.GetOutputCoins()...)
		}
		if tx.Version2 != nil {
			isPRVTx = true
			txHash = tx.TokenVersion2.Hash().String()
			feeToken = common.PRVCoinID.String()
			mdRaw = tx.Version2.GetMetadata()
			outCoins = tx.Version2.GetProof().GetOutputCoins()
		}

	}

	if mdRaw == nil {
		c.JSON(http.StatusOK, gin.H{"Error": errors.New("invalid tx metadata type")})
		return
	}
	md, ok = mdRaw.(*bridge.BurnForCallRequest)
	if !ok {
		c.JSON(http.StatusOK, gin.H{"Error": errors.New("invalid tx metadata type")})
		return
	}

	statusResult, err := checkPappTxSwapStatus(txHash)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			c.JSON(200, gin.H{"Error": err.Error()})
			return
		}
	}
	if len(statusResult) > 0 {
		c.JSON(200, gin.H{"Result": statusResult})
		return
	}

	burnTokenInfo, err := getTokenInfo(md.BurnTokenID.String())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"Error": errors.New("invalid tx metadata type")})
		return
	}
	if burnTokenInfo.CurrencyType == wcommon.UnifiedCurrencyType {
		isUnifiedToken = true
	}

	valid, networkList, err := checkValidTxSwap(md, feeToken, feeAmount, outCoins)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"Error": err.Error()})
		return
	}
	if valid {
		status, err := submitproof.SubmitPappTx(txHash, rawTxBytes, isPRVTx, feeToken, feeAmount, isUnifiedToken, networkList)
		if err != nil {
			c.JSON(200, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"Result": status})
		return
	}

	c.JSON(200, gin.H{"Error": "invalid tx swap"})
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

	networksInfo, err := getBridgeNetworkInfos()
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

func estimateSwapFee(fromToken, toToken, amount string, networkID int, spTkList []PappSupportedTokenData, networksInfo []wcommon.BridgeNetworkData, networkFees *wcommon.ExternalNetworksFeeData, fromTokenInfo *wcommon.TokenInfo) ([]QuoteDataResp, error) {
	result := []QuoteDataResp{}

	log.Println("estimateSwapFee for", fromToken, toToken, amount, networkID)
	networkName := wcommon.GetNetworkName(networkID)
	pappList, err := database.DBRetrievePAppsByNetwork(networkName)
	if err != nil {
		fmt.Println("DBRetrievePAppsByNetwork", err)
		return nil, err
	}

	vaultData, err := database.DBGetPappContractData(networkName, wcommon.PappTypeSwap)
	if err != nil {
		fmt.Println("DBGetPappContractData", err)
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
			fees = append(fees, PappNetworkFee{
				Amount:     estGasUsed,
				FeeAddress: "estGasUsed",
			})
			fees = append(fees, PappNetworkFee{
				Amount:     gasPrice,
				FeeAddress: "gasPrice",
			})

			vaultAddress := ethcommon.Address{}
			err = vaultAddress.UnmarshalText([]byte(vaultData.ContractAddress))
			if err != nil {
				return nil, err
			}

			amountFloat := new(big.Float)
			amountFloat, ok := amountFloat.SetString(amount)
			if !ok {
				return nil, fmt.Errorf("amount is not a number")
			}
			amountBigFloat := ConvertToNanoIncognitoToken(amountFloat, int64(pTokenContract1.Decimals)) //amount *big.Float, decimal int64, return *big.Float
			log.Println("amountBigFloat: ", amountBigFloat.String())

			// convert float to bigin:
			amountInBig, _ := amountBigFloat.Int(nil)

			amountOutBig, _ := new(big.Int).SetString(quote.Data.AmountOutRaw, 10)

			paths := []ethcommon.Address{}

			for _, route := range quote.Data.Route[0] {
				tokenAddress := ethcommon.Address{}
				err = tokenAddress.UnmarshalText([]byte(route.TokenIn.Address))
				if err != nil {
					return nil, err
				}
				paths = append(paths, tokenAddress)
			}

			tokenOutAddress := ethcommon.Address{}
			err = tokenOutAddress.UnmarshalText([]byte(pTokenContract2.ContractID))
			if err != nil {
				return nil, err
			}
			paths = append(paths, tokenOutAddress)

			calldata, err := papps.BuildCallDataUniswap(paths, vaultAddress, []int64{int64(estGasUsed)}, amountInBig, amountOutBig, isUnifiedNativeToken)
			if err != nil {
				log.Println("Error building call data: ", err)
				calldata = err.Error()
			}

			result = append(result, QuoteDataResp{
				AppName:      appName,
				AmountIn:     amount,
				AmountInRaw:  quote.Data.AmountIn,
				AmountOut:    quote.Data.AmountOut,
				AmountOutRaw: quote.Data.AmountOutRaw,
				Route:        quote.Data.Route,
				Fee:          fees,
				Calldata:     calldata,
			})
		case "pancake":
			fmt.Println("pancake", networkID, pTokenContract1.ContractID, pTokenContract2.ContractID)

			tokenMap, err := buildPancakeTokenMap()
			if err != nil {
				return nil, err
			}

			tokenMapBytes, err := json.Marshal(tokenMap)
			if err != nil {
				return nil, err
			}

			log.Println("tokenMapBytes", string(tokenMapBytes))
			data, err := papps.PancakeQuote(pTokenContract1.ContractID, pTokenContract2.ContractID, amount, networkChainId, pTokenContract1.Symbol, pTokenContract2.Symbol, pTokenContract1.Decimals, pTokenContract2.Decimals, false, endpoint, string(tokenMapBytes))
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

			amountOut, _ := new(big.Float).SetString(quote.Data.Outputs[1])

			pTokenAmount := new(big.Float).Mul(amountOut, big.NewFloat(math.Pow10(-pTokenContract2.Decimals)))

			result = append(result, QuoteDataResp{
				AppName:      appName,
				AmountIn:     amount,
				AmountOut:    pTokenAmount.String(),
				AmountOutRaw: quote.Data.Outputs[1],
				Route:        quote.Data.Route,
				Fee:          fees,
			})
		case "curve":
			poolList, err := getCurvePoolIndex()
			if err != nil {
				return nil, err
			}
			token1PoolIndex, curvePoolAddress1, err := getTokenCurvePoolIndex(pTokenContract1.ContractIDGetRate, poolList)
			if err != nil {
				return nil, err
			}
			token2PoolIndex, _, err := getTokenCurvePoolIndex(pTokenContract2.ContractIDGetRate, poolList)
			if err != nil {
				return nil, err
			}

			amountFloat := new(big.Float)
			amountFloat, ok := amountFloat.SetString(amount)
			if !ok {
				return nil, fmt.Errorf("amount is not a number")
			}
			amountBigFloat := ConvertToNanoIncognitoToken(amountFloat, int64(pTokenContract1.Decimals)) //amount *big.Float, decimal int64, return *big.Float
			log.Println("amountBigFloat: ", amountBigFloat.String())

			// convert float to bigin:
			amountBigInt, _ := amountBigFloat.Int(nil)

			log.Println("amountBigInt: ", amountBigInt)

			if amountBigInt == nil {
				return nil, errors.New("invalid amount")
			}

			i := big.NewInt(int64(token1PoolIndex))
			j := big.NewInt(int64(token2PoolIndex))

			curvePool := ethcommon.HexToAddress(curvePoolAddress1)

			networkInfo, err := database.DBGetBridgeNetworkInfo(networkName)
			if err != nil {
				return nil, err
			}
			var amountOut *big.Int

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
				}
			}

			expectOutputAmountBigFloat := new(big.Float)
			expectOutputAmountBigFloat, _ = expectOutputAmountBigFloat.SetString(amountOut.String())
			amountOutDecimal := ConvertToNanoIncognitoToken(expectOutputAmountBigFloat, int64(pTokenContract2.Decimals))

			fees := []PappNetworkFee{}

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

			result = append(result, QuoteDataResp{
				AppName:      appName,
				AmountIn:     amount,
				AmountOut:    amountOutDecimal.String(),
				AmountOutRaw: amountOut.String(),
				Fee:          fees,
			})
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

func checkValidTxSwap(md *bridge.BurnForCallRequest, feeToken string, feeAmount uint64, outCoins []coin.Coin) (bool, []string, error) {
	var result bool
	spTkList, err := getPappSupportedTokenList()
	if err != nil {
		return result, nil, err
	}
	networkInfo, err := getBridgeNetworkInfos()
	if err != nil {
		return result, nil, err
	}
	networkFees, err := database.DBRetrieveFeeTable()
	if err != nil {
		return result, nil, err
	}
	tokenInfo, err := getTokenInfo(md.BurnTokenID.String())
	if err != nil {
		return result, nil, err
	}
	callNetworkList := []string{}

	for _, v := range md.Data {
		callNetworkList = append(callNetworkList, wcommon.GetNetworkName(int(v.ExternalNetworkID)))
		receiveTokenID, err := getTokenIDByContractID(v.ReceiveToken, int(v.ExternalNetworkID), spTkList)
		if err != nil {
			return result, nil, err
		}
		burnAmountFloat := float64(v.BurningAmount) / math.Pow10(tokenInfo.PDecimals)
		burnAmountStr := fmt.Sprintf("%v", burnAmountFloat)
		quoteData, err := estimateSwapFee(v.IncTokenID.String(), receiveTokenID, burnAmountStr, int(v.ExternalNetworkID), spTkList, networkInfo, networkFees, tokenInfo)
		_ = quoteData

		//TODO
	}
	result = true
	return result, callNetworkList, nil
}

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

func buildPancakeTokenMap() (map[string]PancakeTokenMapItem, error) {
	result := make(map[string]PancakeTokenMapItem)
	tokenList, err := getPappSupportedTokenList()
	if err != nil {
		return nil, err
	}

	for _, token := range tokenList {
		if token.Protocol == "pancake" {
			result[strings.ToLower(token.ContractIDGetRate)] = PancakeTokenMapItem{
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

func getCurvePoolIndex() ([]CurvePoolIndex, error) {
	var responseBodyData struct {
		Result []CurvePoolIndex
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

func getTokenCurvePoolIndex(contractID string, poolList []CurvePoolIndex) (int, string, error) {
	for _, v := range poolList {
		if v.DappTokenAddress == contractID {
			return v.CurveTokenIndex, v.CurvePoolAddress, nil
		}
	}
	return -1, "", errors.New("pool not found")
}

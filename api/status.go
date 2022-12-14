package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/incognitochain/incognito-web-based-backend/common"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"go.mongodb.org/mongo-driver/mongo"
)

func APIGetStatusByShieldService(c *gin.Context) {
	pyd := c.Query("paymentaddress")
	shieldType := c.Query("type")

	var responseBodyData struct {
		Result []HistoryAddressResp `json:"Result"`
		Error  *struct {
			Code    int
			Message string
		} `json:"Error"`
	}

	var requestBody struct {
		WalletAddress       string
		PrivacyTokenAddress string
	}
	requestBody.WalletAddress = pyd
retry:
	re, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").SetHeader("Authorization", "Bearer "+usa.token).SetBody(requestBody).
		Post(config.ShieldService + "/eta/history")
	if err != nil {
		c.JSON(400, gin.H{"Error": err.Error()})
		return
	}

	err = json.Unmarshal(re.Body(), &responseBodyData)
	if err != nil {
		c.JSON(400, gin.H{"Error": err.Error()})
		return
	}

	if responseBodyData.Error != nil {
		if responseBodyData.Error.Code != 401 {
			c.JSON(400, gin.H{"Error": responseBodyData.Error})
			return
		} else {
			err = requestUSAToken(config.ShieldService)
			if err != nil {
				c.JSON(400, gin.H{"Error": err.Error()})
				return
			}
			goto retry
		}
	}

	filteredHistory := []HistoryAddressResp{}
	if shieldType == "unshield" {
		for _, v := range responseBodyData.Result {
			// 2 == unshield
			if v.AddressType == 2 {
				filteredHistory = append(filteredHistory, v)
			}
		}
	} else {
		for _, v := range responseBodyData.Result {
			// 1 == shield
			if v.AddressType == 1 {
				filteredHistory = append(filteredHistory, v)
			}
		}
	}

	resp := struct {
		Result []HistoryAddressResp
		Error  interface{}
	}{filteredHistory, nil}

	c.JSON(200, resp)
}

func APIGetFailedShieldTx(c *gin.Context) {

}

func APIGetPendingShieldTx(c *gin.Context) {

}

func APIGetPendingSwapTx(c *gin.Context) {

}

func APIGetUnshieldStatus(c *gin.Context) {

}

func APIGetShieldStatus(c *gin.Context) {
	//Todo: implement
	var req SubmitTxListRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
}

func APIGetSwapTxStatus(c *gin.Context) {
	var req SubmitTxListRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	result := make(map[string]interface{})

	spTkList := getSupportedTokenList()
	var wg sync.WaitGroup
	var lock sync.Mutex
	for _, txHash := range req.TxList {
		wg.Add(1)
		go func(txh string) {
			statusResult := checkPappTxSwapStatus(txh, spTkList)
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

func checkPappTxSwapStatus(txhash string, spTkList []PappSupportedTokenData) map[string]interface{} {
	result := make(map[string]interface{})
	data, err := database.DBGetPappTxData(txhash)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			result["error"] = err.Error()
			return result
		}
		return getPdexSwapTxStatus(txhash)
	}

	result["inc_request_tx_status"] = data.Status
	// result["inc_swap_detail"] = data.
	if data.Status != common.StatusAccepted {
		if data.Error != "" {
			result["error"] = data.Error
		}
	} else {
		networkList := []interface{}{}
		for _, network := range data.Networks {
			networkResult := make(map[string]interface{})
			networkResult["network"] = network
			outchainTx, err := database.DBGetExternalTxByIncTx(txhash, network)
			if err != nil {
				if err != mongo.ErrNoDocuments {
					networkResult["error"] = err.Error()
				} else {
					networkResult["swap_tx_status"] = common.StatusSubmitting
				}
				networkList = append(networkList, networkResult)
				continue
			}
			if outchainTx.Status == wcommon.StatusAccepted {
				networkResult["swap_tx_status"] = "success"
			} else {
				networkResult["swap_tx_status"] = outchainTx.Status
			}

			networkResult["swap_tx"] = outchainTx.Txhash
			if outchainTx.Error != "" {
				networkResult["swap_err"] = outchainTx.Error
			}
			if outchainTx.Status == common.StatusAccepted && outchainTx.OtherInfo != "" {
				var outchainTxResult wcommon.ExternalTxSwapResult
				err = json.Unmarshal([]byte(outchainTx.OtherInfo), &outchainTxResult)
				if err != nil {
					networkResult["error"] = err.Error()
					networkList = append(networkList, networkResult)
					continue
				}
				if outchainTxResult.IsReverted {
					networkResult["swap_outcome"] = "reverted"
				} else {
					networkResult["swap_outcome"] = "success"
				}
				networkResult["is_redeposit"] = outchainTxResult.IsRedeposit
				if outchainTxResult.IsFailed {
					networkResult["swap_outcome"] = "failed"
				}
				redepositTxStr := ""
				if outchainTxResult.IsRedeposit {
					networkID := wcommon.GetNetworkID(network)
					redepositTx, err := database.DBGetShieldTxByExternalTx(outchainTx.Txhash, networkID)
					if err != nil {
						if err != mongo.ErrNoDocuments {
							networkResult["error"] = err.Error()
						} else {
							networkResult["redeposit_status"] = common.StatusSubmitting
						}
						networkList = append(networkList, networkResult)
						continue
					}
					if redepositTx.Status == wcommon.StatusAccepted {
						networkResult["redeposit_status"] = "success"
					} else {
						networkResult["redeposit_status"] = redepositTx.Status
					}
					networkResult["redeposit_inctx"] = redepositTx.IncTx
					if data.BurntToken == "" {
						networkResult["swap_outcome"] = "unvailable"
					} else {
						if redepositTx.UTokenID == data.BurntToken {
							networkResult["swap_outcome"] = "reverted"
						} else {
							if redepositTx.TokenID == "" {
								networkResult["swap_outcome"] = "pending"
							} else {
								networkResult["swap_outcome"] = "success"
							}
						}
					}
					redepositTxStr = redepositTx.IncTx
				}
				if networkResult["swap_outcome"] == "success" {
					if outchainTxResult.TokenContract != "" {
						outTokenID, isNative, err := getTokenIDByContractID(outchainTxResult.TokenContract, common.GetNetworkID(network), spTkList, true)
						if err != nil {
							result["error"] = err.Error()
							continue
						}
						swapDetail := buildSwapDetail(data.BurntToken, outTokenID, common.GetNetworkID(network), data.BurntAmount, outchainTxResult.Amount.Uint64(), false, redepositTxStr, outchainTxResult.IsRedeposit, isNative)
						result["swap_detail"] = swapDetail
					}

				}
			}
			networkList = append(networkList, networkResult)
		}
		result["network_result"] = networkList
	}
	return result
}

func getPdexSwapTxStatus(txhash string) map[string]interface{} {
	result := make(map[string]interface{})

	type APIRespond struct {
		Result []wcommon.TradeDataRespond
		Error  *string
	}

	var responseBodyData APIRespond

	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).
		Get(config.CoinserviceURL + "/pdex/v3/tradedetail?txhash=" + txhash)
	if err != nil {
		log.Println("getPdexSwapTxStatus", err)
		return nil
	}
	if responseBodyData.Error != nil {
		log.Println("getPdexSwapTxStatus", errors.New(*responseBodyData.Error))
		return nil
	}

	if len(responseBodyData.Result) == 0 {
		result["error"] = "not found"
		return result
	}

	swapResult := responseBodyData.Result[0]

	result["is_pdex_swap"] = true
	result["inc_request_tx_status"] = wcommon.StatusPending
	if len(swapResult.RespondTxs) > 0 {
		result["inc_request_tx_status"] = swapResult.Status
		result["inc_respond_tx"] = swapResult.RespondTxs[0]
	}

	if swapResult.Status == "accepted" {
		swapDetail := buildSwapDetail(swapResult.SellTokenID, swapResult.BuyTokenID, 0, swapResult.Amount, swapResult.RespondAmounts[0], true, "", false, false)
		result["swap_detail"] = swapDetail
	}

	return result
}

func buildSwapDetail(tokenIn, tokenOut string, networkID int, inAmount uint64, outAmount uint64, isPdex bool, redepositTx string, willRedeposit, isNative bool) map[string]interface{} {
	result := make(map[string]interface{})

	tokenInInfo, err := getTokenInfo(tokenIn)
	if err != nil {
		return nil
	}
	tokenOutInfo, err := getTokenInfo(tokenOut)
	if err != nil {
		return nil
	}

	inAmountBig := new(big.Float).SetUint64(inAmount)
	inDecimal := new(big.Float).SetFloat64(math.Pow10(-tokenInInfo.PDecimals))
	inAmountfl64, _ := inAmountBig.Mul(inAmountBig, inDecimal).Float64()

	outAmountBig := new(big.Float).SetUint64(outAmount)
	var outAmountfl64 float64
	var outDecimal *big.Float
	if isPdex {
		outDecimal = new(big.Float).SetFloat64(math.Pow10(-tokenOutInfo.PDecimals))
		outAmountfl64, _ = new(big.Float).Mul(outAmountBig, outDecimal).Float64()
	} else {
		if tokenOutInfo.CurrencyType == wcommon.UnifiedCurrencyType {
			for _, ctk := range tokenOutInfo.ListUnifiedToken {
				netID, _ := wcommon.GetNetworkIDFromCurrencyType(ctk.CurrencyType)
				if netID == networkID {
					if isNative {
						outDecimal = new(big.Float).SetFloat64(math.Pow10(-int(ctk.Decimals)))
					} else {
						if willRedeposit {
							outDecimal = new(big.Float).SetFloat64(math.Pow10(-int(ctk.PDecimals)))
						} else {
							outDecimal = new(big.Float).SetFloat64(math.Pow10(-int(ctk.Decimals)))
						}
					}
					outAmountfl64, _ = new(big.Float).Mul(outAmountBig, outDecimal).Float64()
					break
				}
			}
		} else {
			if isNative {
				outDecimal = new(big.Float).SetFloat64(math.Pow10(-int(tokenOutInfo.Decimals)))
			} else {
				if willRedeposit {
					outDecimal = new(big.Float).SetFloat64(math.Pow10(-int(tokenOutInfo.PDecimals)))
				} else {
					outDecimal = new(big.Float).SetFloat64(math.Pow10(-int(tokenOutInfo.Decimals)))
				}
			}
			outAmountfl64, _ = new(big.Float).Mul(outAmountBig, outDecimal).Float64()
		}
	}

	result["token_in"] = tokenIn
	result["token_out"] = tokenOut
	result["in_amount"] = fmt.Sprintf("%f", inAmountfl64)

	if redepositTx != "" {
		shieldStatus, err := getShieldStatus(config.FullnodeURL, redepositTx)
		if err != nil {
			result["err"] = err.Error()
			result["out_amount"] = fmt.Sprintf("%f", outAmountfl64)
			return result
		}
		if len(shieldStatus.Data) == 0 {
			result["out_amount"] = fmt.Sprintf("%f", outAmountfl64)
			return result
		}
		if shieldStatus.Data[0].Reward > 0 {
			outDecimal = new(big.Float).SetFloat64(math.Pow10(-int(tokenOutInfo.PDecimals)))
			rewardAmountBig := new(big.Float).SetUint64(shieldStatus.Data[0].Reward)
			rewardAmountfl64, _ := new(big.Float).Mul(rewardAmountBig, outDecimal).Float64()
			result["reward"] = fmt.Sprintf("%f", rewardAmountfl64)
		} else {
			result["reward"] = 0
		}
	}
	result["out_amount"] = fmt.Sprintf("%f", outAmountfl64)
	return result
}

func APITrackDEXSwap(c *gin.Context) {
	var req DexSwap
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	if req.Txhash == "" || req.TokenSell == "" || req.TokenBuy == "" || req.AmountIn == "" || req.MinAmountOut == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid request"})
		return
	}

	txdata := wcommon.DexSwapTrackData{
		IncTx:        req.Txhash,
		Status:       wcommon.StatusPending,
		TokenSell:    req.TokenSell,
		TokenBuy:     req.TokenBuy,
		AmountIn:     req.AmountIn,
		MinAmountOut: req.MinAmountOut,
		UserAgent:    c.Request.UserAgent(),
	}
	err = database.DBSaveDexSwapTxData(txdata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	log.Println("APITrackDEXSwap", req.Txhash)
	c.JSON(http.StatusOK, gin.H{"Result": "ok"})
	return
}

func checkUnshieldTxStatus(txhash string) map[string]interface{} {
	result := make(map[string]interface{})
	data, err := database.DBGetUnshieldTxByIncTx(txhash)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			result["error"] = err.Error()
			return result
		}
	}

	result["inc_request_tx_status"] = data.Status
	if data.Status != common.StatusAccepted {
		if data.Error != "" {
			result["error"] = data.Error
		}
	} else {
		network := data.Networks[0]
		result["network"] = network
		outchainTx, err := database.DBGetExternalTxByIncTx(txhash, network)
		if err != nil {
			if err != mongo.ErrNoDocuments {
				result["error"] = err.Error()
			} else {
				result["outchain_status"] = common.StatusSubmitting
			}
			return result
		}
		if outchainTx.Status == wcommon.StatusAccepted {
			result["outchain_status"] = "success"
		} else {
			result["outchain_status"] = outchainTx.Status
		}

		result["outchain_tx"] = outchainTx.Txhash
		if outchainTx.Error != "" {
			result["outchain_err"] = outchainTx.Error
		}
		if outchainTx.Status == common.StatusAccepted && outchainTx.OtherInfo != "" {
			var outchainTxResult wcommon.ExternalTxSwapResult
			err = json.Unmarshal([]byte(outchainTx.OtherInfo), &outchainTxResult)
			if err != nil {
				result["error"] = err.Error()
			}
			if outchainTxResult.IsReverted {
				result["outchain_status"] = "reverted"
			} else {
				result["outchain_status"] = "success"
			}
			if outchainTxResult.IsFailed {
				result["outchain_status"] = "failed"
			}

		}
	}
	return result
}

package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

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

func APIGetUnshieldStatus(c *gin.Context) {

}

func APIGetShieldStatus(c *gin.Context) {
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
	for _, txHash := range req.TxList {
		statusResult := checkPappTxSwapStatus(txHash)
		if len(statusResult) == 0 {
			statusResult["error"] = "tx not found"
			result[txHash] = statusResult
		} else {
			result[txHash] = statusResult
		}
	}
	c.JSON(200, gin.H{"Result": result})
}

func checkPappTxSwapStatus(txhash string) map[string]interface{} {
	result := make(map[string]interface{})
	data, err := database.DBGetPappTxData(txhash)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			result["error"] = err.Error()
		} else {
			return getPdexSwapTxStatus(txhash)
		}
		result["error"] = "not found"
		return result
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
			networkResult["swap_tx_status"] = outchainTx.Status
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
					networkResult["redeposit_status"] = redepositTx.Status
					networkResult["redeposit_inctx"] = redepositTx.IncTx
					if data.BurntToken == "" {
						networkResult["swap_outcome"] = "unvailable"
					} else {
						if redepositTx.UTokenID == data.BurntToken {
							networkResult["swap_outcome"] = "reverted"
						} else {
							if redepositTx.UTokenID == "" {
								networkResult["swap_outcome"] = "pending"
							} else {
								networkResult["swap_outcome"] = "success"
							}
						}
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
		Result []TradeDataRespond
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
	result["inc_request_tx_status"] = swapResult.Status
	result["inc_respond_tx"] = swapResult.RespondTxs[0]

	return result
}

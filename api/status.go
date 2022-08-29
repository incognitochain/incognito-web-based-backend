package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/incognitochain/incognito-web-based-backend/common"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
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
		statusResult, err := checkPappTxSwapStatus(txHash)
		if err != nil {
			result[txHash] = err
			continue
		}
		result[txHash] = statusResult
	}
	c.JSON(200, gin.H{"Result": result})
}

func checkPappTxSwapStatus(txhash string) (map[string]string, error) {
	result := make(map[string]string)
	data, err := database.DBGetPappTxData(txhash)
	if err != nil {
		return result, err
	}

	if data.Status != common.StatusAccepted {
		result["submitswaptx"] = data.Status
		if data.Error != "" {
			result["error"] = data.Error
		}
	} else {
		result["submitswaptx"] = common.StatusAccepted
		for _, network := range data.Networks {
			outchainTx, err := database.DBGetExternalTxByIncTx(txhash, network)
			if err != nil {
				return result, err
			}
			result[network] = outchainTx.Status
			result[network+"_exttx"] = outchainTx.Txhash
			if outchainTx.Error != "" {
				result[network+"_err"] = outchainTx.Error
			}
			if outchainTx.Status == common.StatusAccepted && outchainTx.OtherInfo != "" {
				var outchainTxResult wcommon.ExternalTxSwapResult
				err = json.Unmarshal([]byte(outchainTx.OtherInfo), &outchainTxResult)
				if err != nil {
					return result, err
				}
				if outchainTxResult.IsRedeposit {
					networkID := wcommon.GetNetworkID(network)
					redepositTx, err := database.DBGetShieldTxByExternalTx(outchainTx.Txhash, networkID)
					if err != nil {
						return result, err
					}
					result[network+"_redeposit"] = redepositTx.Status
					result[network+"_redeposit_inctx"] = redepositTx.IncTx
				}
			}
		}
	}
	return result, nil
}

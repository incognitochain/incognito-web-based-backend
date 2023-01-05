package api

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
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

func APIGetStatusByIncTx(c *gin.Context) {
	txHash := c.Query("txhash")
	shieldType := c.Query("type")
	_ = txHash
	_ = shieldType
}

func APIGetFailedShieldTx(c *gin.Context) {

}

func APIGetShieldStatus(c *gin.Context) {

}

func APIGetUnshieldStatus(c *gin.Context) {

}

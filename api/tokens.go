package api

import (
	"github.com/gin-gonic/gin"
	"github.com/incognitochain/coin-service/apiservice"
)

func APIGetSupportedToken(c *gin.Context) {

}

func APIGetTokenList(c *gin.Context) {
	var responseBodyData apiservice.APIRespond
	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).
		Get(config.CoinserviceURL + "/coins/tokenlist")
	if err != nil {
		c.JSON(400, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(200, responseBodyData)
}

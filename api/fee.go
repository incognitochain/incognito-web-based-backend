package api

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func APIEstimateSwapFee(c *gin.Context) {
	var req EstimateSwapRequest
	err := c.MustBindWith(&req, binding.JSON)
	if err != nil {
		c.JSON(400, gin.H{"Error": err.Error()})
		return
	}

	// var result EstimateSwapRespond
	var response struct {
		Result interface{}
		Error  interface{}
	}

	c.JSON(200, response)

}

func APIEstimateReward(c *gin.Context) {
	var req EstimateRewardRequest
	err := c.MustBindWith(&req, binding.JSON)
	if err != nil {
		c.JSON(400, gin.H{"Error": err.Error()})
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
		c.JSON(400, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(200, responseBodyData)
}

func APIEstimateUnshield(c *gin.Context) {
	var req EstimateUnshieldRequest
	err := c.MustBindWith(&req, binding.JSON)
	if err != nil {
		c.JSON(400, gin.H{"Error": err.Error()})
		return
	}

	if req.ExpectedAmount > 0 && req.BurntAmount > 0 {
		c.JSON(400, gin.H{"Error": errors.New("either ExpectedAmount or BurntAmount can > 0, not both")})
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
		c.JSON(400, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(200, responseBodyData)
}

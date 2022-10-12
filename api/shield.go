package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/submitproof"
)

func APIEstimateReward(c *gin.Context) {
	var req EstimateRewardRequest
	err := c.MustBindWith(&req, binding.JSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(200, responseBodyData)
}

func APIEstimateUnshield(c *gin.Context) {
	var req EstimateUnshieldRequest
	err := c.MustBindWith(&req, binding.JSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	if req.Network == "btc" {
		getBTCUnshieldFee(c)
		return
	}

	if req.ExpectedAmount > 0 && req.BurntAmount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("either ExpectedAmount or BurntAmount can > 0, not both")})
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
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(200, responseBodyData)
}

func getBTCUnshieldFee(c *gin.Context) {
	var responseBodyData struct {
		Result float64
		Error  interface{}
	}
	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).
		Get(config.BTCShieldPortal + "/getestimatedunshieldingfee")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	unshieldFee := responseBodyData.Result
	minUnshield := ""

	var responseRPCData struct {
		Result interface{}
		Error  interface{}
	}
	methodRPC := "getportalv4params"
	beaconHeight, err := getCurrentBeaconHeight(config.FullnodeURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	beaconHeightStr := fmt.Sprintf("%v", beaconHeight)
	reqRPC := genRPCBody(methodRPC, []interface{}{
		map[string]interface{}{
			"BeaconHeight": beaconHeightStr,
		},
	})

	_, err = restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseRPCData).SetBody(reqRPC).
		Post(config.FullnodeURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	result := make(map[string]interface{})
	result["Fee"] = unshieldFee
	result["MinUnshield"] = minUnshield

	c.JSON(200, gin.H{"Result": result})
}

func APIRetryShieldTx(c *gin.Context) {
	var req SubmitShieldTx
	err := c.MustBindWith(&req, binding.JSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	if config.CaptchaSecret != "" {
		if ok, err := VerifyCaptcha(req.Captcha, config.CaptchaSecret); !ok {
			if err != nil {
				log.Println("VerifyCaptcha", err)
				c.JSON(http.StatusBadRequest, gin.H{"Error": err})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid captcha")})
			return
		}

	}

	if req.Txhash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid params").Error()})
		return
	}

	status, err := database.DBGetShieldTxStatusByExternalTx(req.Txhash, req.Network)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err})
		return
	}

	if status == common.StatusSubmitFailed {
		statusRetry, err := submitproof.SubmitShieldProof(req.Txhash, req.Network, req.TokenID, submitproof.TxTypeShielding, true)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"Result": statusRetry})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "status isn't submit_failed"})
		return
	}

}

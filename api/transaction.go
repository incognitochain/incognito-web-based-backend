package api

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/mongodb/mongo-tools/common/json"
)

func APISubmitUnshieldTx(c *gin.Context) {
	var req SubmitUnshieldTxRequest
	err := c.MustBindWith(&req, binding.JSON)
	if err != nil {
		c.JSON(400, gin.H{"Error": err.Error()})
		return
	}

	switch req.Network {
	case "eth", "bsc", "plg", "ftm":
		re, err := restyClient.R().
			EnableTrace().
			SetHeader("Content-Type", "application/json").SetHeader("Authorization", "Bearer "+usa.token).SetBody(req).
			Post(config.ShieldService + "/" + req.Network + "/add-tx-withdraw")
		if err != nil {
			c.JSON(400, gin.H{"Error": err.Error()})
			return
		}
		var responseBodyData struct {
			Result interface{}
			Error  interface{}
		}
		err = json.Unmarshal(re.Body(), &responseBodyData)
		if err != nil {
			c.JSON(400, gin.H{"Error": err})
			return
		}
		c.JSON(200, responseBodyData)
		return
	default:
		c.JSON(400, gin.H{"Error": errors.New("unsupport network")})
		return
	}
}

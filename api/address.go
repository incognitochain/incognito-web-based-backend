package api

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/mongodb/mongo-tools/common/json"
)

func APIGenUnshieldAddress(c *gin.Context) {
	var req GenUnshieldAddressRequest
	err := c.MustBindWith(&req, binding.JSON)
	if err != nil {
		c.JSON(400, gin.H{"Error": err.Error()})
		return
	}

	switch req.Network {
	case "eth", "bsc", "plg", "ftm":
	retry:
		re, err := restyClient.R().
			EnableTrace().
			SetHeader("Content-Type", "application/json").SetHeader("Authorization", "Bearer "+usa.token).SetBody(req).
			Post(config.ShieldService + "/" + req.Network + "/estimate-fees")
		if err != nil {
			c.JSON(400, gin.H{"Error": err.Error()})
			return
		}
		var responseBodyData struct {
			Result interface{}
			Error  *struct {
				Code    int
				Message string
			} `json:"Error"`
		}
		err = json.Unmarshal(re.Body(), &responseBodyData)
		if err != nil {
			c.JSON(400, gin.H{"Error": err})
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

		c.JSON(200, responseBodyData)
		return
	default:
		c.JSON(400, gin.H{"Error": errors.New("unsupport network")})
		return
	}
}

func APIGenShieldAddress(c *gin.Context) {
	var req GenShieldAddressRequest
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
			Post(config.ShieldService + "/" + req.Network + "/generate")
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

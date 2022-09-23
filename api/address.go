package api

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/incognitochain/incognito-web-based-backend/common"
)

func APIGenUnshieldAddress(c *gin.Context) {
	var req GenUnshieldAddressRequest
	err := c.MustBindWith(&req, binding.JSON)
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}

	switch req.Network {
	case "centralized":
		genCentralizedUnshieldAddress(c, req)
	case "eth", "bsc", "plg", "ftm":
		genUnshieldAddress(c, req)
	default:
		c.JSON(200, gin.H{"Error": errors.New("unsupport network")})
		return
	}
}

func genCentralizedUnshieldAddress(c *gin.Context, req GenUnshieldAddressRequest) {
	reqWarpped := GenShieldAddressRequest{
		AddressType:         req.AddressType,
		CurrencyType:        req.CurrencyType,
		PrivacyTokenAddress: req.PrivacyTokenAddress,
		WalletAddress:       req.WalletAddress,
		PaymentAddress:      req.PaymentAddress,
	}
	genCentralizedShieldAddress(c, reqWarpped)
}

func genUnshieldAddress(c *gin.Context, req GenUnshieldAddressRequest) {
retry:
	re, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").SetHeader("Authorization", "Bearer "+usa.token).SetBody(req).
		Post(config.ShieldService + "/" + req.Network + "/estimate-fees")
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
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
		c.JSON(200, gin.H{"Error": err})
		return
	}

	if responseBodyData.Error != nil {
		if responseBodyData.Error.Code != 401 {
			c.JSON(200, gin.H{"Error": responseBodyData.Error})
			return
		} else {
			err = requestUSAToken(config.ShieldService)
			if err != nil {
				c.JSON(200, gin.H{"Error": err.Error()})
				return
			}
			goto retry
		}
	}

	c.JSON(200, responseBodyData)
	return
}

func APIGenShieldAddress(c *gin.Context) {
	var req GenShieldAddressRequest
	err := c.MustBindWith(&req, binding.JSON)
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}
	if req.WalletAddress == "" {
		req.WalletAddress = "12stRSuMnrJLnNZNWBP2K66gzQMBL4WyzzJHHn1diGnbeuYJNzohByEYiFYS1rfazEWrtLD6i8du2i4LeZMLiTeCRpQ1cSTyAuLyumCc21FdZNTSp6Gs5JjobsAWJR8q5YLDzB4HWpQZSxpRBfGT"
	}

	switch req.Network {
	case "btc":
		genBTCShieldAddress(c, req)
	case "centralized":
		genCentralizedShieldAddress(c, req)
	case "eth", "bsc", "plg", "ftm":
		genEVMShieldAddress(c, req)
	default:
		c.JSON(200, gin.H{"Error": errors.New("unsupport network")})
		return
	}
}

func genBTCOTMultisigAddress(incAddress string) (string, error) {
	tokenID := common.MainnetPortalV4BTCID
	if config.NetworkID == "testnet" {
		tokenID = common.TestnetPortalV4BTCID
	}

	methodRPC := "generateportalshieldmultisigaddress"
	reqRPC := genRPCBody(methodRPC, []interface{}{
		map[string]interface{}{
			"IncAddressStr": incAddress,
			"TokenID":       tokenID,
		},
	})

	var responseBodyData struct {
		Result interface{}
		Error  interface{}
	}
	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).SetBody(reqRPC).
		Post(config.FullnodeURL)
	if err != nil {
		return "", err
	}
	if responseBodyData.Error != nil {
		log.Println("genBTCOTMultisigAddress", err)
		return "", errors.New("gen BTC OTMulsig err")
	}

	return responseBodyData.Result.(string), nil

}

func genBTCShieldAddress(c *gin.Context, req GenShieldAddressRequest) {
	btcOT, err := genBTCOTMultisigAddress(req.BTCIncAddress)
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}

	btcReq := GenBTCShieldAddressRequest{
		IncAddress:    req.BTCIncAddress,
		ShieldAddress: btcOT,
	}

	var responseBodyData struct {
		Result interface{}
		Error  string
	}
	_, err = restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").SetBody(btcReq).SetResult(&responseBodyData).
		Post(config.BTCShieldPortal + "/addportalshieldingaddress")
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
		return
	}
	if responseBodyData.Result == true {
		c.JSON(200, gin.H{"Result": btcOT})
		return
	} else {
		if responseBodyData.Error == "Record has already been inserted" {
			c.JSON(200, gin.H{"Result": btcOT})
			return
		}
	}
	c.JSON(200, gin.H{"Error": responseBodyData.Error})
}

func genCentralizedShieldAddress(c *gin.Context, req GenShieldAddressRequest) {
retry:
	re, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").SetHeader("Authorization", "Bearer "+usa.token).SetBody(req).
		Post(config.ShieldService + "/ota/generate")
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
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
		c.JSON(200, gin.H{"Error": err})
		return
	}
	if responseBodyData.Error != nil {
		if responseBodyData.Error.Code != 401 {
			c.JSON(200, gin.H{"Error": responseBodyData.Error})
			return
		} else {
			err = requestUSAToken(config.ShieldService)
			if err != nil {
				c.JSON(200, gin.H{"Error": err.Error()})
				return
			}
			goto retry
		}
	}

	c.JSON(200, responseBodyData)
}

func genEVMShieldAddress(c *gin.Context, req GenShieldAddressRequest) {
retry:
	re, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").SetHeader("Authorization", "Bearer "+usa.token).SetBody(req).
		Post(config.ShieldService + "/" + req.Network + "/generate")
	if err != nil {
		c.JSON(200, gin.H{"Error": err.Error()})
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
		c.JSON(200, gin.H{"Error": err})
		return
	}
	if responseBodyData.Error != nil {
		if responseBodyData.Error.Code != 401 {
			c.JSON(200, gin.H{"Error": responseBodyData.Error})
			return
		} else {
			err = requestUSAToken(config.ShieldService)
			if err != nil {
				c.JSON(200, gin.H{"Error": err.Error()})
				return
			}
			goto retry
		}
	}

	c.JSON(200, responseBodyData)
}

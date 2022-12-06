package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/submitproof"
)

func APISubmitUnshieldTx(c *gin.Context) {
	var req SubmitUnshieldTxRequest
	err := c.MustBindWith(&req, binding.JSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	authToken := c.Request.Header.Get("Authorization")
	// err = requestUSAToken(config.ShieldService)
	// if err != nil {
	// 	c.JSON(400, gin.H{"Error": err.Error()})
	// 	return
	// }
	switch req.Network {
	case "centralized":
		re, err := restyClient.R().
			EnableTrace().
			SetHeader("Content-Type", "application/json").SetHeader("Authorization", authToken).SetBody(req).
			Post(config.ShieldService + "/ota/update-fee")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		var responseBodyData struct {
			Result interface{}
			Error  interface{}
		}
		err = json.Unmarshal(re.Body(), &responseBodyData)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err})
			return
		}
		c.JSON(200, responseBodyData)
		return
	case "eth", "bsc", "plg", "ftm", "avax", "aurora", "near":
		re, err := restyClient.R().
			EnableTrace().
			SetHeader("Content-Type", "application/json").SetHeader("Authorization", authToken).SetBody(req).
			Post(config.ShieldService + "/" + req.Network + "/add-tx-withdraw")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		var responseBodyData struct {
			Result interface{}
			Error  interface{}
		}
		err = json.Unmarshal(re.Body(), &responseBodyData)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err})
			return
		}
		c.JSON(200, responseBodyData)
		return
	default:
		if req.ID == 0 && req.PaymentAddress == "" && req.WalletAddress != "" && req.IncognitoTx != "" {
			txdetail, err := getTxDetails(req.IncognitoTx)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": err})
				return
			}

			ID, PaymentAddress, PrivacyTokenAddress, IncognitoAmount, Network, err := extractUnshieldInfoField(txdetail)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": err})
				return
			}
			newReq := SubmitUnshieldTxRequest{
				IncognitoAmount:     IncognitoAmount,
				PaymentAddress:      PaymentAddress,
				PrivacyTokenAddress: PrivacyTokenAddress,
				WalletAddress:       req.WalletAddress,
				UserFeeLevel:        1,
				IncognitoTx:         req.IncognitoTx,
				ID:                  ID,
				UserFeeSelection:    1,
			}

			a, _ := json.Marshal(newReq)
			fmt.Println("newReq", string(a))

			switch Network {
			case "eth", "bsc", "plg", "ftm", "avax", "aurora":
				re, err := restyClient.R().
					EnableTrace().
					SetHeader("Content-Type", "application/json").SetHeader("Authorization", authToken).SetBody(newReq).
					Post(config.ShieldService + "/" + Network + "/add-tx-withdraw")
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
					return
				}
				var responseBodyData struct {
					Result interface{}
					Error  interface{}
				}
				err = json.Unmarshal(re.Body(), &responseBodyData)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"Error": err})
					return
				}
				c.JSON(200, responseBodyData)
				return
			default:
				c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("unsupport network")})
				return
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("unsupport network")})
		return
	}
}

func APISubmitShieldTx(c *gin.Context) {
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
	status, err := submitproof.SubmitShieldProof(req.Txhash, req.Network, req.TokenID, submitproof.TxTypeShielding, false)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"Result": status})
}

func getTxDetails(txhash string) (*TransactionDetail, error) {
	reqRPC := genRPCBody("gettransactionbyhash", []interface{}{
		txhash,
	})

	type TxDetailRespond struct {
		Result TransactionDetail
		Error  *string
	}

	var responseBodyData TxDetailRespond
	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).SetBody(reqRPC).
		Post(config.FullnodeURL)
	if err != nil {
		return nil, err
	}

	if responseBodyData.Error != nil {
		return nil, fmt.Errorf("%v", responseBodyData.Error)
	}
	return &responseBodyData.Result, nil
}

func extractUnshieldInfoField(txdetail *TransactionDetail) (ID int, PaymentAddress string, PrivacyTokenAddress string, IncognitoAmount string, Network string, err error) {
	ID, err = strconv.Atoi(txdetail.Info)
	if err != nil {
		return
	}
	var unshieldMeta UnshieldRequest

	err = json.Unmarshal([]byte(txdetail.Metadata), &unshieldMeta)
	if err != nil {
		return
	}
	if unshieldMeta.Type != 345 {
		err = errors.New("Invalid metadata type")
		return
	}
	PrivacyTokenAddress = unshieldMeta.Data[0].IncTokenID.String()
	networkID := getTokenNetwork(unshieldMeta.UnifiedTokenID.String(), PrivacyTokenAddress)

	switch networkID {
	case 0:
		err = errors.New("unsupported network")
		return
	case 1:
		Network = "eth"
	case 2:
		Network = "bsc"
	case 3:
		Network = "plg"
	case 4:
		Network = "ftm"
	case 5:
		Network = "avax"
	case 6:
		Network = "aurora"
	}
	IncognitoAmount = fmt.Sprintf("%d", unshieldMeta.Data[0].BurningAmount)
	PaymentAddress = "0x" + unshieldMeta.Data[0].RemoteAddress
	return
}

func getTokenInfo(pUTokenID string) (*common.TokenInfo, error) {
	type APIRespond struct {
		Result []common.TokenInfo
		Error  *string
	}

	reqBody := struct {
		TokenIDs []string
	}{
		TokenIDs: []string{pUTokenID},
	}

	var responseBodyData APIRespond
	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).SetBody(reqBody).
		Post(config.CoinserviceURL + "/coins/tokeninfo")
	if err != nil {
		return nil, err
	}

	if len(responseBodyData.Result) == 1 {
		return &responseBodyData.Result[0], nil
	}
	return nil, errors.New(fmt.Sprintf("token not found"))
}

func getTokensInfo(pUTokenID []string) ([]common.TokenInfo, error) {
	type APIRespond struct {
		Result []common.TokenInfo
		Error  *string
	}

	reqBody := struct {
		TokenIDs []string
	}{
		TokenIDs: pUTokenID,
	}

	var responseBodyData APIRespond
	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).SetBody(reqBody).
		Post(config.CoinserviceURL + "/coins/tokeninfo")
	if err != nil {
		return nil, err
	}

	if len(responseBodyData.Result) == 0 {
		return nil, errors.New(fmt.Sprintf("tokens not found"))
	}
	return responseBodyData.Result, nil
}

func getTokenNetwork(pUTokenID string, tokenID string) int {
	tokenInfo, err := getTokenInfo(pUTokenID)
	if err != nil {
		log.Println("getLinkedTokenID", err)
		return 0
	}
	for _, v := range tokenInfo.ListUnifiedToken {
		if v.TokenID == tokenID {
			return v.NetworkID
		}
	}
	return 0
}

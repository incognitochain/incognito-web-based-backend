package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/incognitochain/go-incognito-sdk-v2/common/base58"
	"github.com/incognitochain/go-incognito-sdk-v2/metadata"
	"github.com/incognitochain/go-incognito-sdk-v2/metadata/bridge"
	"github.com/incognitochain/go-incognito-sdk-v2/transaction"
)

func APISubmitSwapTx(c *gin.Context) {
	var req SubmitSwapTxRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if req.TxRaw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid txhash")})
		return
	}
	rawTxBytes, _, err := base58.Base58Check{}.Decode(req.TxRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid txhash")})
		return
	}

	// Unmarshal from json data to object tx))
	tx, err := transaction.DeserializeTransactionJSON(rawTxBytes)
	// var tx transaction.Tx
	// err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		tx, err = transaction.DeserializeTransactionJSON(rawTxBytes)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
	}
	if tx.TokenVersion2 != nil {
		if tx.TokenVersion2.GetMetadataType() != metadata.BurnForCallRequestMeta {
			md := tx.TokenVersion2.GetMetadata().(*bridge.BurnForCallRequest)
			_ = md
		}
	}
	if tx.Version2 != nil {
		if tx.Version2.GetMetadataType() != metadata.BurnForCallRequestMeta {
			md := tx.Version2.GetMetadata().(*bridge.BurnForCallRequest)
			_ = md
		}
	}

}

func checkValidTxSwap(md *bridge.BurnForCallRequest) {}

func APIGetVaultState(c *gin.Context) {
	var responseBodyData APIRespond
	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).
		Get(config.CoinserviceURL + "/bridge/aggregatestate")
	if err != nil {
		c.JSON(400, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(200, responseBodyData)
}

func sendSwapTxAndStoreDB(txhash string, txRaw string, isTokenTx bool) error {
	if isTokenTx {
		err := incClient.SendRawTokenTx([]byte(txRaw))
		if err != nil {
			return err
		}
	} else {
		err := incClient.SendRawTx([]byte(txRaw))
		if err != nil {
			return err
		}
	}
	return nil
}

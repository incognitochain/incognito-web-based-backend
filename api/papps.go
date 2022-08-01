package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/incognitochain/go-incognito-sdk-v2/common/base58"
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

	transaction.DeserializeTransactionJSON()
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
		tx.TokenVersion2.GetMetadataType() == 
	}
	if tx.Version2 != nil {
	}

}

func checkValidTxSwap() {}

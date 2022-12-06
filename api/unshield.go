package api

import (
	"errors"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/incognitochain/go-incognito-sdk-v2/common/base58"
	"github.com/incognitochain/go-incognito-sdk-v2/metadata/bridge"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
)

func APIGetUnshieldTxStatus(c *gin.Context) {
	var req SubmitTxListRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	result := make(map[string]interface{})

	var wg sync.WaitGroup
	var lock sync.Mutex
	for _, txHash := range req.TxList {
		wg.Add(1)
		go func(txh string) {
			statusResult := checkPappTxSwapStatus(txh, spTkList)
			lock.Lock()
			if len(statusResult) == 0 {
				statusResult["error"] = "tx not found"
				result[txh] = statusResult
			} else {
				result[txh] = statusResult
			}
			lock.Unlock()
			wg.Done()
		}(txHash)
	}
	wg.Wait()
	c.JSON(200, gin.H{"Result": result})
}

func APIUnshieldFeeEstimate(c *gin.Context) {
	network := c.Query("network")
	tokenid := c.Query("token")
	amount := c.Query("amount")

	_ = network
	_ = tokenid
	_ = amount

}

func APISubmitUnshieldTxNew(c *gin.Context) {
	var req SubmitSwapTxRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	rawTxBytes, _, err := base58.Base58Check{}.Decode(req.TxRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid txhash")})
		return
	}

	mdRaw, isPRVTx, outCoins, txHash, err := extractDataFromRawTx(rawTxBytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	var mdUnified *bridge.UnshieldRequest
	var md *bridge.BurningRequest
	md, ok := mdRaw.(*bridge.BurningRequest)
	if !ok {
		mdUnified, ok = mdRaw.(*bridge.UnshieldRequest)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
	}

	if md == nil {
		//unshield unified
		mdUnified.UnifiedTokenID
	}

	burnTokenInfo, err := getTokenInfo(md.BurnTokenID.String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid tx metadata type")})
		return
	}
	if burnTokenInfo.CurrencyType == wcommon.UnifiedCurrencyType {
		isUnifiedToken = true
	}

	estimateUnshieldFee()
}

func estimateUnshieldFee(amount uint64, tokenID string, network string) uint64 {
	var result uint64
	return result
}

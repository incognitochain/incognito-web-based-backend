package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/incognitochain/incognito-web-based-backend/papps/popensea"
)

func APIEstimateBuyFee(c *gin.Context) {}
func APISubmitBuyTx(c *gin.Context)    {}
func APIGetBuyTxStatus(c *gin.Context) {}

func APIGetCollections(c *gin.Context) {
	collections, err := popensea.RetrieveCollectionList(config.OpenSeaAPI, config.OpenSeaAPIKey, 20, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"Result": collections})
}
func APINFTDetail(c *gin.Context) {
	contract := c.Query("contract")
	tokenID := c.Query("token_id")
	nftDetail, err := popensea.RetrieveNFTDetail(config.OpenSeaAPI, config.OpenSeaAPIKey, contract, tokenID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"Result": nftDetail})
}
func APICollectionAssets(c *gin.Context) {
	contract := c.Query("contract")
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	assetList, err := popensea.RetrieveCollectionAssets(config.OpenSeaAPI, config.OpenSeaAPIKey, contract, limit, offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"Result": assetList})
}
func APICollectionDetail(c *gin.Context) {}

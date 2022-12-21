package api

import "github.com/gin-gonic/gin"

// pOpenSeaGroup.POST("/estimatebuyfee", APIEstimateBuyFee)
// pOpenSeaGroup.POST("/submitbuytx", APISubmitBuyTx)
// pOpenSeaGroup.POST("/buystatus", APIGetBuyTxStatus)
// //opensea api
// pOpenSeaGroup.GET("/collections", APIGetCollections)
// pOpenSeaGroup.GET("/nft-detail", APINFTDetail)
// pOpenSeaGroup.GET("/collection-assets", APICollectionDetail)
// pOpenSeaGroup.GET("/collection-detail", APICollectionDetail)

func APIEstimateBuyFee(c *gin.Context)   {}
func APISubmitBuyTx(c *gin.Context)      {}
func APIGetBuyTxStatus(c *gin.Context)   {}
func APIGetCollections(c *gin.Context)   {}
func APINFTDetail(c *gin.Context)        {}
func APICollectionAssets(c *gin.Context) {}
func APICollectionDetail(c *gin.Context) {}

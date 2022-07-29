package api

import (
	"log"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/patrickmn/go-cache"
)

var config common.Config

func StartAPIservice(cfg common.Config) {
	log.Println("initiating api-service...")
	config = cfg
	cachedb = cache.New(5*time.Minute, 5*time.Minute)

	r := gin.Default()

	r.Use(cors.Default())

	r.Use(gzip.Gzip(gzip.DefaultCompression))

	r.GET("/tokenlist", APIGetTokenList)

	r.POST("/estimateshieldreward", APIEstimateReward)

	r.POST("/estimateunshieldfee", APIEstimateUnshield)

	r.POST("/genunshieldaddress", APIGenUnshieldAddress)

	// r.POST("/genshieldaddress", APIGenShieldAddress)

	r.POST("/submitunshieldtx", APISubmitUnshieldTx)

	r.POST("/submitshieldtx", APISubmitShieldTx)

	// r.GET("/statusbyinctx", APIGetStatusByIncTx)

	// r.GET("/statusbyservice", APIGetStatusByShieldService)

	r.POST("/statusbyinctxs", APIGetStatusByIncTxs)

	//papps
	r.POST("/estimateswapfee", APIEstimateSwapFee)

	r.POST("/submitswaptx", APISubmitSwapTx)

	//admin
	adminGroup := r.Group("/admin")
	adminGroup.GET("/failedshieldtx")
	adminGroup.GET("/shieldstatus")
	adminGroup.GET("/unshieldstatus")

	err := r.Run("0.0.0.0:" + strconv.Itoa(cfg.Port))
	if err != nil {
		panic(err)
	}
}

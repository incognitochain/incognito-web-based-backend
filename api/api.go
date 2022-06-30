package api

import (
	"log"
	"strconv"
	"time"

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
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	r.GET("/tokenlist", APIGetSupportedToken)

	r.POST("/estimateshieldreward", APIEstimateReward)

	r.POST("/estimateunshieldfee", APIEstimateUnshield)

	r.GET("/statusbyservice", APIGetStatusByShieldService)

	r.POST("/genunshieldaddress", APIGenUnshieldAddress)

	r.POST("/genshieldaddress", APIGenShieldAddress)

	r.POST("/submitunshieldtx", APISubmitUnshieldTx)

	r.GET("/statusbyinctx", APIGetStatusByIncTx)

	err := r.Run("0.0.0.0:" + strconv.Itoa(cfg.Port))
	if err != nil {
		panic(err)
	}
}

package api

import (
	"log"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/incognitochain/go-incognito-sdk-v2/wallet"
	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/patrickmn/go-cache"
)

var config common.Config
var incFeeKeySet *wallet.KeyWallet

func StartAPIservice(cfg common.Config) {
	log.Println("initiating api-service...")
	config = cfg
	cachedb = cache.New(5*time.Minute, 5*time.Minute)
	network := config.NetworkID
	if cfg.IncKey != "" {
		err := loadOTAKey(cfg.IncKey)
		if err != nil {
			panic(err)
		}
	}

	err := initIncClient(network)
	if err != nil {
		panic(err)
	}

	go cacheVaultState()
	go cacheSupportedPappsTokens()
	go cacheTokenList()
	go cacheBridgeNetworkInfos()

	r := gin.Default()

	r.Use(cors.Default())

	r.Use(gzip.Gzip(gzip.DefaultCompression))

	r.GET("/tokenlist", APIGetSupportedToken)

	r.POST("/estimateshieldreward", APIEstimateReward)

	r.POST("/estimateunshieldfee", APIEstimateUnshield)

	r.POST("/genunshieldaddress", APIGenUnshieldAddress)

	r.POST("/genshieldaddress", APIGenShieldAddress)

	r.POST("/submitunshieldtx", APISubmitUnshieldTx)

	r.POST("/submitshieldtx", APISubmitShieldTx)

	r.POST("/shieldstatus", APIGetShieldStatus)

	//papps
	pAppsGroup := r.Group("/papps")

	pAppsGroup.POST("/estimateswapfee", APIEstimateSwapFee)

	pAppsGroup.POST("/submitswaptx", APISubmitSwapTx)

	pAppsGroup.POST("/swapstatus", APIGetSwapTxStatus)

	pAppsGroup.GET("/getvaultstate", APIGetVaultState)

	pAppsGroup.GET("/getsupportedtokens", APIGetSupportedToken)

	//admin
	adminGroup := r.Group("/admin")
	adminGroup.GET("/failedshieldtx", APIGetFailedShieldTx)
	adminGroup.GET("/pendingshieldtx", APIGetPendingShieldTx)
	adminGroup.GET("/retryshield")
	adminGroup.GET("/retrievenetworksfee", APIGetNetworksFee)

	err = r.Run("0.0.0.0:" + strconv.Itoa(cfg.Port))
	if err != nil {
		panic(err)
	}
}

func loadOTAKey(key string) error {
	wl, err := wallet.Base58CheckDeserialize(key)
	if err != nil {
		return err
	}
	if wl.KeySet.OTAKey.GetOTASecretKey() == nil {

		return err
	}
	incFeeKeySet = wl
	return nil
}

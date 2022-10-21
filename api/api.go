package api

import (
	"errors"
	"log"
	"strconv"
	"sync"
	"time"

	gincache "github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
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
	store := persistence.NewInMemoryStore(time.Second)

	go cacheVaultState()
	go cacheSupportedPappsTokens()
	go cacheTokenList()
	go cacheBridgeNetworkInfos()

	r := gin.Default()

	r.Use(cors.Default())

	r.Use(gzip.Gzip(gzip.DefaultCompression))

	r.GET("/tokenlist", gincache.CachePage(store, 5*time.Second, APIGetSupportedToken))
	r.POST("/tokeninfo", gincache.CachePage(store, 5*time.Second, APIGetSupportedTokenInfo))

	r.POST("/estimateshieldreward", APIEstimateReward)

	r.POST("/estimateunshieldfee", APIEstimateUnshield)

	r.POST("/genunshieldaddress", APIGenUnshieldAddress)

	r.POST("/genshieldaddress", APIGenShieldAddress)

	r.POST("/submitunshieldtx", APISubmitUnshieldTx)

	r.GET("/validaddress", APIValidateAddress)

	r.POST("/submitshieldtx", APISubmitShieldTx) //depercated

	r.POST("/shieldstatus", APIGetShieldStatus)

	//papps
	pAppsGroup := r.Group("/papps")

	pAppsGroup.POST("/estimateswapfee", APIEstimateSwapFee)

	pAppsGroup.POST("/submitswaptx", APISubmitSwapTx)

	pAppsGroup.POST("/swapstatus", APIGetSwapTxStatus)

	pAppsGroup.GET("/getvaultstate", APIGetVaultState)

	pAppsGroup.GET("/getsupportedtokens", gincache.CachePage(store, 5*time.Second, APIGetSupportedToken))

	//admin
	adminGroup := r.Group("/admin")
	adminGroup.GET("/failedshieldtx", APIGetFailedShieldTx)
	adminGroup.GET("/pendingshieldtx", APIGetPendingShieldTx)
	adminGroup.GET("/pendingswaptx", APIGetPendingSwapTx)
	adminGroup.POST("/retryshield", gincache.CachePage(store, time.Minute, APIRetryShieldTx))
	adminGroup.POST("/retryswaptx", gincache.CachePage(store, time.Minute, APIRetrySwapTx))
	adminGroup.GET("/retrievenetworksfee", APIGetNetworksFee)
	adminGroup.GET("/getsupportedtokens", APIGetSupportedTokenInternal)

	go prefetchSupportedTokenList()

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

var supportTokenList []PappSupportedTokenData
var childToUnifiedTokenLock sync.RWMutex
var childToUnifiedTokenMap map[string]string

func prefetchSupportedTokenList() {
	childToUnifiedTokenMap = make(map[string]string)
	for {
		tokenList, err := retrieveTokenList()
		if err != nil {
			log.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}
		childToUnifiedTokenLock.Lock()
		for _, tk := range tokenList {
			if tk.CurrencyType == common.UnifiedCurrencyType {
				for _, ctk := range tk.ListUnifiedToken {
					childToUnifiedTokenMap[ctk.TokenID] = tk.TokenID
				}
			}
		}
		childToUnifiedTokenLock.Unlock()

		spTkList, err := getPappSupportedTokenList(tokenList)
		if err != nil {
			log.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}

		supportTokenList = spTkList

		time.Sleep(5 * time.Second)
	}
}

func getUnifiedTokenFromChildToken(childToken string) (string, error) {
	childToUnifiedTokenLock.RLock()
	defer childToUnifiedTokenLock.RUnlock()
	tkID, ok := childToUnifiedTokenMap[childToken]
	if !ok {
		return "", errors.New("can't find child token")
	}
	return tkID, nil
}

func getSupportedTokenList() []PappSupportedTokenData {
	result := []PappSupportedTokenData{}
	result = append(result, supportTokenList...)
	return result

}

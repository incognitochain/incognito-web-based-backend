package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
)

// func APIGetTokenList(c *gin.Context) {
// 	var responseBodyData APIRespond
// 	_, err := restyClient.R().
// 		EnableTrace().
// 		SetHeader("Content-Type", "application/json").
// 		SetResult(&responseBodyData).
// 		Get(config.CoinserviceURL + "/coins/tokenlist")
// 	if err != nil {
// 		c.JSON(400, gin.H{"Error": err.Error()})
// 		return
// 	}
// 	c.JSON(200, responseBodyData)
// }

func APIGetSupportedToken(c *gin.Context) {

	tokenList, err := retrieveTokenList()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	pappTokens, err := getPappSupportedTokenList(tokenList)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	var result []wcommon.TokenInfo
	dupChecker := make(map[string]struct{})

	for _, tk := range tokenList {
		if _, exist := dupChecker[tk.TokenID]; !exist {
			if tk.CurrencyType == wcommon.UnifiedCurrencyType {
				tk.IsSwapable = true
				newUTkList := []wcommon.TokenInfo{}
				for _, utk := range tk.ListUnifiedToken {
					var swapContractID string
					if wcommon.IsNativeCurrency(utk.CurrencyType) {
						swapContractID = "0x0000000000000000000000000000000000000000"
					} else {
						netID, err := wcommon.GetNetworkIDFromCurrencyType(utk.CurrencyType)
						if err == nil {
							swapContractID, err = getSwapContractID(tk.TokenID, netID, pappTokens)
							if err != nil {
								swapContractID, err = getSwapContractID(utk.TokenID, netID, pappTokens)
								if err != nil {
									log.Println(err)
								}
							}
						}
					}
					if swapContractID != "" {
						utk.IsSwapable = true
						utk.ContractIDSwap = swapContractID
					}
					newUTkList = append(newUTkList, utk)
				}
				tk.ListUnifiedToken = newUTkList
			} else {
				var swapContractID string
				if wcommon.IsNativeCurrency(tk.CurrencyType) {
					swapContractID = "0x0000000000000000000000000000000000000000"
				} else {
					netID, err := wcommon.GetNetworkIDFromCurrencyType(tk.CurrencyType)
					if err == nil {
						swapContractID, err = getSwapContractID(tk.TokenID, netID, pappTokens)
						if err != nil {
							log.Println(err)
						}
					}
				}
				if swapContractID != "" {
					tk.IsSwapable = true
					tk.ContractIDSwap = swapContractID
				}
			}
			result = append(result, tk)
			dupChecker[tk.TokenID] = struct{}{}
		}

	}

	var response struct {
		Result interface{}
		Error  interface{}
	}
	response.Result = result

	c.JSON(200, response)
}

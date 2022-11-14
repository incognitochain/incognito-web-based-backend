package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/incognitochain/incognito-web-based-backend/common"
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

func APIGetSupportedTokenInternal(c *gin.Context) {
	spTkList := getSupportedTokenList()
	var response struct {
		Result interface{}
		Error  interface{}
	}
	response.Result = spTkList

	c.JSON(200, response)
}

func APIGetSupportedTokenInfo(c *gin.Context) {
	var req APITokenInfoRequest
	err := c.MustBindWith(&req, binding.JSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	tokenList, err := getTokensInfo(req.TokenIDs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	pappTokens := getSupportedTokenList()

	var result []wcommon.TokenInfo
	dupChecker := make(map[string]struct{})

	for _, tk := range tokenList {
		// if !tk.Verified {
		// 	continue
		// }
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
						utk.ContractID = swapContractID
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
					tk.ContractID = swapContractID
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

func APIGetSupportedToken(c *gin.Context) {
	tokenList, err := retrieveTokenList()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	pappTokens := getSupportedTokenList()

	var result []wcommon.TokenInfo
	dupChecker := make(map[string]struct{})

	for _, tk := range tokenList {
		// if !tk.Verified {
		// 	continue
		// }
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
						utk.ContractID = swapContractID
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
					tk.ContractID = swapContractID
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

func APIGetDefaultTokenList(c *gin.Context) {
	var response struct {
		Result interface{}
		Error  interface{}
	}
	response.Result = defaultTokenList

	c.JSON(200, response)
}

func whiteListCurrencyType(currencyType int) bool {
	switch currencyType {
	case common.TOMO, common.ZIL, common.XMR, common.NEO, common.DASH, common.LTC, common.DOGE, common.ZEC, common.DOT, common.ETH, common.NEAR, common.AVAX, common.AURORA_ETH, common.BNB_BSC, common.MATIC, common.FTM, common.UNIFINE_TOKEN:
		return true
	default:
		return false
	}
}

var whiteListTokenContract map[string]struct{}

func parseDefaultToken() error {
	whiteListTokenContract = make(map[string]struct{})
	tokenList := []TokenStruct{}
	bscList := []TokenStruct{}
	err := json.Unmarshal([]byte(bscDefault), &bscList)
	if err != nil {
		return err
	}

	ethList := []TokenStruct{}
	err = json.Unmarshal([]byte(ethDefault), &ethList)
	if err != nil {
		return err
	}

	plgList := []TokenStruct{}
	err = json.Unmarshal([]byte(plgDefault), &plgList)
	if err != nil {
		return err
	}

	ftmList := []TokenStruct{}
	err = json.Unmarshal([]byte(ftmDefault), &ftmList)
	if err != nil {
		return err
	}

	tokenList = append(tokenList, bscList...)
	tokenList = append(tokenList, ethList...)
	tokenList = append(tokenList, plgList...)
	tokenList = append(tokenList, ftmList...)

	for _, token := range tokenList {
		whiteListTokenContract[strings.ToLower(token.ID)] = struct{}{}
	}
	fmt.Println("tokenList", len(whiteListTokenContract))

	return nil
}

func APISearchToken(c *gin.Context) {
	searchStr := c.Query("token")
	searchStr = strings.ToLower(searchStr)
	var result []wcommon.TokenInfo
	for _, tokenInfo := range allTokenList {
		strLen := len(searchStr)
		if strLen == 64 {
			if strings.Contains(strings.ToLower(tokenInfo.TokenID), searchStr) {
				result = append(result, tokenInfo)
				break
			}
		} else {
			if strLen <= 5 {
				if strings.Contains(strings.ToLower(tokenInfo.Symbol), searchStr) {
					result = append(result, tokenInfo)
				} else {
					if strings.Contains(strings.ToLower(tokenInfo.Name), searchStr) {
						result = append(result, tokenInfo)
					}
				}
			} else {
				if strings.Contains(strings.ToLower(tokenInfo.Name), searchStr) {
					result = append(result, tokenInfo)
				}
			}
		}
	}
	var response struct {
		Result interface{}
		Error  interface{}
	}

	if len(result) == 0 {
		response.Error = fmt.Errorf("not found")
		c.JSON(http.StatusBadRequest, response)
	} else {
		response.Result = result
		c.JSON(200, response)
	}
}

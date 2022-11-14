package api

import (
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/incognitochain/incognito-web-based-backend/common"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
)

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
		err = preCalcDefaultTokenList(tokenList, supportTokenList)
		if err != nil {
			log.Println(err)
			continue
		}
		err = preCalcAllTokenList(tokenList, supportTokenList)
		if err != nil {
			log.Println(err)
			continue
		}
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

var defaultTokenList []wcommon.TokenInfo
var allTokenList []wcommon.TokenInfo

func preCalcDefaultTokenList(tokenList []wcommon.TokenInfo, pappTokens []PappSupportedTokenData) error {
	var result []wcommon.TokenInfo
	dupChecker := make(map[string]struct{})

	for _, tk := range tokenList {
		if !tk.Verified {
			continue
		}
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
				if common.IsNativeCurrency(tk.CurrencyType) {
					swapContractID = "0x0000000000000000000000000000000000000000"
				} else {
					netID, err := common.GetNetworkIDFromCurrencyType(tk.CurrencyType)
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
			if config.NetworkID == "mainnet" {
				if tk.DefaultPairToken == "" {
					if !(whiteListCurrencyType(tk.CurrencyType)) {
						if _, exist := whiteListTokenContract[strings.ToLower(tk.ContractID)]; !exist {
							continue
						}
					}
				}
			}
			result = append(result, tk)
			dupChecker[tk.TokenID] = struct{}{}
		}
	}

	defaultTokenList = result
	return nil
}

func preCalcAllTokenList(tokenList []wcommon.TokenInfo, pappTokens []PappSupportedTokenData) error {
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
						utk.ContractID = swapContractID
						utk.IsSwapable = true
						utk.ContractIDSwap = swapContractID
					}
					newUTkList = append(newUTkList, utk)
				}
				tk.ListUnifiedToken = newUTkList
			} else {
				var swapContractID string
				if common.IsNativeCurrency(tk.CurrencyType) {
					swapContractID = "0x0000000000000000000000000000000000000000"
				} else {
					netID, err := common.GetNetworkIDFromCurrencyType(tk.CurrencyType)
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

	allTokenList = result
	return nil
}

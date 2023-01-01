package api

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"sync"

	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/interswap"
)

func PappsEstimator(ctx context.Context, req wcommon.EstimateSwapRequest) (*wcommon.EstimateSwapRespond, error) {
	amount := new(big.Float)
	amount, errBool := amount.SetString(req.Amount)
	if !errBool {
		return nil, errors.New("invalid amount")
	}

	switch req.Network {
	case "inc", "eth", "bsc", "plg", "ftm", "aurora", "avax":
	default:
		// c.JSON(http.StatusBadRequest, gin.H{"Error": })
		return nil, errors.New("unsupported network")
	}

	_, ok := new(big.Float).SetString(req.Amount)
	if !ok {
		// c.JSON(http.StatusBadRequest, gin.H{"Error":})
		return nil, errors.New("Amount isn't a valid number")
	}

	_, ok = new(big.Float).SetString(req.Slippage)
	if !ok {
		// c.JSON(http.StatusBadRequest, gin.H{"Error": })
		return nil, errors.New("Slippage isn't a valid number")
	}

	slippage, err := verifySlippage(req.Slippage)
	if err != nil {
		// c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return nil, err
	}

	networkID := wcommon.GetNetworkID(req.Network)
	if networkID == -1 {
		// c.JSON(http.StatusBadRequest, gin.H{"Error": })
		return nil, errors.New("invalid network")
	}

	var result wcommon.EstimateSwapRespond
	result.Networks = make(map[string]interface{})
	result.NetworksError = make(map[string]interface{})

	// estimate with Interswap
	if !req.IsFromInterswap {
		fmt.Println("Starting estimate interswap")
		interSwapParams := &interswap.EstimateSwapParam{
			Network:   req.Network,
			Amount:    req.Amount,
			Slippage:  req.Slippage,
			FromToken: req.FromToken,
			ToToken:   req.ToToken,
			ShardID:   req.ShardID,
		}

		interSwapRes, err := interswap.EstimateSwap(interSwapParams, config)
		if err != nil {
			result.NetworksError[interswap.InterSwapStr] = err.Error()
			fmt.Println("Estimate interswap with err", err)
		} else {
			for k, v := range interSwapRes {
				result.Networks[k] = v
			}
		}
	}
	fmt.Println("Finish estimate interswap")

	tkFromInfo, err := getTokenInfo(req.FromToken)
	if err != nil {
		// c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return nil, err
	}

	tkToInfo, err := getTokenInfo(req.ToToken)
	if err != nil {
		// c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return nil, err
	}
	tkToNetworkID := 0
	if tkToInfo.CurrencyType != wcommon.UnifiedCurrencyType {
		tkToNetworkID, _ = getNetworkIDFromCurrencyType(tkToInfo.CurrencyType)
	}

	networkErr := make(map[string]interface{})
	var pdexEstimate []QuoteDataResp

	if req.Network == "inc" {
		pdexresult := estimateSwapFeeWithPdex(req.FromToken, req.ToToken, req.Amount, slippage, tkFromInfo)
		if pdexresult != nil {
			pdexEstimate = append(pdexEstimate, *pdexresult)
		}
	}

	var resultLock sync.Mutex
	var wg sync.WaitGroup

	supportedNetworks := []int{}
	outofVaultNetworks := []int{}
	supportedOutNetworks := []int{}
	if tkFromInfo.CurrencyType == wcommon.UnifiedCurrencyType {
		dm := new(big.Float)
		dm.SetFloat64(math.Pow10(tkFromInfo.PDecimals))
		amountUint64, _ := amount.Mul(amount, dm).Uint64()

		for _, v := range tkFromInfo.ListUnifiedToken {
			if networkID == 0 {
				//check all vaults
				isEnoughVault, err := checkEnoughVault(req.FromToken, v.TokenID, amountUint64)
				if err != nil {
					// c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
					return nil, err
				}
				if isEnoughVault {
					supportedOutNetworks = append(supportedOutNetworks, v.NetworkID)
				} else {
					outofVaultNetworks = append(outofVaultNetworks, v.NetworkID)
					networkErr[wcommon.GetNetworkName(v.NetworkID)] = "not enough token in vault"
				}
			} else {
				//check 1 vault only
				if networkID == v.NetworkID {
					isEnoughVault, err := checkEnoughVault(req.FromToken, v.TokenID, amountUint64)
					if err != nil {
						// c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
						// return
						return nil, err
					}
					if isEnoughVault {
						supportedOutNetworks = append(supportedOutNetworks, v.NetworkID)
					} else {
						outofVaultNetworks = append(outofVaultNetworks, v.NetworkID)
						networkErr[wcommon.GetNetworkName(v.NetworkID)] = "not enough token in vault"
					}
				}
			}
		}
		if len(supportedOutNetworks) == 0 {
			// c.JSON(http.StatusBadRequest, gin.H{"Error": })
			return nil, errors.New("The amount exceeds the swap limit. Please retry with smaller amount.")
		}
		fmt.Println("pass check vault", "supportedOutNetworks", supportedOutNetworks)

		// check supported to token
		for _, spNetID := range supportedOutNetworks {
			if tkToInfo.CurrencyType == wcommon.UnifiedCurrencyType {
				for _, v := range tkToInfo.ListUnifiedToken {
					if v.NetworkID == spNetID {
						supportedNetworks = append(supportedNetworks, spNetID)
					}
				}
			} else {
				if tkToNetworkID == spNetID {
					supportedNetworks = append(supportedNetworks, spNetID)
				}
			}
		}
	} else {
		tkFromNetworkID, _ := getNetworkIDFromCurrencyType(tkFromInfo.CurrencyType)
		if tkFromNetworkID > 0 {
			if networkID == tkFromNetworkID {
				supportedOutNetworks = append(supportedOutNetworks, tkFromNetworkID)
			} else {
				if networkID == 0 {
					supportedOutNetworks = append(supportedOutNetworks, tkFromNetworkID)
				} else {
					// c.JSON(http.StatusBadRequest, gin.H{"Error": })
					return nil, errors.New("No supported networks found")
				}
			}

			// check supported to token
			for _, spNetID := range supportedOutNetworks {
				if tkToInfo.CurrencyType == wcommon.UnifiedCurrencyType {
					for _, v := range tkToInfo.ListUnifiedToken {
						if v.NetworkID == spNetID {
							supportedNetworks = append(supportedNetworks, spNetID)
						}
					}
				} else {
					if tkToNetworkID == spNetID {
						supportedNetworks = append(supportedNetworks, spNetID)
					}
				}
			}
		}
	}
	if len(supportedNetworks) == 0 {
		for net, v := range networkErr {
			result.NetworksError[net] = v
		}
		if req.Network == "inc" && len(pdexEstimate) != 0 {
			result.Networks["inc"] = pdexEstimate
		}
		if len(result.Networks) > 0 {
			return &result, nil
			// response.Result = result
			// c.JSON(200, response)
			// return
		}
		// response.Error = NotTradeable.Error()
		// c.JSON(http.StatusBadRequest, response)
		return nil, NotTradeable
	}

	networksInfo, err := getBridgeNetworkInfos()
	if err != nil {
		return nil, err
	}

	tokenList, err := retrieveTokenList()
	if err != nil {
		return nil, err
	}

	spTkList, err := getPappSupportedTokenList(tokenList)
	if err != nil {
		return nil, err
	}

	networkFees, err := database.DBRetrieveFeeTable()
	if err != nil {
		return nil, err
	}
	for _, network := range supportedNetworks {
		wg.Add(1)
		go func(net int) {
			data, err := estimateSwapFee(req.FromToken, req.ToToken, req.Amount, net, spTkList, networksInfo, networkFees, tkFromInfo, slippage)
			resultLock.Lock()
			if err != nil {
				networkErr[wcommon.GetNetworkName(net)] = err.Error()
			} else {
				result.Networks[wcommon.GetNetworkName(net)] = data
			}
			resultLock.Unlock()
			wg.Done()
		}(network)
	}
	wg.Wait()

	for net, v := range networkErr {
		result.NetworksError[net] = v
	}

	if len(pdexEstimate) != 0 {
		result.Networks["inc"] = pdexEstimate
	}
	if len(result.Networks) == 0 && len(pdexEstimate) == 0 {
		return &result, NotTradeable
	}
	return &result, nil
}

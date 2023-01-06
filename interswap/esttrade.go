package interswap

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/incognitochain/go-incognito-sdk-v2/coin"
	"github.com/incognitochain/incognito-web-based-backend/common"
	beCommon "github.com/incognitochain/incognito-web-based-backend/common"
)

type InterSwapPath struct {
	Paths     []*QuoteData
	FromToken string
	ToToken   string
	MidToken  string
	TotalFee  PappNetworkFee
}

type InterSwapEstRes struct {
	QuoteData
	PathType  int
	FromToken string
	ToToken   string
	MidToken  string
	MidOTA    string
	TotalFee  PappNetworkFee
	Details   []*QuoteData
}

type AddOnSwapInfo struct {
	AppName      string
	CallContract string
	FromToken    string
	ToToken      string

	MinExpectedAmt uint64
}

func IsOnlySwappablePDexToken(tokenInfo common.TokenInfo) bool {
	if tokenInfo.TokenID == common.PRV_TOKENID {
		return true
	}
	for _, currencyType := range common.OnlyPDexTokenCurrency {
		if tokenInfo.CurrencyType == currencyType {
			return true
		}
	}

	return false
}

func heuristicPathType(params *EstimateSwapParam, config common.Config, tokenInfos map[string]*beCommon.TokenInfo) (int, map[string]*beCommon.TokenInfo) {
	fromTokenInfo, tokenInfos, err := getTokenInfoWithCache(params.FromToken, tokenInfos, config)
	toTokenInfo, tokenInfos, err2 := getTokenInfoWithCache(params.ToToken, tokenInfos, config)
	if err != nil || err2 != nil {
		return -1, tokenInfos
	}

	// both fromToken & toToken can only trade on pDEX => can't estimate with Interswap
	if IsOnlySwappablePDexToken(*toTokenInfo) && IsOnlySwappablePDexToken(*fromTokenInfo) {
		return -1, tokenInfos
	}

	if params.Network != IncNetworkStr {
		if IsOnlySwappablePDexToken(*toTokenInfo) {
			return -1, tokenInfos
		}
		return PdexToPApp, tokenInfos
	}

	// fromToken: PRV, centralized token,
	if IsOnlySwappablePDexToken(*fromTokenInfo) {
		return PdexToPApp, tokenInfos
	}

	if IsOnlySwappablePDexToken(*toTokenInfo) {
		return PAppToPdex, tokenInfos
	}

	return PdexToPApp, tokenInfos
}

// InterSwap estimate swap
func EstimateSwap(params *EstimateSwapParam, config common.Config) (map[string][]InterSwapEstRes, error) {
	// validation

	time1 := time.Now()
	// * don't estimate inter swap if:
	//    there is one of tokens is mid tokens
	if IsMidTokens(params.FromToken) || IsMidTokens(params.ToToken) {
		return nil, nil
	}

	// get token infos
	tokenInfos := map[string]*beCommon.TokenInfo{}
	tokenIDs := []string{params.FromToken, params.ToToken}
	tokenIDs = append(tokenIDs, SupportedMidTokens...)
	tmp, err := getTokensInfo(tokenIDs, config)
	if err != nil {
		return nil, err
	}

	for _, i := range tmp {
		tokenInfos[i.TokenID] = &i
	}
	log.Printf("HHH Time1: %v", time.Since(time1).Seconds())

	// optimize
	estPathType, tokenInfos := heuristicPathType(params, config, tokenInfos)
	log.Printf("Estimate path type: %v\n", estPathType)
	estParamNetwork1 := IncNetworkStr
	estParamNetwork2 := params.Network
	var expectedPath1 int
	var expectedPath2 int

	switch estPathType {
	case PdexToPApp:
		{
			estParamNetwork1 = common.NETWORK_PDEX
			expectedPath1 = pDEXType
			expectedPath2 = pAppType
		}
	case PAppToPdex:
		{
			estParamNetwork2 = common.NETWORK_PDEX
			expectedPath1 = pAppType
			expectedPath2 = pDEXType
		}
	case -1:
		{
			return nil, nil
		}
	}
	log.Printf("HHH Time2: %v", time.Since(time1).Seconds())

	// available trade paths for Interswap
	paths := []InterSwapPath{}
	for _, midToken := range SupportedMidTokens {
		// estimate fromToken => midToken
		p1 := &EstimateSwapParam{
			Network:   estParamNetwork1,
			Amount:    params.Amount,
			FromToken: params.FromToken,
			ToToken:   midToken,
			Slippage:  params.Slippage,
		}
		p1Bytes, _ := json.Marshal(p1)
		fmt.Printf("Param 1: %s\n", string(p1Bytes))

		est1, err := CallEstimateSwap(p1, config, "")
		if err != nil {
			continue
		}

		bytes, _ := json.Marshal(est1)
		fmt.Printf("Est 1: %+v\n", bytes)

		log.Printf("HHH Time3: %v", time.Since(time1).Seconds())

		bestPath1 := GetBestRoute(est1.Networks, expectedPath1)
		bytes, _ = json.Marshal(bestPath1)
		fmt.Printf("bestPath1: %+v\n", string(bytes))
		log.Printf("HHH Time4: %v", time.Since(time1).Seconds())

		if params.Network != IncNetworkStr {
			if bestPath1[IncNetworkStr] == nil {
				// not found the suitable path in this case
				continue
			}
		}

		for network1, swapInfo1 := range bestPath1 {
			p2 := &EstimateSwapParam{
				Network:   estParamNetwork2,
				Amount:    swapInfo1.AmountOut,
				FromToken: midToken,
				ToToken:   params.ToToken,
				Slippage:  params.Slippage,
			}
			p2Bytes, _ := json.Marshal(p2)
			fmt.Printf("Param 2: %s\n", string(p2Bytes))

			// Estimate for addon swap tx
			est2, err := CallEstimateSwap(p2, config, "")
			if err != nil {
				continue
			}
			log.Printf("HHH Time5: %v", time.Since(time1).Seconds())

			bytes, _ := json.Marshal(est2)
			fmt.Printf("Est 2: %+v\n", bytes)

			bestPath2 := GetBestRoute(est2.Networks, expectedPath2)
			bytes, _ = json.Marshal(bestPath2)
			fmt.Printf("bestPath2: %+v\n", bytes)
			log.Printf("HHH Time6: %v", time.Since(time1).Seconds())

			swapInfo2 := new(QuoteData)
			switch network1 {
			case IncNetworkStr:
				{
					swapInfo2 = bestPath2[PAppStr]
				}
			default:
				{
					if params.Network != IncNetworkStr {
						continue
					}
					swapInfo2 = bestPath2[IncNetworkStr]
				}
			}

			if swapInfo2 == nil {
				continue
			}
			// the addon swap must pay fee in midToken
			if len(swapInfo2.Fee) == 0 || swapInfo2.Fee[0].TokenID != midToken {
				continue
			}

			path := InterSwapPath{
				Paths:     []*QuoteData{swapInfo1, swapInfo2},
				MidToken:  midToken,
				FromToken: params.FromToken,
				ToToken:   params.ToToken,
			}
			paths = append(paths, path)
			log.Printf("HHH Time7: %v", time.Since(time1).Seconds())
		}
	}
	bytes, _ := json.Marshal(paths)
	fmt.Printf("paths: %v\n", string(bytes))
	if len(paths) == 0 {
		return nil, nil
	}

	// find the best path
	bestPath := new(InterSwapPath)
	var totalFee *PappNetworkFee
	for i := 0; i < len(paths); i++ {
		path := paths[i]

		totalFee, tokenInfos, err = calTotalFee(path, config, tokenInfos)
		if err != nil {
			log.Printf("Error calculate total fee: %v\n", err)
			continue
		}
		path.TotalFee = *totalFee

		if bestPath == nil || len(bestPath.Paths) == 0 {
			bestPath = &path
		} else if isBetter, err := isBetterInterSwapPath(path, *bestPath); err == nil && isBetter {
			bestPath = &path
		}
	}
	log.Printf("HHH Time8: %v", time.Since(time1).Seconds())

	bytes, _ = json.Marshal(bestPath)
	fmt.Printf("bestPath: %v\n", string(bytes))

	if len(bestPath.Paths) != 2 {
		return nil, errors.New("Interswap Invalid best path")
	}

	// deduct the fee of the second of tx from MinAcceptedAmt
	feeAddon := bestPath.Paths[1].Fee[0]
	feeAmountInBuyToken := feeAddon.AmountInBuyToken
	if feeAmountInBuyToken == "" {
		feeAddonTokenInfo, tokenInfos, err := getTokenInfoWithCache(feeAddon.TokenID, tokenInfos, config)
		if err != nil {
			return nil, err
		}
		tmp, err := convertAmountFromToTokenInfo(feeAddon.Amount, *feeAddonTokenInfo, *tokenInfos[bestPath.ToToken])
		if err != nil {
			return nil, err
		}
		feeAmountInBuyTokenFloat64 := divDec(tmp, uint64(tokenInfos[bestPath.ToToken].PDecimals))
		feeAmountInBuyToken = float64ToStr(feeAmountInBuyTokenFloat64)

	}
	amountOut, err := subStrs(bestPath.Paths[1].AmountOut, feeAmountInBuyToken)
	if err != nil {
		return nil, err
	}

	amountOutPreSlippage, err := subStrs(bestPath.Paths[1].AmountOutPreSlippage, feeAmountInBuyToken)
	if err != nil {
		return nil, err
	}
	amountOutFloat64, err := strToFloat64(amountOut)
	if err != nil {
		return nil, err
	}
	amountOutRawTmp, err := convertToDecAmtWithTokenInfo(amountOutFloat64, *tokenInfos[bestPath.ToToken])
	if err != nil {
		return nil, err
	}
	amountOutRaw := float64ToStr(float64(amountOutRawTmp))

	rate, err := divStrs(amountOutPreSlippage, params.Amount)
	if err != nil {
		return nil, err
	}

	midOTA := ""
	if params.ShardID != "" {
		otaReceiver := &coin.OTAReceiver{}
		keyWallet := InterswapIncKeySets[params.ShardID]
		if keyWallet == nil {
			return nil, errors.New("Invalid shardID")
		}
		err = otaReceiver.FromAddress(keyWallet.KeySet.PaymentAddress)
		if err != nil {
			return nil, err
		}
		midOTA = otaReceiver.String()
	}

	log.Printf("HHH Time9: %v", time.Since(time1).Seconds())

	// convert AmountInBuyToken
	feeRes := bestPath.Paths[0].Fee[0]
	feeResTokenInfo, tokenInfos, err := getTokenInfoWithCache(feeRes.TokenID, tokenInfos, config)
	if err != nil {
		return nil, err
	}
	tmp2, err := convertAmountFromToTokenInfo(feeRes.Amount, *feeResTokenInfo, *tokenInfos[bestPath.ToToken])
	if err != nil {
		return nil, err
	}
	feeAmountInBuyTokenFloat64 := divDec(tmp2, uint64(tokenInfos[bestPath.ToToken].PDecimals))
	feeAmountInBuyTokenRes := float64ToStr(feeAmountInBuyTokenFloat64)

	bestPath.Paths[0].Fee[0].AmountInBuyToken = feeAmountInBuyTokenRes

	swapInfo := InterSwapEstRes{
		// this object to get info to show on UI
		QuoteData: QuoteData{
			AppName:              InterSwapStr,
			AmountIn:             params.Amount,
			AmountInRaw:          bestPath.Paths[0].AmountInRaw,
			AmountOut:            amountOut,
			AmountOutRaw:         amountOutRaw,
			AmountOutPreSlippage: amountOutPreSlippage,
			Rate:                 rate,
			Fee:                  bestPath.Paths[0].Fee, // only show the fee of the first tx
			FeeAddress:           bestPath.Paths[0].FeeAddress,
			FeeAddressShardID:    bestPath.Paths[0].FeeAddressShardID,
			Paths:                "", // Frontend will build the path
			ImpactAmount:         "", // TODO
		},
		MidOTA:    midOTA,
		FromToken: bestPath.FromToken,
		ToToken:   bestPath.ToToken,
		MidToken:  bestPath.MidToken,
		TotalFee:  bestPath.TotalFee,
		// use the first item in the array to create the first tx
		Details: bestPath.Paths,
	}

	res := map[string][]InterSwapEstRes{
		InterSwapStr: []InterSwapEstRes{swapInfo},
	}
	return res, nil

}

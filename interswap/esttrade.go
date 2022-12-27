package interswap

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/go-incognito-sdk-v2/coin"
	"github.com/incognitochain/incognito-web-based-backend/common"
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

// InterSwap estimate swap
func EstimateSwap(params *EstimateSwapParam, config common.Config) (map[string]InterSwapEstRes, error) {
	// validation

	// * don't estimate inter swap if:
	//    there is one of tokens is mid tokens
	//    network is different to "inc"
	if IsMidTokens(params.FromToken) || IsMidTokens(params.ToToken) {
		return nil, nil
	}
	// if params.Network != IncNetworkStr {
	// 	return nil, nil
	// }

	paths := []InterSwapPath{}
	for _, midToken := range SupportedMidTokens {
		// estimate fromToken => midToken
		p1 := &EstimateSwapParam{
			Network:   IncNetworkStr,
			Amount:    params.Amount,
			FromToken: params.FromToken,
			ToToken:   midToken,
			Slippage:  params.Slippage,
		}
		p1Bytes, _ := json.Marshal(p1)
		fmt.Printf("Param 1: %s\n", string(p1Bytes))

		est1, err := CallEstimateSwap(p1, config)
		if err != nil {
			continue
		}
		bytes, _ := json.Marshal(est1)
		fmt.Printf("Est 1: %+v\n", bytes)

		bestPath1 := GetBestRoute(est1.Networks)
		bytes, _ = json.Marshal(bestPath1)
		fmt.Printf("bestPath1: %+v\n", string(bytes))

		if params.Network != IncNetworkStr {
			if bestPath1[IncNetworkStr] == nil {
				// not found the suitable path in this case
				continue
			}
		}

		for network1, swapInfo1 := range bestPath1 {
			p2 := &EstimateSwapParam{
				Network:   params.Network,
				Amount:    swapInfo1.AmountOut,
				FromToken: midToken,
				ToToken:   params.ToToken,
				Slippage:  params.Slippage,
			}
			p2Bytes, _ := json.Marshal(p2)
			fmt.Printf("Param 2: %s\n", string(p2Bytes))

			est2, err := CallEstimateSwap(p2, config)
			if err != nil {
				continue
			}

			bytes, _ := json.Marshal(est2)
			fmt.Printf("Est 2: %+v\n", bytes)

			bestPath2 := GetBestRoute(est2.Networks)
			bytes, _ = json.Marshal(bestPath2)
			fmt.Printf("bestPath2: %+v\n", bytes)

			switch network1 {
			case IncNetworkStr:
				{
					swapInfo2 := bestPath2[PAppStr]
					if swapInfo2 == nil {
						continue
					}
					path := InterSwapPath{
						Paths:     []*QuoteData{swapInfo1, swapInfo2},
						MidToken:  midToken,
						FromToken: params.FromToken,
						ToToken:   params.ToToken,
					}
					paths = append(paths, path)
				}
			default:
				{
					if params.Network != IncNetworkStr {
						continue
					}
					swapInfo2 := bestPath2[IncNetworkStr]
					if swapInfo2 == nil {
						continue
					}

					path := InterSwapPath{
						Paths:     []*QuoteData{swapInfo1, swapInfo2},
						MidToken:  midToken,
						FromToken: params.FromToken,
						ToToken:   params.ToToken,
					}
					paths = append(paths, path)
				}
			}
		}
	}
	bytes, _ := json.Marshal(paths)
	fmt.Printf("paths: %v\n", string(bytes))
	if len(paths) == 0 {
		return nil, nil
	}

	// find the best path
	bestPath := paths[0]
	for i := 1; i < len(paths); i++ {
		path := paths[i]

		totalFee, err := calTotalFee(path)
		if err != nil {
			fmt.Printf("Error cal fee: %v\n", err)
			continue
		}
		path.TotalFee = *totalFee

		if len(bestPath.Paths) == 0 {
			bestPath = path
		} else if isBetter, err := isBetterInterSwapPath(path, bestPath); err == nil && isBetter {
			bestPath = path
		}
	}

	bytes, _ = json.Marshal(bestPath)
	fmt.Printf("bestPath: %v\n", string(bytes))

	if len(bestPath.Paths) != 2 {
		return nil, errors.New("Interswap Invalid best path")
	}

	// deduct the fee of the second of tx from MinAcceptedAmt
	amountOut, err := subStrs(bestPath.Paths[1].AmountOut, bestPath.Paths[1].Fee[0].AmountInBuyToken)
	if err != nil {
		return nil, err
	}

	amountOutPreSlippage, err := subStrs(bestPath.Paths[1].AmountOutPreSlippage, bestPath.Paths[1].Fee[0].AmountInBuyToken)
	if err != nil {
		return nil, err
	}
	amountOutRaw, err := convertToDecAmtStr(amountOut, bestPath.ToToken)
	if err != nil {
		return nil, err
	}

	rate, err := divStrs(amountOutPreSlippage, params.Amount)
	if err != nil {
		return nil, err
	}

	otaReceiver := &coin.OTAReceiver{}
	err = otaReceiver.FromAddress(InterswapIncKeySet.KeySet.PaymentAddress)
	if err != nil {
		return nil, err
	}

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
			Paths:                "", // TODO
			ImpactAmount:         "", // TODO
		},
		MidOTA:    otaReceiver.String(),
		FromToken: bestPath.FromToken,
		ToToken:   bestPath.ToToken,
		MidToken:  bestPath.MidToken,
		TotalFee:  bestPath.TotalFee,
		// use the first item in the array to create the first tx
		Details: bestPath.Paths,
	}

	res := map[string]InterSwapEstRes{
		InterSwapStr: swapInfo,
	}
	return res, nil
}

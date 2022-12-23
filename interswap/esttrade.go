package interswap

import (
	"encoding/json"
	"fmt"
)

type InterSwapPath struct {
	Paths     []*QuoteData
	FromToken string
	ToToken   string
	MidToken  string
	TotalFee  PappNetworkFee
}

type InterSwapInfo struct {
	QuoteData
	FromToken string
	ToToken   string
	MidToken  string
	TotalFee  PappNetworkFee
	Details   []*QuoteData
}

// InterSwap estimate swap
func EstimateSwap(params *EstimateSwapParam) (map[string]InterSwapInfo, error) {
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
			Network:     IncNetworkStr,
			Amount:      params.Amount,
			FromToken:   params.FromToken,
			ToToken:     midToken,
			Slippage:    params.Slippage,
			IsInterswap: true,
		}
		p1Bytes, _ := json.Marshal(p1)
		fmt.Printf("Param 1: %s\n", string(p1Bytes))

		est1, err := CallEstimateSwap(p1)
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
				Network:     params.Network,
				Amount:      swapInfo1.AmountOut,
				FromToken:   midToken,
				ToToken:     params.ToToken,
				Slippage:    params.Slippage,
				IsInterswap: true,
			}
			p2Bytes, _ := json.Marshal(p2)
			fmt.Printf("Param 2: %s\n", string(p2Bytes))

			est2, err := CallEstimateSwap(p2)
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

					// bytes, _ = json.Marshal(path)
					// fmt.Printf("Path case 1: %+v\n", bytes)
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
					// bytes, _ = json.Marshal(path)
					// fmt.Printf("Path case 2: %+v\n", bytes)
					paths = append(paths, path)
				}
			}

		}

	}
	bytes, _ := json.Marshal(paths)
	fmt.Printf("paths: %v\n", string(bytes))

	// find the best path
	bestPath := InterSwapPath{}
	for _, path := range paths {

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

	swapInfo := InterSwapInfo{
		QuoteData: QuoteData{
			AppName:              InterSwapStr,
			AmountIn:             params.Amount,
			AmountInRaw:          bestPath.Paths[0].AmountInRaw,
			AmountOut:            bestPath.Paths[1].AmountOut,
			AmountOutRaw:         bestPath.Paths[1].AmountOutRaw,
			AmountOutPreSlippage: bestPath.Paths[1].AmountOutPreSlippage,
			Rate:                 "", // TODO
			Fee:                  []PappNetworkFee{bestPath.TotalFee},
			FeeAddress:           bestPath.Paths[0].FeeAddress,
			FeeAddressShardID:    bestPath.Paths[0].FeeAddressShardID,
			Paths:                "", // TODO
			ImpactAmount:         "", // TODO
		},
		FromToken: bestPath.FromToken,
		ToToken:   bestPath.ToToken,
		MidToken:  bestPath.MidToken,
		TotalFee:  bestPath.TotalFee,
		Details:   bestPath.Paths,
	}

	res := map[string]InterSwapInfo{
		InterSwapStr: swapInfo,
	}
	return res, nil
}

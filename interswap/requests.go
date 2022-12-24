package interswap

import (
	"errors"
	"fmt"
	"log"
)

type EstimateSwapParam struct {
	// FeeAddress  string
	// MultiTrades bool
	// MinSplit    int

	Network   string
	Amount    string
	Slippage  string
	FromToken string // IncTokenID
	ToToken   string // IncTokenID

	IsInterswap bool
}

type PappNetworkFee struct {
	TokenID          string  `json:"tokenid"`
	Amount           uint64  `json:"amount"`
	AmountInBuyToken string  `json:"amountInBuyToken"`
	PrivacyFee       uint64  `json:"privacyFee"`
	FeeInUSD         float64 `json:"feeInUSD"`
}

type QuoteData struct {
	AppName              string
	CallContract         string
	AmountIn             string
	AmountInRaw          string
	AmountOut            string
	AmountOutRaw         string
	AmountOutPreSlippage string
	RedepositReward      string
	Rate                 string
	Fee                  []PappNetworkFee
	FeeAddress           string
	FeeAddressShardID    int
	Paths                interface{}
	PathsContract        interface{}
	PoolPairs            []string
	Calldata             string
	ImpactAmount         string
	RouteDebug           interface{}
}

type EstimateSwapResult struct {
	Networks      map[string][]QuoteData
	NetworksError map[string]interface{}
}

type EstimateSwapResponse struct {
	Result EstimateSwapResult
	Error  interface{}
}

type SubmitpAppSwapTxRequest struct {
	TxRaw        string
	TxHash       string
	FeeRefundOTA string
	// FeeRefundAddress string  // NOTE: don't use this field
}

type SubmitpAppSwapTxResponse struct {
	Result map[string]interface{}
}

type TxStatusRespond struct {
	TxHash string
	Status string
	Error  string
}

// CallEstimateSwap call request to estimate swap
// for both pdex and papp
func CallEstimateSwap(params *EstimateSwapParam) (*EstimateSwapResult, error) {
	req := struct {
		Network     string
		Amount      string // without decimal
		FromToken   string // IncTokenID
		ToToken     string // IncTokenID
		Slippage    string
		IsInterswap bool
	}{
		Network:     params.Network,
		Amount:      params.Amount,
		FromToken:   params.FromToken,
		ToToken:     params.ToToken,
		Slippage:    params.Slippage,
		IsInterswap: params.IsInterswap,
	}

	estSwapResponse := EstimateSwapResponse{}

	fmt.Printf("APIEndpoint: %v\n", APIEndpoint)
	response, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").SetBody(req).
		SetResult(&estSwapResponse).
		Post(APIEndpoint + "/papps/estimateswapfee")
	if err != nil {
		err := fmt.Errorf("[ERR] Call API /papps/estimateswapfee request error: %v", err)
		log.Println(err)
		return nil, err
	}
	if response.StatusCode() != 200 {
		err := fmt.Errorf("[ERR] Call API /papps/estimateswapfee status code error: %v", response.StatusCode())
		log.Println(err)
		return nil, err
	}

	if estSwapResponse.Error != nil {
		err := fmt.Errorf("[ERR] Call API /papps/estimateswapfee response error: %v", err)
		log.Println(err)
		return nil, err
	}

	return &estSwapResponse.Result, nil
}

// CallSubmitPappSwapTx calls request to submit tx papp
func CallSubmitPappSwapTx(txRaw, txHash, feeRefundOTA string) (map[string]interface{}, error) {
	req := SubmitpAppSwapTxRequest{
		TxRaw:        txRaw,
		TxHash:       txHash,
		FeeRefundOTA: feeRefundOTA,
	}

	estSwapResponse := SubmitpAppSwapTxResponse{}

	fmt.Printf("APIEndpoint: %v\n", APIEndpoint)
	response, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").SetBody(req).
		SetResult(&estSwapResponse).
		Post(APIEndpoint + "/papps/submitswaptx")
	if err != nil {
		err := fmt.Errorf("[ERR] Call API /papps/submitswaptx request error: %v", err)
		log.Println(err)
		return nil, err
	}
	if response.StatusCode() != 200 {
		err := fmt.Errorf("[ERR] Call API /papps/submitswaptx status code error: %v", response.StatusCode())
		log.Println(err)
		return nil, err
	}

	// if estSwapResponse.Error != nil {
	// 	err := fmt.Errorf("[ERR] Call API /papps/submitswaptx response error: %v", err)
	// 	log.Println(err)
	// 	return nil, err
	// }

	return estSwapResponse.Result, nil
}

// // CallEstimateSwap call request to estimate swap
// // for both pdex and papp
// func CallGetPdexTxStatus(params *EstimateSwapParam) (*EstimateSwapResult, error) {
// 	req := struct {
// 		Network     string
// 		Amount      string // without decimal
// 		FromToken   string // IncTokenID
// 		ToToken     string // IncTokenID
// 		Slippage    string
// 		IsInterswap bool
// 	}{
// 		Network:     params.Network,
// 		Amount:      params.Amount,
// 		FromToken:   params.FromToken,
// 		ToToken:     params.ToToken,
// 		Slippage:    params.Slippage,
// 		IsInterswap: params.IsInterswap,
// 	}

// 	estSwapResponse := EstimateSwapResponse{}

// 	fmt.Printf("APIEndpoint: %v\n", APIEndpoint)
// 	response, err := restyClient.R().
// 		EnableTrace().
// 		SetHeader("Content-Type", "application/json").SetBody(req).
// 		SetResult(&estSwapResponse).
// 		Post(APIEndpoint + "/papps/estimateswapfee")
// 	if err != nil {
// 		err := fmt.Errorf("[ERR] Call API /papps/estimateswapfee request error: %v", err)
// 		log.Println(err)
// 		return nil, err
// 	}
// 	if response.StatusCode() != 200 {
// 		err := fmt.Errorf("[ERR] Call API /papps/estimateswapfee status code error: %v", response.StatusCode())
// 		log.Println(err)
// 		return nil, err
// 	}

// 	if estSwapResponse.Error != nil {
// 		err := fmt.Errorf("[ERR] Call API /papps/estimateswapfee response error: %v", err)
// 		log.Println(err)
// 		return nil, err
// 	}

// 	return &estSwapResponse.Result, nil
// }

// isBetterQuoteData returns true if d1 is better than d2
func isBetterQuoteData(d1 QuoteData, d2 QuoteData) (bool, error) {
	// calculate actual received amount 1
	if d1.Fee == nil || len(d1.Fee) == 0 {
		return false, errors.New("Invalid format Fee is empty")
	}
	amt1, err1 := strToFloat64(d1.AmountOutPreSlippage)
	fee1, err2 := strToFloat64(d1.Fee[0].AmountInBuyToken)
	if err1 != nil || err2 != nil {
		return false, errors.New("Invalid format AmountOutPreSlippage or feeAmountInBuyToken")
	}
	actualAmt1 := amt1 - fee1

	// calculate actual received amount 2
	if d2.Fee == nil || len(d2.Fee) == 0 {
		return false, errors.New("Invalid format Fee is empty")
	}
	amt2, err1 := strToFloat64(d2.AmountOutPreSlippage)
	fee2, err2 := strToFloat64(d2.Fee[0].AmountInBuyToken)
	if err1 != nil || err2 != nil {
		return false, errors.New("Invalid format AmountOutPreSlippage or feeAmountInBuyToken")
	}

	actualAmt2 := amt2 - fee2

	return actualAmt1 > actualAmt2, nil
}

// isBetterQuoteData returns true if d1 is better than d2
func isBetterInterSwapPath(path1 InterSwapPath, path2 InterSwapPath) (bool, error) {

	if len(path1.Paths) != 2 || len(path2.Paths) != 2 {
		return false, errors.New("Invalid format interswap path")
	}

	// calculate actual received amount 1
	amt1, err1 := strToFloat64(path1.Paths[1].AmountOutPreSlippage)
	fee1, err2 := strToFloat64(path1.TotalFee.AmountInBuyToken)
	if err1 != nil || err2 != nil {
		return false, errors.New("Invalid format AmountOutPreSlippage or feeAmountInBuyToken")
	}
	actualAmt1 := amt1 - fee1

	// calculate actual received amount 2
	amt2, err1 := strToFloat64(path2.Paths[1].AmountOutPreSlippage)
	fee2, err2 := strToFloat64(path2.TotalFee.AmountInBuyToken)
	if err1 != nil || err2 != nil {
		return false, errors.New("Invalid format AmountOutPreSlippage or feeAmountInBuyToken")
	}

	actualAmt2 := amt2 - fee2

	return actualAmt1 > actualAmt2, nil
}

// GetBestRoute return the best one for each network (pdex & papps), and the best one for all pApps
func GetBestRoute(paths map[string][]QuoteData) map[string]*QuoteData {
	res := map[string]*QuoteData{}
	bestPAppPath := QuoteData{}

	for network, datas := range paths {
		// find the best one for each network
		tmpBestPath := QuoteData{}
		for _, d := range datas {
			if tmpBestPath.AppName == "" {
				tmpBestPath = d
			} else if isBetter, err := isBetterQuoteData(d, tmpBestPath); err == nil && isBetter {
				tmpBestPath = d
			}
		}
		res[network] = &tmpBestPath

		// find the best one for all papp
		if network != IncNetworkStr {
			if bestPAppPath.AppName == "" {
				bestPAppPath = tmpBestPath
			} else if isBetter, err := isBetterQuoteData(tmpBestPath, bestPAppPath); err == nil && isBetter {
				bestPAppPath = tmpBestPath
			}
		}

	}
	if bestPAppPath.AppName != "" {
		res[PAppStr] = &bestPAppPath
	}

	return res
}

func calTotalFee(interswapPath InterSwapPath) (*PappNetworkFee, error) {
	path := interswapPath.Paths
	if len(path) != 2 || len(path[0].Fee) == 0 || len(path[1].Fee) == 0 {
		return nil, errors.New("Invalid path to calculate total fee")
	}

	fee1 := path[0].Fee[0]
	fee2 := path[1].Fee[0]

	// total fee paid in the token fee of the first swap info
	feeToken := fee1.TokenID
	amount := fee1.Amount + convertAmountUint64(fee2.Amount, fee2.TokenID, feeToken)
	amountInBuyToken := convertAmountUint64(amount, feeToken, interswapPath.ToToken)
	amountInBuyTokenStr := convertToWithoutDecStr(amountInBuyToken, interswapPath.ToToken)

	res := &PappNetworkFee{
		TokenID:          feeToken,
		Amount:           amount,
		AmountInBuyToken: amountInBuyTokenStr,
		// PrivacyFee: ,
		// FeeInUSD:
	}

	fmt.Printf("Calculate total fee : %+v\n", res)

	return res, nil
}

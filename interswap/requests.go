package interswap

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	"github.com/incognitochain/go-incognito-sdk-v2/metadata"
	"github.com/incognitochain/go-incognito-sdk-v2/privacy"
	beCommon "github.com/incognitochain/incognito-web-based-backend/common"
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
	ShardID   string
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
	PAppNetwork          string
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

type TransactionDetail struct {
	BlockHash   string `json:"BlockHash"`
	BlockHeight uint64 `json:"BlockHeight"`
	TxSize      uint64 `json:"TxSize"`
	Index       uint64 `json:"Index"`
	ShardID     byte   `json:"ShardID"`
	Hash        string `json:"Hash"`
	Version     int8   `json:"Version"`
	Type        string `json:"Type"` // Transaction type
	LockTime    string `json:"LockTime"`
	RawLockTime int64  `json:"RawLockTime,omitempty"`
	Fee         uint64 `json:"Fee"` // Fee applies: always consant
	Image       string `json:"Image"`

	IsPrivacy bool `json:"IsPrivacy"`
	// Proof           privacy.Proof `json:"Proof"`
	// ProofDetail     interface{}   `json:"ProofDetail"`
	InputCoinPubKey string `json:"InputCoinPubKey"`
	SigPubKey       string `json:"SigPubKey,omitempty"` // 64 bytes
	RawSigPubKey    []byte `json:"RawSigPubKey,omitempty"`
	Sig             string `json:"Sig,omitempty"` // 64 bytes

	Metadata                      string      `json:"Metadata"`
	CustomTokenData               string      `json:"CustomTokenData"`
	PrivacyCustomTokenID          string      `json:"PrivacyCustomTokenID"`
	PrivacyCustomTokenName        string      `json:"PrivacyCustomTokenName"`
	PrivacyCustomTokenSymbol      string      `json:"PrivacyCustomTokenSymbol"`
	PrivacyCustomTokenData        string      `json:"PrivacyCustomTokenData"`
	PrivacyCustomTokenProofDetail interface{} `json:"PrivacyCustomTokenProofDetail"`
	PrivacyCustomTokenIsPrivacy   bool        `json:"PrivacyCustomTokenIsPrivacy"`
	PrivacyCustomTokenFee         uint64      `json:"PrivacyCustomTokenFee"`

	IsInMempool bool `json:"IsInMempool"`
	IsInBlock   bool `json:"IsInBlock"`

	Info string `json:"Info"`
}

// CallEstimateSwap call request to estimate swap
// for both pdex and papp
func CallEstimateSwap(params *EstimateSwapParam, config beCommon.Config, endpoint string) (*EstimateSwapResult, error) {
	req := struct {
		Network         string
		Amount          string // without decimal
		FromToken       string // IncTokenID
		ToToken         string // IncTokenID
		Slippage        string
		IsFromInterswap bool
	}{
		Network:         params.Network,
		Amount:          params.Amount,
		FromToken:       params.FromToken,
		ToToken:         params.ToToken,
		Slippage:        params.Slippage,
		IsFromInterswap: true,
	}

	if endpoint == "" {
		endpoint = "http://localhost:" + strconv.Itoa(config.Port)
	}

	estSwapResponse := EstimateSwapResponse{}
	response, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").SetBody(req).
		SetResult(&estSwapResponse).
		Post(endpoint + "/papps/estimateswapfee")
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
func CallSubmitPappSwapTx(txRaw, txHash, feeRefundOTA string, config beCommon.Config, endpoint string) (map[string]interface{}, error) {
	log.Printf("CallSubmitPappSwapTx txHash: %v\n", txHash)
	req := SubmitpAppSwapTxRequest{
		TxRaw:        txRaw,
		TxHash:       txHash,
		FeeRefundOTA: feeRefundOTA,
	}

	estSwapResponse := SubmitpAppSwapTxResponse{}

	if endpoint == "" {
		endpoint = "http://localhost:" + strconv.Itoa(config.Port)
	}

	response, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").SetBody(req).
		SetResult(&estSwapResponse).
		Post(endpoint + "/papps/submitswaptx")
	if err != nil {
		err := fmt.Errorf("[ERR] Call API /papps/submitswaptx request error: %v", err)
		log.Println(err)
		return nil, err
	}

	if response.StatusCode() != 200 {
		err := fmt.Errorf("[ERR] Call API /papps/submitswaptx status code error: %v", response.StatusCode())
		log.Println(err)
		log.Printf("response: %+v\n", response)
		return nil, err
	}
	return estSwapResponse.Result, nil
}

type PappStatus struct {
	BurnStatus    string
	SwapStatus    string
	IsRedeposit   bool
	IsRedeposited bool
	BuyAmount     string
	Reward        string
}

// CallSubmitPappSwapTx calls request to submit tx papp
// func CallGetPappSwapTxStatus(txID string, config beCommon.Config) (*PappStatus, error) {
// 	req := struct {
// 		TxList []string
// 	}{
// 		TxList: []string{txID},
// 	}

// 	estSwapResponse := SubmitpAppSwapTxResponse{}

// 	response, err := restyClient.R().
// 		EnableTrace().
// 		SetHeader("Content-Type", "application/json").SetBody(req).
// 		SetResult(&estSwapResponse).
// 		Post("http://localhost:" + strconv.Itoa(config.Port) + "/papps/swapstatus")
// 	if err != nil {
// 		err := fmt.Errorf("[ERR] Call API /papps/swapstatus request error: %v", err)
// 		log.Println(err)
// 		return nil, err
// 	}
// 	if response.StatusCode() != 200 {
// 		err := fmt.Errorf("[ERR] Call API /papps/swapstatus status code error: %v", response.StatusCode())
// 		log.Println(err)
// 		return nil, err
// 	}

// 	m := estSwapResponse.Result[txID].(map[string]interface{})
// 	if len(m) == 0 {
// 		return nil, errors.New("[ERR] Call API /papps/swapstatus status not found")
// 	}
// 	if m["error"] != "" {
// 		return nil, fmt.Errorf("[ERR] Call API /papps/swapstatus status error %v", m["error"])
// 	}
// 	if m["swap_err"] != "" {
// 		return nil, fmt.Errorf("[ERR] Call API /papps/swapstatus swap error %v", m["swap_err"])
// 	}
// 	burnStatus := ""
// 	swapStatus := ""
// 	isRedeposit := false
// 	isRedeposited := false
// 	buyAmount := ""
// 	reward := ""

// 	if m["inc_request_tx_status"] != "" {
// 		burnStatus = m["inc_request_tx_status"].(string)
// 	}

// 	if m["swap_outcome"] != "" {
// 		swapStatus = m["swap_outcome"].(string)
// 	}

// 	if m["is_redeposit"] == true {
// 		isRedeposit = true
// 	}

// 	if m["redeposit_status"] == "success" {
// 		isRedeposited = true
// 	}

// 	if tmp, ok := m["swap_detail"].(map[string]interface{}); ok {
// 		if buyAmtTmp, ok := tmp["out_amount"].(string); ok {
// 			buyAmount = buyAmtTmp
// 		}
// 		if rewardTmp, ok := tmp["reward"].(string); ok {
// 			reward = rewardTmp
// 		}
// 	}

// 	pappStatus := PappStatus{
// 		BurnStatus:    burnStatus,
// 		SwapStatus:    swapStatus,
// 		IsRedeposit:   isRedeposit,
// 		IsRedeposited: isRedeposited,
// 		BuyAmount:     buyAmount,
// 		Reward:        reward,
// 	}

// 	// "inc_request_tx_status":

// 	return &pappStatus, nil
// }

func genRPCBody(method string, params []interface{}) interface{} {
	type RPC struct {
		ID      int           `json:"id"`
		JsonRPC string        `json:"jsonrpc"`
		Method  string        `json:"method"`
		Params  []interface{} `json:"params"`
	}

	req := RPC{
		ID:      1,
		JsonRPC: "1.0",
		Method:  method,
		Params:  params,
	}
	return req
}

func CallGetTxDetails(txhash string, config beCommon.Config) (*TransactionDetail, error) {
	reqRPC := genRPCBody("gettransactionbyhash", []interface{}{
		txhash,
	})

	type TxDetailRespond struct {
		Result TransactionDetail
		Error  *string
	}

	var responseBodyData TxDetailRespond
	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).SetBody(reqRPC).
		Post(config.FullnodeURL)
	if err != nil {
		return nil, err
	}

	if responseBodyData.Error != nil {
		return nil, fmt.Errorf("%v", responseBodyData.Error)
	}
	return &responseBodyData.Result, nil
}

func CallGetPdexSwapTxStatus(txhash string, config beCommon.Config) (bool, *beCommon.TradeDataRespond, error) {
	type APIRespond struct {
		Result []beCommon.TradeDataRespond
		Error  *string
	}

	var responseBodyData APIRespond

	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).
		Get(config.CoinserviceURL + "/pdex/v3/tradedetail?txhash=" + txhash)
	if err != nil {
		log.Println("CallGetPdexSwapTxStatus", err)
		return false, nil, err
	}
	if responseBodyData.Error != nil {
		log.Println("CallGetPdexSwapTxStatus", errors.New(*responseBodyData.Error))
		return false, nil, errors.New(*responseBodyData.Error)
	}

	if len(responseBodyData.Result) == 0 {
		return false, nil, errors.New("not found")
	}

	swapResult := responseBodyData.Result[0]
	return true, &swapResult, nil
}

type ReceivedTransactionV2 struct {
	TxDetail struct {
		BlockHash   string `json:"BlockHash"`
		BlockHeight uint64 `json:"BlockHeight"`
		TxSize      uint64 `json:"TxSize"`
		Index       uint64 `json:"Index"`
		ShardID     byte   `json:"ShardID"`
		Hash        string `json:"Hash"`
		Version     int8   `json:"Version"`
		Type        string `json:"Type"` // Transaction type
		LockTime    string `json:"LockTime"`
		Fee         uint64 `json:"Fee"` // Fee applies: always consant
		Image       string `json:"Image"`

		IsPrivacy bool          `json:"IsPrivacy"`
		Proof     privacy.Proof `json:"Proof"`
		// ProofDetail      jsonresult.ProofDetail `json:"ProofDetail"`
		InputCoinPubKey  string   `json:"InputCoinPubKey"`
		OutputCoinPubKey []string `json:"OutputCoinPubKey"`
		OutputCoinSND    []string `json:"OutputCoinSND"`

		TokenInputCoinPubKey  string   `json:"TokenInputCoinPubKey"`
		TokenOutputCoinPubKey []string `json:"TokenOutputCoinPubKey"`
		TokenOutputCoinSND    []string `json:"TokenOutputCoinSND"`

		SigPubKey string `json:"SigPubKey,omitempty"` // 64 bytes
		Sig       string `json:"Sig,omitempty"`       // 64 bytes

		Metatype                 int    `json:"Metatype"`
		Metadata                 string `json:"Metadata"`
		CustomTokenData          string `json:"CustomTokenData"`
		PrivacyCustomTokenID     string `json:"PrivacyCustomTokenID"`
		PrivacyCustomTokenName   string `json:"PrivacyCustomTokenName"`
		PrivacyCustomTokenSymbol string `json:"PrivacyCustomTokenSymbol"`
		PrivacyCustomTokenData   string `json:"PrivacyCustomTokenData"`
		// PrivacyCustomTokenProofDetail jsonresult.ProofDetail `json:"PrivacyCustomTokenProofDetail"`
		PrivacyCustomTokenIsPrivacy bool   `json:"PrivacyCustomTokenIsPrivacy"`
		PrivacyCustomTokenFee       uint64 `json:"PrivacyCustomTokenFee"`

		IsInMempool bool `json:"IsInMempool"`
		IsInBlock   bool `json:"IsInBlock"`

		Info string `json:"Info"`
	}
	FromShardID byte
}

// call through coin service
func CallGetTxsByCoinPubKeys(coinPubKeys []string, config beCommon.Config) ([]ReceivedTransactionV2, error) {
	type APIRespond struct {
		Result []ReceivedTransactionV2
		Error  *string
	}

	var responseBodyData APIRespond

	reqBody := struct {
		Pubkeys []string
		Base58  bool
	}{
		Pubkeys: coinPubKeys,
		Base58:  true,
	}

	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		SetResult(&responseBodyData).
		Post(config.CoinserviceURL + "/gettxsbypubkey")
	if err != nil {
		log.Println("CallGetTxsByCoinPubKeys", err)
		return nil, err
	}
	if responseBodyData.Error != nil {
		log.Println("CallGetTxsByCoinPubKeys", errors.New(*responseBodyData.Error))
		return nil, errors.New(*responseBodyData.Error)
	}

	if len(responseBodyData.Result) == 0 {
		return nil, errors.New("CallGetTxsByCoinPubKeys result empty")
	}

	return responseBodyData.Result, nil
}

func CallGetTxsByCoinPubKeys2(coinPubKeys []string, config beCommon.Config, incClient *incclient.IncClient) (map[string]map[string]metadata.Transaction, error) {
	// response := map[string]map[string]metadata.Transaction{}
	resMap, err := incClient.GetTransactionsByPublicKeys(coinPubKeys)
	if err != nil || resMap == nil {
		log.Println("CallGetTxsByCoinPubKeys", err)
		return nil, err
	}

	return resMap, nil

	// for pubKey, txs := range resMap {
	// 	for txID, tx := range txs {
	// 		txDetail, err := incClient.GetTxDetail(tx)
	// 		if err != nil || txDetail == nil {
	// 			continue
	// 		}

	// 		response[pubKey] = append(response[pubKey], txDetail)
	// 	}
	// }

	// return response, nil
}

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
		if network == IncNetworkStr {
			res[network] = &tmpBestPath
		} else {
			// find the best one for all papp
			tmpBestPath.PAppNetwork = network
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

func calTotalFee(interswapPath InterSwapPath, config beCommon.Config, tokenInfos map[string]*beCommon.TokenInfo) (*PappNetworkFee, map[string]*beCommon.TokenInfo, error) {
	path := interswapPath.Paths
	if len(path) != 2 || len(path[0].Fee) == 0 || len(path[1].Fee) == 0 {
		return nil, tokenInfos, errors.New("Invalid path to calculate total fee")
	}

	fee1 := path[0].Fee[0]
	fee2 := path[1].Fee[0]

	// total fee paid in the token fee of the first swap info
	feeToken := fee1.TokenID

	feeTokenInfo, tokenInfos, err := getTokenInfoWithCache(feeToken, tokenInfos, config)
	if err != nil {
		return nil, tokenInfos, err
	}

	feeTokenInfo2, tokenInfos, err := getTokenInfoWithCache(fee2.TokenID, tokenInfos, config)
	if err != nil {
		return nil, tokenInfos, err
	}

	toTokenInfo, tokenInfos, err := getTokenInfoWithCache(interswapPath.ToToken, tokenInfos, config)
	if err != nil {
		return nil, tokenInfos, err
	}
	// feeAmt2, err := convertAmountUint64(fee2.Amount, fee2.TokenID, feeToken, config)
	feeAmt2, err := convertAmountFromToTokenInfo(fee2.Amount, *feeTokenInfo2, *feeTokenInfo)
	if err != nil {
		return nil, tokenInfos, err
	}
	amount := fee1.Amount + feeAmt2
	// amountInBuyToken, err := convertAmountUint64(amount, feeToken, interswapPath.ToToken, config)
	amountInBuyToken, err := convertAmountFromToTokenInfo(amount, *feeTokenInfo, *toTokenInfo)
	if err != nil {
		return nil, tokenInfos, err
	}
	amountInBuyTokenStr, err := convertToWithoutDecStrWithTokenInfo(amountInBuyToken, toTokenInfo)
	if err != nil {
		return nil, tokenInfos, err
	}

	res := &PappNetworkFee{
		TokenID:          feeToken,
		Amount:           amount,
		AmountInBuyToken: amountInBuyTokenStr,
		// PrivacyFee: ,
		// FeeInUSD:
	}

	fmt.Printf("Calculate total fee : %+v\n", res)

	return res, tokenInfos, nil
}

func getTokenInfo(pUTokenID string, config beCommon.Config) (*beCommon.TokenInfo, error) {
	type APIRespond struct {
		Result []beCommon.TokenInfo
		Error  *string
	}

	reqBody := struct {
		TokenIDs []string
	}{
		TokenIDs: []string{pUTokenID},
	}

	var responseBodyData APIRespond
	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).SetBody(reqBody).
		Post(config.CoinserviceURL + "/coins/tokeninfo")
	if err != nil {
		return nil, err
	}

	if len(responseBodyData.Result) == 1 {
		return &responseBodyData.Result[0], nil
	}
	return nil, errors.New(fmt.Sprintf("token not found"))
}

func getTokensInfo(pUTokenID []string, config beCommon.Config) ([]beCommon.TokenInfo, error) {
	type APIRespond struct {
		Result []beCommon.TokenInfo
		Error  *string
	}

	reqBody := struct {
		TokenIDs []string
	}{
		TokenIDs: pUTokenID,
	}

	var responseBodyData APIRespond
	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).SetBody(reqBody).
		Post(config.CoinserviceURL + "/coins/tokeninfo")
	if err != nil {
		return nil, err
	}

	if len(responseBodyData.Result) == 0 {
		return nil, errors.New(fmt.Sprintf("tokens not found"))
	}
	return responseBodyData.Result, nil
}

// GetChildTokenUnified returns child token of unified token
// if token is not unified, return itself
func GetChildTokenUnified(tokenID string, networkID int, config beCommon.Config) (string, error) {
	tokenInfo, err := getTokenInfo(tokenID, config)
	if err != nil {
		return "", err
	}

	listUnified := tokenInfo.ListUnifiedToken
	if len(listUnified) == 0 {
		fmt.Printf("%v is not unified token\n", tokenID)
		return tokenID, nil
	}

	for _, token := range listUnified {
		if token.NetworkID == networkID {
			return token.TokenID, nil
		}
	}

	return "", errors.New("Invalid networkID")
}

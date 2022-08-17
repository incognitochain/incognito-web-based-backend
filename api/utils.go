package api

import (
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/mongodb/mongo-tools/common/json"
)

var incClient *incclient.IncClient

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

func VerifyCaptcha(clientCaptcha string, secret string) (bool, error) {
	data := make(map[string]string)
	data["response"] = clientCaptcha
	data["secret"] = secret

	re, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").SetHeader("Authorization", "Bearer "+usa.token).SetFormData(data).
		Post("https://hcaptcha.com/siteverify")
	if err != nil {
		return false, err
	}

	var responseBodyData struct {
		Success bool `json:"success"`
	}

	err = json.Unmarshal(re.Body(), &responseBodyData)
	if err != nil {
		return false, err
	}

	return responseBodyData.Success, nil
}

func initIncClient(network string) error {
	var err error
	switch network {
	case "mainnet":
		incClient, err = incclient.NewMainNetClient()
	case "testnet-2": // testnet2
		incClient, err = incclient.NewTestNetClient()
	case "testnet-1":
		incClient, err = incclient.NewTestNet1Client()
	case "devnet":
		return errors.New("unsupported network")
	}
	if err != nil {
		return err
	}
	return nil
}

// convert nano coin to nano token: ex: 2000000000000000 (nano eth) => 2000000 (nano pETH)
func ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(amountStr string, decimal, pDecimals int64) uint64 {
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return 0
	}
	if decimal == pDecimals {
		return amount.Uint64()
	}

	pTokenAmount := new(big.Int).Mul(amount, big.NewInt(int64(math.Pow10(int(pDecimals)))))
	tokenAmount := new(big.Int).Div(pTokenAmount, big.NewInt(int64(math.Pow10(int(decimal)))))

	return tokenAmount.Uint64()
}

func ConvertNanoIncogTokenToOutChainToken(amountStr string, decimal, pDecimals int64) uint64 {
	amount, ok := new(big.Int).SetString(amountStr, 10)
	if !ok {
		return 0
	}
	if decimal == pDecimals {
		return amount.Uint64()
	}

	pTokenAmount := new(big.Int).Mul(amount, big.NewInt(int64(math.Pow10(int(decimal)))))
	fmt.Println("* decimal: ", pTokenAmount.String())

	tokenAmount := new(big.Int).Div(pTokenAmount, big.NewInt(int64(math.Pow10(int(pDecimals)))))
	fmt.Println("* pdecimal: ", tokenAmount.String())
	return tokenAmount.Uint64()
}

func getpTokenContractID(tokenID string, networkID int, supportedTokenList []PappSupportedTokenData) (*PappSupportedTokenData, error) {
	for _, v := range supportedTokenList {
		if v.ID == tokenID && v.NetworkID == networkID {
			return &v, nil
		}
	}
	return nil, errors.New("can't find contractID for token " + tokenID)
}

func uniswapDataExtractor(data []byte) (*UniswapQuote, error) {
	if len(data) == 0 {
		return nil, errors.New("can't extract data from empty byte array")
	}
	result := UniswapQuote{}
	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	if result.Message != "ok" {
		return nil, errors.New(result.Message)
	}
	return &result, nil
}

func pancakeDataExtractor(data []byte) (*PancakeQuote, error) {
	if len(data) == 0 {
		return nil, errors.New("can't extract data from empty byte array")
	}
	result := PancakeQuote{}
	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	if result.Message != "ok" {
		return nil, errors.New(result.Message)
	}
	return &result, nil
}

func getNativeTokenData(tokenList []PappSupportedTokenData, nativeTokenCurrencyType int) (*PappSupportedTokenData, error) {
	for _, token := range tokenList {
		if token.CurrencyType == nativeTokenCurrencyType {
			return &token, nil
		}
	}
	return nil, errors.New("token not found")
}

func checkEnoughVault(unifiedTokenID string, tokenID string, amount uint64) (bool, error) {
	methodRPC := "bridgeaggEstimateFeeByExpectedAmount"

	reqRPC := genRPCBody(methodRPC, []interface{}{
		map[string]interface{}{
			"UnifiedTokenID": unifiedTokenID,
			"TokenID":        tokenID,
			"ExpectedAmount": amount,
		},
	})

	var responseBodyData struct {
		ID     int `json:"Id"`
		Result *struct {
			ReceivedAmount uint64 `json:"ReceivedAmount"`
			BurntAmount    uint64 `json:"BurntAmount"`
			Fee            uint64 `json:"Fee"`
			MaxFee         uint64 `json:"MaxFee"`
			MaxBurntAmount uint64 `json:"MaxBurntAmount"`
		} `json:"Result"`
		Error *struct {
			Code       int    `json:"Code"`
			Message    string `json:"Message"`
			StackTrace string `json:"StackTrace"`
		} `json:"Error"`
	}
	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).SetBody(reqRPC).
		Post(config.FullnodeURL)
	if err != nil {
		return false, err
	}
	if responseBodyData.Error != nil {
		return false, errors.New(responseBodyData.Error.Message)
	}
	if responseBodyData.Result.Fee > 0 {
		return false, nil
	}
	return true, nil
}

func getNetworkIDFromCurrencyType(currencyType int) (int, error) {
	netID, ok := common.NetworkCurrencyMap[currencyType]
	if !ok {
		return 0, errors.New("unsupported network")
	}
	return netID, nil
}

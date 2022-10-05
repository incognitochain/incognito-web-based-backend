package papps

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	pancakeproxy "github.com/incognitochain/incognito-web-based-backend/papps/pancake"
	"github.com/incognitochain/incognito-web-based-backend/papps/pcurve"
	puniswap "github.com/incognitochain/incognito-web-based-backend/papps/puniswapproxy"
)

func PancakeQuote(tokenIn, tokenOut, amount, chainId, tokenInSymbol, tokenOutSymbol string, tokenInDecimal, tokenOutDecimal int, exactIn bool, endpoint string, tokenList string) ([]byte, error) {
	url := "http://" + endpoint + "/api/pancake/get-best-rate"
	method := "POST"

	payloadText := fmt.Sprintf(`{
	"sourceToken": {
		"contractIdGetRate":"%v",
		"decimals":%v,
		"symbol":"%v"
	},
	"destToken":{
		"contractIdGetRate":"%v",
		"decimals":%v,
		"symbol":"%v"
	},
	"isSwapFromBuyToSell": %v,
	"amount": "%v",
	"chainId": "%v",
	"listDecimals":%v
}`, strings.ToLower(tokenIn), tokenInDecimal, tokenInSymbol, strings.ToLower(tokenOut), tokenOutDecimal, tokenOutSymbol, exactIn, amount, chainId, tokenList)

	payload := strings.NewReader(payloadText)

	fmt.Println()
	fmt.Println("payload", payloadText)
	fmt.Println()

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return body, nil
}

func UniswapQuote(tokenIn, tokenOut, amount, chainId string, exactIn bool, endpoint string) ([]byte, error) {
	url := "http://" + endpoint + "/api/quote"
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf(`{
		"tokenIn": "%v",
		"tokenOut": "%v",
		"amount": "%v",
		"exactIn": true,
		"minSplits": 0,
		"protocols": "v3",
		"router": "alpha",
		"chainId": "%v",
		"debug": false
	}`, tokenIn, tokenOut, amount, chainId))

	if !exactIn {
		payload = strings.NewReader(fmt.Sprintf(`{
			"tokenIn": "%v",
			"tokenOut": "%v",
			"amount": "%v",
			"exactOut": true,
			"minSplits": 0,
			"protocols": "v3",
			"router": "alpha",
			"chainId": "%v",
			"debug": false
		}`, tokenIn, tokenOut, amount, chainId))
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return body, nil
}

func CurveQuote(
	evmClient *ethclient.Client,
	srcQty *big.Int,
	i *big.Int,
	j *big.Int,
	curvePool common.Address,
) (*big.Int, error) {
	c, err := pcurve.NewPcurvehelper(curvePool, evmClient)
	if err != nil {
		return nil, err
	}
	amountOut, err := c.GetDyUnderlying(nil, i, j, srcQty)
	if err != nil {
		return nil, err
	}

	return amountOut, nil
}

func BuildCallDataUniswap(paths []common.Address, recipient common.Address, fees []int64, srcQty *big.Int, expectedOut *big.Int, isNativeOut bool) (string, error) {
	var result string
	var input []byte
	var err error

	tradeAbi, err := abi.JSON(strings.NewReader(puniswap.PuniswapMetaData.ABI))
	if err != nil {
		return result, err
	}

	if len(fees) > 1 {
		agr := &puniswap.ISwapRouter2ExactInputParams{
			Path:             buildPathUniswap(paths, fees),
			Recipient:        recipient,
			AmountIn:         srcQty,
			AmountOutMinimum: expectedOut,
		}

		agrBytes, _ := json.MarshalIndent(agr, "", "\t")
		log.Println("ISwapRouter2ExactInputParams", isNativeOut, paths[0].String(), paths[1].String(), string(agrBytes))

		input, err = tradeAbi.Pack("tradeInput", agr, isNativeOut)
	} else {
		agr := &puniswap.ISwapRouter2ExactInputSingleParams{
			TokenIn:           paths[0],
			TokenOut:          paths[len(paths)-1],
			Fee:               big.NewInt(fees[0]),
			Recipient:         recipient,
			AmountIn:          srcQty,
			SqrtPriceLimitX96: big.NewInt(0),
			AmountOutMinimum:  expectedOut,
		}
		agrBytes, _ := json.MarshalIndent(agr, "", "\t")
		log.Println("ISwapRouter2ExactInputSingleParams", isNativeOut, string(agrBytes))

		input, err = tradeAbi.Pack("tradeInputSingle", agr, isNativeOut)
	}
	result = hex.EncodeToString(input)
	return result, err
}

func buildPathUniswap(paths []common.Address, fees []int64) []byte {
	var temp []byte
	for i := 0; i < len(fees); i++ {
		temp = append(temp, paths[i].Bytes()...)
		fee, err := hex.DecodeString(fmt.Sprintf("%06x", fees[i]))
		if err != nil {
			return nil
		}
		temp = append(temp, fee...)
	}
	temp = append(temp, paths[len(paths)-1].Bytes()...)

	return temp
}

func BuildCallDataPancake(paths []common.Address, srcQty *big.Int, expectedOut *big.Int, isNative bool) (string, error) {
	var result string
	var input []byte
	var err error

	tradeAbi, err := abi.JSON(strings.NewReader(pancakeproxy.PancakeproxyMetaData.ABI))
	if err != nil {
		return result, err
	}
	deadline := uint(time.Now().Unix() + 259200000)
	input, err = tradeAbi.Pack("trade", paths, srcQty, expectedOut, big.NewInt(int64(deadline)), isNative)
	if err != nil {
		return result, err
	}
	result = hex.EncodeToString(input)

	return result, err
}

func BuildCurveCallData(
	srcQty *big.Int,
	expectOutputAmount *big.Int,
	i *big.Int,
	j *big.Int,
	curvePool common.Address) (string, error) {

	var result string
	var input []byte

	isUnderlying := true
	tradeAbi, err := abi.JSON(strings.NewReader(pcurve.PcurveMetaData.ABI))
	if err != nil {
		return result, err
	}

	if isUnderlying {
		input, err = tradeAbi.Pack("exchangeUnderlying", i, j, srcQty, expectOutputAmount, curvePool)
	} else {
		input, err = tradeAbi.Pack("exchange", i, j, srcQty, expectOutputAmount, curvePool)
	}

	result = hex.EncodeToString(input)

	return result, err
}

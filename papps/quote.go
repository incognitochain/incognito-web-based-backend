package papps

import (
	"encoding/hex"
	"encoding/json"
	"errors"
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
	"github.com/incognitochain/bridge-eth/bridge/blur"
	"github.com/incognitochain/bridge-eth/bridge/pnft"
	pancakeproxy "github.com/incognitochain/incognito-web-based-backend/papps/pancake"
	"github.com/incognitochain/incognito-web-based-backend/papps/pcurve"
	"github.com/incognitochain/incognito-web-based-backend/papps/popensea"
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

func DecodePancakeCalldata(inputHex string) (*PancakeDecodeData, error) {
	tradeAbi, err := abi.JSON(strings.NewReader(pancakeproxy.PancakeproxyMetaData.ABI))
	if err != nil {
		return nil, err
	}
	decodeData, err := hex.DecodeString(inputHex)
	if err != nil {
		log.Fatal(err)
	}
	if method, ok := tradeAbi.Methods["trade"]; ok {
		params := make(map[string]interface{})
		err := method.Inputs.UnpackIntoMap(params, decodeData[4:])
		if err != nil {
			return nil, err
		}

		result := PancakeDecodeData{
			AmountOutMin: params["amountOutMin"].(*big.Int),
			SrcQty:       params["srcQty"].(*big.Int),
			Deadline:     params["deadline"].(*big.Int),
			Path:         params["path"].([]common.Address),
		}
		return &result, nil
	}
	return nil, errors.New("invalid abi")
}

func DecodeUniswapCalldata(inputHex string) (*UniswapDecodeData, error) {
	tradeAbi, err := abi.JSON(strings.NewReader(puniswap.PuniswapMetaData.ABI))
	if err != nil {
		return nil, err
	}
	decodeData, err := hex.DecodeString(inputHex)
	if err != nil {
		log.Fatal(err)
	}
	if method, ok := tradeAbi.Methods["tradeInputSingle"]; ok {
		params := make(map[string]interface{})
		err := method.Inputs.UnpackIntoMap(params, decodeData[4:])
		if err != nil {
			log.Println(err)
		} else {
			dataBytes, _ := json.Marshal(params["params"])
			result := UniswapDecodeData{}
			err = json.Unmarshal(dataBytes, &result)
			return &result, err
		}

	}
	if method, ok := tradeAbi.Methods["tradeInput"]; ok {
		params := make(map[string]interface{})
		err := method.Inputs.UnpackIntoMap(params, decodeData[4:])
		if err != nil {
			return nil, err
		}
		dataBytes, _ := json.Marshal(params["params"])
		result := UniswapDecodeData{}
		err = json.Unmarshal(dataBytes, &result)
		return &result, err
	}
	return nil, errors.New("invalid abi")
}

func DecodeCurveCalldata(inputHex string) (*CurveDecodeData, error) {
	tradeAbi, err := abi.JSON(strings.NewReader(pcurve.PcurveMetaData.ABI))
	if err != nil {
		return nil, err
	}
	decodeData, err := hex.DecodeString(inputHex)
	if err != nil {
		log.Fatal(err)
	}
	if method, ok := tradeAbi.Methods["exchangeUnderlying"]; ok {
		params := make(map[string]interface{})
		err := method.Inputs.UnpackIntoMap(params, decodeData[4:])
		if err != nil {
			return nil, err
		}

		result := CurveDecodeData{
			Amount:    params["amount"].(*big.Int),
			MinAmount: params["minAmount"].(*big.Int),
			I:         params["i"].(*big.Int),
			J:         params["i"].(*big.Int),
			CurvePool: params["curvePool"].(common.Address),
		}
		return &result, nil
	}
	return nil, errors.New("invalid abi")
}

func BuildOpenSeaCalldata(nftDetal *popensea.NFTDetail, recipient string) (string, error) {
	sellorder := nftDetal.SeaportSellOrders[0]
	offerer := common.HexToAddress(sellorder.ProtocolData.Parameters.Offerer)
	zone := common.HexToAddress(sellorder.ProtocolData.Parameters.Zone)
	offerItemDB := sellorder.ProtocolData.Parameters.Offer[0]
	considerationItemsDB := sellorder.ProtocolData.Parameters.Consideration
	startAmount, _ := new(big.Int).SetString(offerItemDB.StartAmount, 10)
	endAmount, _ := new(big.Int).SetString(offerItemDB.EndAmount, 10)
	offerId, _ := new(big.Int).SetString(offerItemDB.IdentifierOrCriteria, 10)

	offerItem := popensea.OfferItem{
		ItemType:             uint8(offerItemDB.ItemType),
		Token:                common.HexToAddress(offerItemDB.Token),
		IdentifierOrCriteria: offerId,
		StartAmount:          startAmount,
		EndAmount:            endAmount,
	}
	considerations := []popensea.ConsiderationItem{}

	for _, consider := range considerationItemsDB {

		considerItem := popensea.ConsiderationItem{
			ItemType:  uint8(consider.ItemType),
			Token:     common.HexToAddress(consider.Token),
			Recipient: common.HexToAddress(consider.Recipient),
		}
		considerItem.IdentifierOrCriteria, _ = new(big.Int).SetString(consider.IdentifierOrCriteria, 10)
		considerItem.StartAmount, _ = new(big.Int).SetString(consider.StartAmount, 10)
		considerItem.EndAmount, _ = new(big.Int).SetString(consider.EndAmount, 10)

		considerations = append(considerations, considerItem)
	}

	startTime, _ := new(big.Int).SetString(sellorder.ProtocolData.Parameters.StartTime, 10)
	endTime, _ := new(big.Int).SetString(sellorder.ProtocolData.Parameters.EndTime, 10)
	salt, _ := new(big.Int).SetString(sellorder.ProtocolData.Parameters.Salt[2:], 16)

	signature, err := hex.DecodeString(sellorder.ProtocolData.Signature[2:])
	if err != nil {
		return "", err
	}
	advanceOrder := popensea.AdvancedOrder{
		Parameters: popensea.OrderParameters{
			Offerer:                         offerer,
			Zone:                            zone,
			Offer:                           []popensea.OfferItem{offerItem},
			Consideration:                   considerations,
			OrderType:                       uint8(sellorder.ProtocolData.Parameters.OrderType),
			StartTime:                       startTime,
			EndTime:                         endTime,
			ZoneHash:                        [32]byte{},
			Salt:                            salt,
			ConduitKey:                      toByte32(common.HexToHash(sellorder.ProtocolData.Parameters.ConduitKey).Bytes()),
			TotalOriginalConsiderationItems: big.NewInt(int64(len(considerations))),
		},
		Numerator:   big.NewInt(1),
		Denominator: big.NewInt(1),
		Signature:   signature,
		ExtraData:   []byte{},
	}
	fulfillerConduitKey := toByte32(common.HexToHash(sellorder.ProtocolData.Parameters.ConduitKey).Bytes())
	// address will receive nft
	recipientAddr := common.HexToAddress(recipient)

	iopenseaAbi, err := abi.JSON(strings.NewReader(popensea.IopenseaMetaData.ABI))
	if err != nil {
		return "", err
	}
	calldata, err := iopenseaAbi.Pack("fulfillAdvancedOrder", advanceOrder, []popensea.CriteriaResolver{}, fulfillerConduitKey, recipientAddr)
	if err != nil {
		return "", err
	}
	fmt.Println("build opensea calldata: " + hex.EncodeToString(calldata) + "\n")

	// final burncalldata
	openseaProxyAbi, _ := abi.JSON(strings.NewReader(popensea.OpenseaMetaData.ABI))
	calldata, err = openseaProxyAbi.Pack("forward", calldata)
	if err != nil {
		return "", err
	}
	fmt.Println("final burn calldata: " + hex.EncodeToString(calldata) + "\n")
	return hex.EncodeToString(calldata), nil
}

func DecodeOpenSeaCalldata() {}

func BuildpNFTBuyCalldata(sellInputs []pnft.Execution, proxyAddrStr string, recipientStr string) (string, error) {
	proxyAddr := common.HexToAddress(proxyAddrStr)
	recipient := common.HexToAddress(recipientStr)
	blurProxy, _ := abi.JSON(strings.NewReader(blur.BlurMetaData.ABI))
	// buyOrder := sellInput.Order
	// buyOrder.Trader = proxyAddr
	// buyOrder.Side = 0
	// buyInput := pnft.Input{
	// 	Order:       buyOrder,
	// 	BlockNumber: big.NewInt(0),
	// }
	// calldata, err := blurProxy.Pack("execute", sellInput, buyInput, recipient)
	// if err != nil {
	// 	return "", err
	// }
	// return hex.EncodeToString(calldata), nil
	for i, _ := range sellInputs {
		_buyOrder := sellInputs[i].Sell.Order
		_buyOrder.Trader = proxyAddr
		_buyOrder.Side = 0
		_buyInput := pnft.Input{
			Order:       _buyOrder,
			BlockNumber: big.NewInt(0),
		}
		sellInputs[i].Buy = _buyInput
	}
	calldata, err := blurProxy.Pack("bulkExecute", sellInputs, recipient)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(calldata), nil
}

func toByte32(s []byte) [32]byte {
	a := [32]byte{}
	copy(a[:], s)
	return a
}

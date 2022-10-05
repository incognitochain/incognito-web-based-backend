package tools

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-resty/resty/v2"
	"github.com/incognitochain/incognito-web-based-backend/api"
	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/papps"
)

var PLG_ChainID = "137"
var ETH_ChainID = "1"

func TestEstimateTrade(t *testing.T) {
	service_endpoint := "https://api-service.incognito.org"
	couldNotEsimatePairs := make(map[string]map[string]string)
	tokenToEstimate := make(map[string][]api.PappSupportedTokenData)

	tokenList, err := getPappSupportedTokenList(service_endpoint)
	if err != nil {
		t.Fatal(err)
	}

	curvePoolIndex, err := getCurvePoolIndex(service_endpoint)
	if err != nil {
		t.Fatal(err)
	}
	_ = curvePoolIndex
	for _, tk := range tokenList {
		tokenToEstimate[tk.Protocol] = append(tokenToEstimate[tk.Protocol], tk)
	}

	for app, tkList := range tokenToEstimate {
		estimateFailed := make(map[string]string)
		for _, tk := range tkList {
			if tk.NetworkID != 3 {
				continue
			}
			if !tk.Verify {
				continue
			}
			for _, tk2 := range tkList {
				if !tk2.Verify {
					continue
				}
				if tk.ContractID != tk2.ContractID {
					pairs := []string{(tk.Name + "-" + tk.ContractIDGetRate), (tk2.Name + "-" + tk2.ContractIDGetRate)}
					sort.Strings(pairs)
					pairJoin := strings.Join(pairs, "-")
					if _, ok := estimateFailed[pairJoin]; ok {
						continue
					}
					switch app {
					case "pancake":

					case "uniswap":
						data, err := papps.UniswapQuote(tk.ContractIDGetRate, tk2.ContractIDGetRate, "1", PLG_ChainID, true, "51.161.117.193:3051")
						if err != nil {
							log.Println("UniswapQuote", pairJoin, err)
							estimateFailed[tk.ContractIDGetRate+"-"+tk2.ContractIDGetRate] = err.Error()
							continue
						}
						quote, feePaths, err := uniswapDataExtractor(data)
						if err != nil {
							log.Println("uniswapDataExtractor", pairJoin, err)
							estimateFailed[pairJoin] = err.Error()
							continue
						}
						log.Printf("pair %v estimate ok with %v -> %v\n", pairJoin, quote.Data.AmountIn, quote.Data.AmountOut)
						_ = quote
						_ = feePaths
					case "curve":

					}
				}
			}
		}
		log.Printf("app %v tkList %v expected pairs %v estimateFailed %v \n", app, len(tkList), len(tkList)*(len(tkList)+1)/2, len(estimateFailed))

		couldNotEsimatePairs[app] = estimateFailed
	}
}

// USDT - Tether USD  - 076a4423fa20922526bd50b0d7b0dc1c593ce16e15ba141ede5fb5a28aa3f229  - 0xdAC17F958D2ee523a2206206994597C13D831ec7
// MATIC - Matic  - 26df4d1bca9fd1a8871a24b9b84fc97f3dd62ca8809975c6d971d1b79d1d9f31  - 0x7d1afa7b718fb893db30a3abc0cfc608aacfebb0
// ETH - Ethereum  - 3ee31eba6376fc16cadb52c8765f20b6ebff92c0b1c5ab5fc78c8c25703bb19e  - 0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2 0x3bcec99c756aa54432d620a2a7de6e52b3869fc0
// DAI - Dai Stablecoin  - 0d953a47a7a488cee562e64c80c25d3dbe29d3b477ccd2b54408c0553a93f126  - 0x6B175474E89094C44Da98b954EedeAC495271d0F
// USDC - USD Coin  - 545ef6e26d4d428b16117523935b6be85ec0a63e8c2afeb0162315eb0ce3d151  - 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48
// USDC - USD Coin  - 1ff2da446abfebea3ba30385e2ca99b0f0bbeda5c6371f4c23c939672b429a42  - 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48
// DAI - Dai Stablecoin  - 3f89c75324b46f13c7b036871060e641d996a24c09b3065835cb1d38b799d6c1  - 0x6B175474E89094C44Da98b954EedeAC495271d0F
// USDT - Tether USD  - 716fd1009e2a1669caacc36891e707bfdf02590f96ebd897548e8963c95ebac0  - 0xdAC17F958D2ee523a2206206994597C13D831ec7
// MATIC - Matic (Ethereum)  - dae027b21d8d57114da11209dce8eeb587d01adf59d4fc356a8be5eedc146859  - 0x7d1afa7b718fb893db30a3abc0cfc608aacfebb0
// WBTC - Wrapped BTC  - 2eed019bb725eac3509e8e0df912229d142e54219f5b9452b1f9194039b51d4c  - 0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599
// AAVE - Aave Token  - e2499b2b18d03e9995764630b5aa9616b44649ed3516b811e3030451686b034d  - 0x7fc66500c84a76ad7e9c93437bfc5ac33e2ddae9
// BAT - Basic Attention Token  - 1fe75e9afa01b85126370a1583c7af9f1a5731625ef076ece396fcc6584c2b44  - 0x0d8775f648430679a709e98d2b0cb6250d2887ef
// SHIT - ShitCoin  - 28fecd1e4b7aef27283fc3f7928e70359d9e4537a4823656f00174ac86a96a92  - 0xaa7FB1c8cE6F18d4fD4Aabb61A2193d4D441c54F
// ETH - ETH  - ffd8d42dc40a8d166ea4848baf8b5f6e912ad79875f4373070b59392b1756c8f  - null

func TestEstimateTrade2(t *testing.T) {
	// service_endpoint := "https://api-coinservice.incognito.org"
	// tokenToEstimate := []common.TokenInfo{}
	// tokenList, err := retrieveTokenList(service_endpoint)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// for _, token := range tokenList {
	// 	if token.Verified {
	// 		if (token.CurrencyType == 1 || token.CurrencyType == 3) && token.ContractID != "" && token.PriceUsd > 0 && token.IsBridge {
	// 			tokenToEstimate = append(tokenToEstimate, token)
	// 		}
	// 	}
	// }

	tokenToEstimate := []common.TokenInfo{{
		ContractID: "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
		Name:       "Wrapped ETH",
	}, {
		ContractID: "0xdAC17F958D2ee523a2206206994597C13D831ec7",
		Name:       "USDT",
	}, {
		ContractID: "0x7d1afa7b718fb893db30a3abc0cfc608aacfebb0",
		Name:       "MATIC",
	}, {
		ContractID: "0x6B175474E89094C44Da98b954EedeAC495271d0F",
		Name:       "DAI",
	}, {
		ContractID: "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599",
		Name:       "Wrapped BTC",
	}, {
		ContractID: "0x7fc66500c84a76ad7e9c93437bfc5ac33e2ddae9",
		Name:       "AAVE",
	}, {
		ContractID: "0x0d8775f648430679a709e98d2b0cb6250d2887ef",
		Name:       "BAT",
	}, {
		ContractID: "0xaa7FB1c8cE6F18d4fD4Aabb61A2193d4D441c54F",
		Name:       "SHIT",
	}}

	tokenToEstimateLen := len(tokenToEstimate)
	pairsToEstimate := tokenToEstimateLen * (tokenToEstimateLen + 1) / 2
	estimatedPairs := make(map[string]string)

	estimateFailed := 0
	start := time.Now()
	for _, tk := range tokenToEstimate {
		for _, tk2 := range tokenToEstimate {
			if tk.ContractID != tk2.ContractID {
				log.Printf("Pairs to estimate: %v | estimated pairs: %v | failed pairs: %v| time-passed: %v \n", pairsToEstimate, len(estimatedPairs), estimateFailed, time.Since(start))

				pairs := []string{(tk.Name + "-" + tk.ContractID), (tk2.Name + "-" + tk2.ContractID)}
				sort.Strings(pairs)
				pairJoin := strings.Join(pairs, "-")
				if _, ok := estimatedPairs[pairJoin]; ok {
					continue
				}
				data, err := papps.UniswapQuote(tk.ContractID, tk2.ContractID, "1", ETH_ChainID, true, "51.161.117.193:3050")
				if err != nil {
					log.Println("UniswapQuote", pairJoin, err)
					estimatedPairs[pairJoin] = err.Error()
					estimateFailed++
					estimateFailedToFile(fmt.Sprintln("UniswapQuote", pairJoin, err))
					continue
				}
				quote, feePaths, err := uniswapDataExtractor(data)
				if err != nil {
					log.Println("uniswapDataExtractor", pairJoin, err)
					estimatedPairs[pairJoin] = err.Error()
					estimateFailed++
					estimateFailedToFile(fmt.Sprintln("uniswapDataExtractor", pairJoin, err))
					continue
				}

				r := fmt.Sprintf("pair %v estimate ok with %v -> %v\n", pairJoin, quote.Data.AmountIn, quote.Data.AmountOut)
				log.Println(r)
				estimateSuccessToFile(r)
				estimatedPairs[pairJoin] = "ok"
				_ = quote
				_ = feePaths
			}
		}
	}

	log.Printf("app %v tkList %v expected pairs %v estimateFailed %v \n", "uniswap", tokenToEstimateLen, pairsToEstimate, estimateFailed)

}
func TestEstimateTradeCurve(t *testing.T) {
	service_endpoint := "https://api-service.incognito.org"
	tokenToEstimate := []api.PappSupportedTokenData{}
	app := "curve"
	tokenList, err := getPappSupportedTokenList(service_endpoint)
	if err != nil {
		t.Fatal(err)
	}

	curvePoolIndex, err := getCurvePoolIndex(service_endpoint)
	if err != nil {
		t.Fatal(err)
	}
	_ = curvePoolIndex
	for _, tk := range tokenList {
		if tk.Verify && tk.Protocol == app {
			tokenToEstimate = append(tokenToEstimate, tk)
		}
	}

	tokenToEstimateLen := len(tokenToEstimate)
	pairsToEstimate := tokenToEstimateLen * (tokenToEstimateLen + 1) / 2
	estimatedPairs := make(map[string]string)
	estimateFailed := 0
	start := time.Now()

	evmClient, err := ethclient.Dial("https://polygon-mainnet.infura.io/v3/9bc873177cf74a03a35739e45755a9ac")
	if err != nil {
		t.Fatal(err)
	}

	for _, tk := range tokenToEstimate {
		for _, tk2 := range tokenToEstimate {
			if tk.ContractID != tk2.ContractID {
				pairs := []string{(tk.Name + "-" + tk.ContractIDGetRate), (tk2.Name + "-" + tk2.ContractIDGetRate)}
				pairRaw := strings.Join(pairs, "-")
				sort.Strings(pairs)
				pairJoin := strings.Join(pairs, "-")
				if _, ok := estimatedPairs[pairJoin]; ok {
					continue
				}

				log.Printf("Pairs to estimate: %v | estimated pairs: %v | failed pairs: %v| time-passed: %v \n", pairsToEstimate, len(estimatedPairs), estimateFailed, time.Since(start))

				token1PoolIndex, curvePoolAddress1, err := getTokenCurvePoolIndex(tk.ContractIDGetRate, curvePoolIndex)
				if err != nil {
					log.Println("getTokenCurvePoolIndex", tk.ContractIDGetRate, err)
					estimateFailedToFile(fmt.Sprintln("getTokenCurvePoolIndex", pairRaw, err))
					estimatedPairs[pairJoin] = err.Error()
					estimateFailed++
					continue
				}
				token2PoolIndex, _, err := getTokenCurvePoolIndex(tk2.ContractIDGetRate, curvePoolIndex)
				if err != nil {
					log.Println("getTokenCurvePoolIndex2", tk2.ContractIDGetRate, err)
					estimateFailedToFile(fmt.Sprintln("getTokenCurvePoolIndex2", pairRaw, err))
					estimatedPairs[pairJoin] = err.Error()
					estimateFailed++
					continue
				}
				i := big.NewInt(int64(token1PoolIndex))
				j := big.NewInt(int64(token2PoolIndex))

				curvePool := ethcommon.HexToAddress(curvePoolAddress1)
				var amountOut *big.Int

				amountFloat := new(big.Float)
				amountFloat, _ = amountFloat.SetString("1")

				amountBigFloat := api.ConvertToNanoIncognitoToken(amountFloat, int64(tk.Decimals)) //amount *big.Float, decimal int64, return *big.Float
				// convert float to bigin:
				amountBigInt, _ := amountBigFloat.Int(nil)

				amountOut, err = papps.CurveQuote(evmClient, amountBigInt, i, j, curvePool)
				if err != nil {
					log.Println("CurveQuote err:", err)
					estimateFailedToFile(fmt.Sprintln("CurveQuote", pairRaw, err))
					estimatedPairs[pairJoin] = err.Error()
					estimateFailed++
					continue
				}
				estimatedPairs[pairJoin] = "ok"
				_ = amountOut
				r := fmt.Sprintf("pair %v estimate ok with %v -> %v\n", pairRaw, amountBigInt.String(), amountOut.String())
				log.Println(r)
				estimateSuccessToFile(r)

			}
		}
	}
	log.Printf("app %v tkList %v expected pairs %v estimateFailed %v \n", app, tokenToEstimateLen, pairsToEstimate, estimateFailed)

}

var restyClient = resty.New()

func getTokenCurvePoolIndex(contractID string, poolList []api.CurvePoolIndex) (int, string, error) {
	for _, v := range poolList {
		if v.DappTokenAddress == contractID {
			return v.CurveTokenIndex, v.CurvePoolAddress, nil
		}
	}
	return -1, "", errors.New("pool not found")
}

func getPappSupportedTokenList(service string) ([]api.PappSupportedTokenData, error) {
	var responseBodyData struct {
		Result []api.PappSupportedTokenData
		Error  *struct {
			Code    int
			Message string
		} `json:"Error"`
	}

	re, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		Get(service + "/trade/supported-tokens")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(re.Body(), &responseBodyData)
	if err != nil {
		return nil, err
	}

	if responseBodyData.Error != nil {
		return nil, errors.New(responseBodyData.Error.Message)
	}
	return responseBodyData.Result, nil
}

func uniswapDataExtractor(data []byte) (*api.UniswapQuote, [][]int64, error) {
	if len(data) == 0 {
		return nil, nil, errors.New("can't extract data from empty byte array")
	}
	feePaths := [][]int64{}
	result := api.UniswapQuote{}
	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, nil, err
	}
	if result.Message != "ok" {
		return nil, nil, errors.New(result.Error)
	}
	for _, route := range result.Data.Route {
		fees := []int64{}
		for _, path := range route {
			fees = append(fees, path.Fee)
		}
		feePaths = append(feePaths, fees)
	}
	return &result, feePaths, nil
}

func getCurvePoolIndex(endpoint string) ([]api.CurvePoolIndex, error) {
	var responseBodyData struct {
		Result []api.CurvePoolIndex
		Error  *struct {
			Code    int
			Message string
		} `json:"Error"`
	}

	re, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		Get(endpoint + "/trade/curve-pool-indices")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(re.Body(), &responseBodyData)
	if err != nil {
		return nil, err
	}

	if responseBodyData.Error != nil {
		return nil, errors.New(responseBodyData.Error.Message)
	}
	return responseBodyData.Result, nil

}

func retrieveTokenList(url string) ([]common.TokenInfo, error) {
	type APIRespond struct {
		Result []common.TokenInfo
		Error  *string
	}

	var responseBodyData APIRespond
	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).
		Get(url + "/coins/tokenlist")
	if err != nil {
		return nil, err
	}
	if responseBodyData.Error != nil {
		return nil, errors.New(*responseBodyData.Error)
	}
	return responseBodyData.Result, nil
}

func estimateFailedToFile(lg string) {
	writeToFile(lg, "failed.log")
}

func estimateSuccessToFile(lg string) {
	writeToFile(lg, "success.log")
}

func writeToFile(lg string, fileName string) {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	datawriter := bufio.NewWriter(file)

	_, _ = datawriter.WriteString(lg)

	datawriter.Flush()
	file.Close()
}

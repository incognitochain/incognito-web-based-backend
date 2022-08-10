package papps

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func PancakeQuote(tokenIn, tokenOut, amount, chainId string, exactIn bool, endpoint string) ([]byte, error) {
	url := "http://" + endpoint + "/api/pancake/get-best-rate"
	method := "POST"

	payload := strings.NewReader(fmt.Sprintf(`{
		"sourceToken": "%v",
		"destToken": "%v",
		"amount": "%v",
		"chainId": "%v",
		"debug": false
	}`, tokenIn, tokenOut, amount, chainId))

	if !exactIn {
		payload = strings.NewReader(fmt.Sprintf(`{
			"sourceToken": "%v",
			"destToken": "%v",
			"amount": "%v",
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

func CurveQuote(tokenIn, tokenOut, amount, chainId string, endpoint string) ([]byte, error) {

	return nil, nil
}

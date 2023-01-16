package pblur

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	cloudflarebp "github.com/DaRealFreak/cloudflare-bp-go"
)

//https://core-api.prod.blur.io/v1/collections
func RetrieveCollectionList(OSEndpoint string, apiToken string, filters string) ([]CollectionDetail, error) {
	var respond struct {
		Collections []CollectionDetail `json:"collections"`
	}

	fmt.Println("filters: ", filters)

	url := fmt.Sprintf("%v/v1/collections?filters=%v", OSEndpoint, url.QueryEscape(filters))

	fmt.Println("url: ", url)

	client := &http.Client{}
	client.Transport = cloudflarebp.AddCloudFlareByPass(client.Transport)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("err0")
		return nil, err
	}

	req.Header.Add("Cookie", fmt.Sprintf("authToken=%s", apiToken))

	res, err := client.Do(req)
	if err != nil {
		log.Println("err1")
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("err2")
		return nil, err
	}
	// log.Println("body", string(body))
	err = json.Unmarshal(body, &respond)
	if err != nil {
		log.Println("err2")
		return nil, err
	}
	return respond.Collections, nil
}

// https://core-api.prod.blur.io/v1/collections/{slug}/tokens
func RetrieveCollectionAssets(OSEndpoint string, apiToken string, slug, filters string) ([]NFTDetail, error) {
	var respond struct {
		Assets []NFTDetail `json:"tokens"`
	}

	fmt.Println("filters: ", filters)
	url := fmt.Sprintf("%v/v1/collections/%v/tokens?filters=%v", OSEndpoint, slug, url.QueryEscape(filters))

	fmt.Println("url: ", url)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("Cookie", fmt.Sprintf("authToken=%s", apiToken))

	client := &http.Client{}
	client.Transport = cloudflarebp.AddCloudFlareByPass(client.Transport)

	res, _ := client.Do(req)

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	// log.Println("body", string(body))
	err = json.Unmarshal(body, &respond)
	if err != nil {
		return nil, err
	}
	if len(respond.Assets) == 0 {
		return nil, fmt.Errorf("failed to retrieve collection assets")
	}
	return respond.Assets, nil
}

// action byy:
/*
	https://core-api.prod.blur.io/v1/buy/0x25ddd361ab3649f331b578c0efd35d7242ffb90a
	Payload: {"tokenPrices":[{"tokenId":"2054","price":{"amount":"0.000000000000001","unit":"ETH"}}],"userAddress":"0x483d205d57f1fF227AF11232Be4acd34ab2C7914"}
	Payload: {"tokenPrices":[{"tokenId":"1647","price":{"amount":"0.0126","unit":"ETH"}},{"tokenId":"3305","price":{"amount":"0.015","unit":"ETH"}},{"tokenId":"8410","price":{"amount":"0.012","unit":"ETH"}}],"userAddress":"0x483d205d57f1fF227AF11232Be4acd34ab2C7914"}
	resonse: {"statusCode":400,"message":"Insufficient funds","error":"Bad Request"}
	{"buys":[],"cancelReasons":[{"tokenId":"2996","reason":"PriceTooLow"}]}

*/

func RetrieveBuyToken(OSEndpoint string, apiToken, blurDecodeKey, contractAddress string, payload BuyPayload) (BuyDataResponse, error) {

	contractAddress = "0x4abef147ccfcff7750c910ee7f62cf2448b8b52d"

	var respond struct {
		Data    string `json:"data"`
		Success bool   `json:"success"`
	}

	var BuyDataResponse BuyDataResponse

	fmt.Println("payload: ", payload)
	url := fmt.Sprintf("%v/v1/buy/%v", OSEndpoint, contractAddress)

	fmt.Println("url: ", url)

	j, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Marshal(payload) err: ", err)
		return BuyDataResponse, err
	}

	fmt.Println("string(payload): ", string(j))

	payloadJson := strings.NewReader(string(j))

	payloadJson = strings.NewReader(`{"tokenPrices":[{"tokenId":"2995","price":{"amount":"0","unit":"ETH"},"isSuspicious":false}],"userAddress":"0x96216849c49358b10257cb55b28ea603c874b05e"}`)

	req, _ := http.NewRequest("POST", url, payloadJson)

	req.Header.Add("Cookie", fmt.Sprintf("authToken=%s", apiToken))
	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	client.Transport = cloudflarebp.AddCloudFlareByPass(client.Transport)

	res, _ := client.Do(req)

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("ReadAll(Body) err: ", err)
		return BuyDataResponse, err
	}
	log.Println("body", string(body))
	err = json.Unmarshal(body, &respond)
	if err != nil {
		fmt.Println("Unmarshal(respond) err: ", err)
		return BuyDataResponse, err
	}
	fmt.Println("respond.Data err: ", respond.Data)
	if !respond.Success {
		fmt.Println("respond.Success err: ", respond.Success)
		return BuyDataResponse, fmt.Errorf("get data unsuccessful %v", respond.Success)
	}

	// decode:
	dataDecode, err := EncodeData(respond.Data, blurDecodeKey)
	if err != nil {
		fmt.Println("EncodeData err: ", err)
		return BuyDataResponse, err
	}

	err = json.Unmarshal([]byte(dataDecode), &BuyDataResponse)
	if err != nil {
		fmt.Println("Unmarshal(dataDecode) err: ", err)
		return BuyDataResponse, err
	}

	return BuyDataResponse, nil
}

func RetrieveAuthChallenge(OSEndpoint string, walletAddress string) (*LoginData, error) {
	var loginData LoginData

	walletJson := fmt.Sprintf(`{"walletAddress": "%s"}`, walletAddress)

	fmt.Println("walletJson: ", walletJson)

	payloadJson := strings.NewReader(walletJson)

	url := fmt.Sprintf("%v/auth/challenge", OSEndpoint)

	fmt.Println("url: ", url)

	req, _ := http.NewRequest("POST", url, payloadJson)

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	client.Transport = cloudflarebp.AddCloudFlareByPass(client.Transport)

	res, _ := client.Do(req)

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	log.Println("body", string(body))
	err = json.Unmarshal(body, &loginData)
	if err != nil {
		return nil, err
	}

	return &loginData, nil
}

func RetrieveAuthLogin(OSEndpoint, privKey string, loginData *LoginData) (string, error) {

	var respond struct {
		AccessToken string `json:"accessToken"`
	}
	signature, err := Sign(loginData.Message, privKey)
	if err != nil {
		fmt.Println("Sign(privKey) err: ", err)
		return "", err
	}

	loginData.Signature = signature
	fmt.Println("Signature: ", signature)

	j, err := json.Marshal(loginData)
	if err != nil {
		fmt.Println("Marshal(loginData) err: ", err)
		return "", err
	}

	fmt.Println("string(payload): ", string(j))

	payloadJson := strings.NewReader(string(j))

	// payloadJson = strings.NewReader(`
	// {"message":"Sign in to Blur\n\nChallenge: 7dfff84966390b3f492c225e69a171f444f2eaffb3051a4c28bfa2fec02d15b9","walletAddress":"0x4ba80ab11176a184ae23ee88d8c4af816862919f","expiresOn":"2023-01-14T17:33:38.261Z","hmac":"7dfff84966390b3f492c225e69a171f444f2eaffb3051a4c28bfa2fec02d15b9","signature":"0x3d3fe91f386cd37ba368cd01c4b326f52c4059a5f39135560cecc6397d022872224ec397f0297a5205aa0e6b57e6eb159db1bca81a8f411605838e6d9dcd639f1b"}`)
	// {"message":"Sign in to Blur\n\nChallenge: 59b403429e9034d561539c2245fec46db517d2ef3eb3605933f4efe8d3ea2ecc","walletAddress":"0x96216849c49358b10257cb55b28ea603c874b05e","expiresOn":"2023-01-14T17:36:12.669Z","hmac":"59b403429e9034d561539c2245fec46db517d2ef3eb3605933f4efe8d3ea2ecc","signature":"0x2e9e57ee584ce5584477c5bed032f88cba24ba86349fb57cf5f0ed7b639dcdc91d3476d2fee844cf7cecb33a485fee3922522797ea62b0487a33ac4b686639ae01"}
	// {"message":"Sign in to Blur\n\nChallenge: c1b6ede4e8a0b06903c945bf7ebb7446acac12da861a899e1f50b6059eabd8b6","walletAddress":"0x4ba80ab11176a184ae23ee88d8c4af816862919f","expiresOn":"2023-01-14T17:40:25.93Z","hmac":"c1b6ede4e8a0b06903c945bf7ebb7446acac12da861a899e1f50b6059eabd8b6","signature":"0xc7bbdb1402dba551aaedfbc564bae9b8be7acf47282e15611dfd5e341a72f0845ebad7a0569534217c44a189550c0ed733584e1ea24795b8822f18830ac374be01"}

	url := fmt.Sprintf("%v/auth/login", OSEndpoint)

	fmt.Println("url: ", url)

	req, _ := http.NewRequest("POST", url, payloadJson)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cookie", fmt.Sprintf("walletAddress=%s", loginData.WalletAddress))
	req.Header.Add("Cookie", fmt.Sprintf("rl_session=%s", loginData.WalletAddress))
	req.Header.Add("Cookie", fmt.Sprintf("__cf_bm=%s", "RudderEncrypt%3AU2FsdGVkX19hkvdTNoNDie8tVWzT4HjWS9v3nJ0OUgKZUaGUfjhQPHyp1MsHq4rbt7yJxz5e6eAGid4TdrX5RJh6vxzrhnbtmW3BbCYyT05kBHHhO1GeSDDXjU5Zqdq%2FjFlhcARRekonObXnuBW4Tg%3D%3D"))

	client := &http.Client{}
	client.Transport = cloudflarebp.AddCloudFlareByPass(client.Transport)

	res, _ := client.Do(req)

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	log.Println("body", string(body))
	err = json.Unmarshal(body, &respond)
	if err != nil {
		return "", err
	}

	return respond.AccessToken, nil
}

func RetrieveAuthAuth(OSEndpoint, walletAddress, privKey string) (string, error) {
	signMessage, err := RetrieveAuthChallenge(OSEndpoint, walletAddress)
	if err != nil {
		return "", err

	}
	time.Sleep(time.Second * 1)
	accessToken, err := RetrieveAuthLogin(OSEndpoint, privKey, signMessage)
	if err != nil {
		return "", err

	}
	return accessToken, nil
}

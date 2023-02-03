package pnft

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// /v1/user/nft_list
func RetrieveGetNftListDeBank(OSEndpoint, apiToken, address string) (string, error) {

	url := fmt.Sprintf("%v/v1/user/nft_list?id=%s&chain_id=eth&is_all=true", OSEndpoint, address)

	fmt.Println("url: ", url)

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("err0")
		return "", err
	}

	fmt.Println("apiToken: ", apiToken)

	req.Header.Add("AccessKey", fmt.Sprintf("%s", apiToken))
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		log.Println("err1")
		return "", err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("err2")
		return "", err
	}
	// log.Println("body", string(body))

	return string(body), nil
}

func RetrieveGetNftListQuickNode(OSEndpoint, address string) (string, error) {

	var respond struct {
		Jsonrpc string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Status  int    `json:"status"`
		Result  struct {
			Assets []Asset `json:"assets"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	payload := strings.NewReader(fmt.Sprintf(`{
		"id": 1,
		"method": "qn_fetchNFTs",
		"params": {
			"wallet": "%s",
			"perPage": 40,
			"page": 1
		}
	}`, address))

	client := &http.Client{}
	req, err := http.NewRequest("POST", OSEndpoint, payload)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	// fmt.Println(string(body))

	err = json.Unmarshal(body, &respond)
	if err != nil {
		log.Println("err3")
		return "", err
	}

	if respond.Error != nil {
		return "", errors.New(respond.Error.Message)
	}

	b, err := json.Marshal(respond.Result.Assets)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	// fmt.Println(string(b))

	return string(b), nil
}

func CheckNFTOwnerQuicknode(OSEndpoint, address string, assets map[string][]string) (map[string][]string, error) {
	notBelongAsset := make(map[string][]string)
	assetsToCheck := []string{}
	assetsToCheckStr := ""
	for coll, list := range assets {
		for _, v := range list {
			a := fmt.Sprintf("%v:%v", coll, v)
			assetsToCheck = append(assetsToCheck, a)
			assetsToCheckStr += fmt.Sprintf(`"%v",`, a)
		}
	}
	assetsToCheckStr = assetsToCheckStr[:len(assetsToCheckStr)-1]

	var respond struct {
		Jsonrpc string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Status  int    `json:"status"`
		Result  struct {
			Assets []string `json:"assets"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	payload := strings.NewReader(fmt.Sprintf(`{
		"id": 1,
		"method": "qn_verifyNFTsOwner",
		"params": [
			"%s",
			[%v]
		]
	}`, address, assetsToCheckStr))

	fmt.Println("payload", fmt.Sprintf(`{
		"id": 1,
		"method": "qn_verifyNFTsOwner",
		"params": [
			"%s",
			[%v]
		]
	}`, address, assetsToCheckStr))

	client := &http.Client{}
	req, err := http.NewRequest("POST", OSEndpoint, payload)

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

	err = json.Unmarshal(body, &respond)
	if err != nil {
		log.Println("err3")
		return nil, err
	}

	if respond.Error != nil {
		return nil, errors.New(respond.Error.Message)
	}

	for _, asset := range respond.Result.Assets {
		isBelong := false
		for _, v := range assetsToCheck {
			if asset == v {
				isBelong = true
				break
			}
		}
		if !isBelong {
			assetData := strings.Split(asset, ":")
			notBelongAsset[assetData[0]] = append(notBelongAsset[assetData[0]], assetData[1])
		}
	}

	return notBelongAsset, nil
}

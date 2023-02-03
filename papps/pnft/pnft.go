package pnft

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

///v1/user/nft_list
func RetrieveGetNftListDeBank(APIEndpoint, apiToken, address string) (string, error) {

	url := fmt.Sprintf("%v/v1/user/nft_list?id=%s&chain_id=eth&is_all=true", APIEndpoint, address)

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

func RetrieveGetNftListQuickNode(APIEndpoint, address string) (string, error) {

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
	req, err := http.NewRequest("POST", APIEndpoint, payload)

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

func RetrieveGetNftListFromMoralis(APIEndpoint, token, address string) (string, error) {

	var respond struct {
		Total    interface{} `json:"total"`
		Page     int         `json:"page"`
		PageSize int         `json:"page_size"`
		Cursor   interface{} `json:"cursor"`
		Result   []struct {
			TokenAddress      string    `json:"token_address"`
			TokenID           string    `json:"token_id"`
			OwnerOf           string    `json:"owner_of"`
			BlockNumber       string    `json:"block_number"`
			BlockNumberMinted string    `json:"block_number_minted"`
			TokenHash         string    `json:"token_hash"`
			Amount            string    `json:"amount"`
			ContractType      string    `json:"contract_type"`
			Name              string    `json:"name"`
			Symbol            string    `json:"symbol"`
			TokenURI          string    `json:"token_uri"`
			Metadata          string    `json:"metadata"`
			LastTokenURISync  time.Time `json:"last_token_uri_sync"`
			LastMetadataSync  time.Time `json:"last_metadata_sync"`
			MinterAddress     string    `json:"minter_address"`
		} `json:"result"`
		Status string `json:"status"`
	}

	url := fmt.Sprintf("%s/%s/nft?chain=eth&format=decimal", APIEndpoint, address)

	log.Println("url: ", url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")

	req.Header.Add("X-API-Key", token)

	fmt.Println("token", token)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		fmt.Println(res.StatusCode)
		return "", errors.New("can't not get data")
	}

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

	b, err := json.Marshal(respond.Result)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	// fmt.Println(string(b))

	return string(b), nil
}

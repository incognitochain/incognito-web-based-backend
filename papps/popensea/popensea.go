package popensea

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func RetrieveCollectionList(OSEndpoint string, apiKey string, limit int, offset int) ([]CollectionDetail, error) {
	var respond struct {
		Collections []CollectionDetail `json:"collections"`
	}
	url := fmt.Sprintf("%v/api/v1/collections?offset=%v&limit=%v", OSEndpoint, offset, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("err0")
		return nil, err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("X-API-KEY", apiKey)

	res, err := http.DefaultClient.Do(req)
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
	log.Println("body", string(body))
	err = json.Unmarshal(body, &respond)
	if err != nil {
		log.Println("err2")
		return nil, err
	}
	return respond.Collections, nil
}

func RetrieveCollectionDetail(OSEndpoint, apiKey string, slug string) (*CollectionDetail, error) {
	var respond struct {
		Collection CollectionDetail `json:"collection"`
	}
	url := fmt.Sprintf("%v/api/v1/collection/%v", OSEndpoint, slug)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("err0")
		return nil, err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("X-API-KEY", apiKey)

	res, err := http.DefaultClient.Do(req)
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
	log.Println("body", string(body))
	err = json.Unmarshal(body, &respond)
	if err != nil {
		log.Println("err2")
		return nil, err
	}
	if len(respond.Collection.PrimaryAssetContracts) == 0 {
		return nil, fmt.Errorf("failed to retrieve collection detail")
	}
	return &respond.Collection, nil
}

func RetrieveCollectionAssets(OSEndpoint string, apiKey string, collectionContract string, limit, offset int) ([]NFTDetail, error) {
	var respond struct {
		Assets []NFTDetail `json:"assets"`
	}
	url := fmt.Sprintf("%v/api/v1/assets?asset_contract_address=%v&order_direction=asc&include_orders=true&offset=%v&limit=%v", OSEndpoint, collectionContract, offset, limit)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("X-API-KEY", apiKey)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &respond)
	if err != nil {
		return nil, err
	}
	if len(respond.Assets) == 0 {
		return nil, fmt.Errorf("failed to retrieve collection assets")
	}
	return respond.Assets, nil
}

func RetrieveNFTDetail(OSEndpoint string, apiKey, collectionContract, tokenID string) (*NFTDetail, error) {
	var respond struct {
		Assets []NFTDetail `json:"assets"`
	}
	url := fmt.Sprintf("%v/api/v1/assets?asset_contract_address=%v&token_ids=%v&order_direction=asc&include_orders=true&offset=0&limit=1", OSEndpoint, collectionContract, tokenID)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("accept", "application/json")
	req.Header.Add("X-API-KEY", apiKey)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &respond)
	if err != nil {
		return nil, err
	}
	if len(respond.Assets) == 0 {
		return nil, fmt.Errorf("failed to retrieve nft detail")
	}
	return &respond.Assets[0], nil
}

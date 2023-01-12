package pblur

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

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

// // support mainnet only
// func RetrieveCollectionAssetByIDs(apiKey string, collectionContract string, ids []string) ([]NFTDetail, error) {
// 	var respond struct {
// 		Assets []NFTDetail `json:"assets"`
// 	}

// 	nftIds := ""
// 	for _, v := range ids {
// 		s := "&token_ids=" + v
// 		nftIds += s
// 	}

// 	url := fmt.Sprintf("https://api.Blur.io/api/v1/assets?order_direction=desc&asset_contract_address=%v&limit=30&include_orders=true%v", collectionContract, nftIds)
// 	// log.Println("RetrieveCollectionAssetByIDs url:", url)
// 	req, _ := http.NewRequest("GET", url, nil)

// 	req.Header.Add("accept", "application/json")
// 	req.Header.Add("X-API-KEY", apiKey)

// 	res, _ := http.DefaultClient.Do(req)

// 	defer res.Body.Close()
// 	body, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = json.Unmarshal(body, &respond)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(respond.Assets) == 0 {
// 		return nil, fmt.Errorf("failed to retrieve collection assets")
// 	}
// 	return respond.Assets, nil
// }

// // support mainnet only
// func RetrieveCollectionListing(apiKey string, collectionSlug string, next string) ([]NFTOrder, string, error) {
// 	var respond struct {
// 		Listings []NFTOrder `json:"listings"`
// 		Next     string     `json:"next"`
// 	}
// 	url := ""
// 	if next != "" {
// 		url = fmt.Sprintf("https://api.Blur.io/v2/listings/collection/%v/all?next=%v", collectionSlug, next)
// 	} else {
// 		url = fmt.Sprintf("https://api.Blur.io/v2/listings/collection/%v/all", collectionSlug)
// 	}
// 	// log.Println("RetrieveCollectionListing url:", url)

// 	req, _ := http.NewRequest("GET", url, nil)

// 	req.Header.Add("accept", "application/json")
// 	req.Header.Add("X-API-KEY", apiKey)

// 	res, _ := http.DefaultClient.Do(req)

// 	defer res.Body.Close()
// 	body, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		return nil, "", err
// 	}
// 	err = json.Unmarshal(body, &respond)
// 	if err != nil {
// 		return nil, "", err
// 	}
// 	if len(respond.Listings) == 0 {
// 		return nil, "", fmt.Errorf("failed to retrieve collection assets")
// 	}
// 	return respond.Listings, respond.Next, nil
// }

// func RetrieveNFTDetail(OSEndpoint string, apiKey, collectionContract, tokenID string) (*NFTDetail, error) {
// 	var respond struct {
// 		Assets []NFTDetail `json:"assets"`
// 	}
// 	url := fmt.Sprintf("%v/api/v1/assets?asset_contract_address=%v&token_ids=%v&order_direction=asc&include_orders=true&offset=0&limit=1", OSEndpoint, collectionContract, tokenID)

// 	req, _ := http.NewRequest("GET", url, nil)

// 	req.Header.Add("accept", "application/json")
// 	req.Header.Add("X-API-KEY", apiKey)

// 	res, _ := http.DefaultClient.Do(req)

// 	defer res.Body.Close()
// 	body, err := ioutil.ReadAll(res.Body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = json.Unmarshal(body, &respond)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if len(respond.Assets) == 0 {
// 		return nil, fmt.Errorf("failed to retrieve nft detail")
// 	}
// 	return &respond.Assets[0], nil
// }

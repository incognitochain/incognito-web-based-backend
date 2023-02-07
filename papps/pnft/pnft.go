package pnft

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/incognitochain/bridge-eth/bridge/pnft"
)

// /v1/user/nft_list
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
			Assets []QuicknodeNftDataResp `json:"assets"`
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

func RetrieveGetNftListFromMoralis(APIEndpoint, token, chain, address string) (string, error) {
	var respond struct {
		Total    interface{}          `json:"total"`
		Page     int                  `json:"page"`
		PageSize int                  `json:"page_size"`
		Cursor   interface{}          `json:"cursor"`
		Result   []MoralisNftDataResp `json:"result"`
		Status   string               `json:"status"`
	}

	url := fmt.Sprintf("%s/%s/nft?chain=%s&format=decimal&normalizeMetadata=true&limit=100", APIEndpoint, address, chain)

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

func RetrieveGetCollectionInfoFromOpensea(APIEndpoint, apiToken, contract string) (*OpenSeaCollectionResp, error) {
	var respond struct {
		// Collections struct {
		Collection OpenSeaCollectionResp `json:"collection"`
		// }
	}
	url := fmt.Sprintf("%v/api/v1/collections?format=json", APIEndpoint)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("err0")
		return nil, err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("X-API-KEY", apiToken)

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
	// log.Println("body", string(body))
	err = json.Unmarshal(body, &respond)
	if err != nil {
		log.Println("err3")
		return nil, err
	}
	return &respond.Collection, nil
}

func CheckNFTsOwnerMoralis(APIEndpoint, token, chain, address string, assetsToCheck map[string][]string) (map[string][]string, error) {
	notBelongAsset := make(map[string][]string)
	type AddressAssetsStruct struct {
		TokenAddress       string      `json:"token_address"`
		TokenID            string      `json:"token_id"`
		OwnerOf            string      `json:"owner_of"`
		BlockNumber        string      `json:"block_number"`
		BlockNumberMinted  string      `json:"block_number_minted"`
		TokenHash          string      `json:"token_hash"`
		Amount             string      `json:"amount"`
		ContractType       string      `json:"contract_type"`
		Name               string      `json:"name"`
		Symbol             string      `json:"symbol"`
		TokenURI           string      `json:"token_uri"`
		Metadata           string      `json:"metadata"`
		NormalizedMetadata interface{} `json:"normalized_metadata"`
		LastTokenURISync   time.Time   `json:"last_token_uri_sync"`
		LastMetadataSync   time.Time   `json:"last_metadata_sync"`
		MinterAddress      string      `json:"minter_address"`
	}
	addressAssetsStr, err := RetrieveGetNftListFromMoralis(APIEndpoint, token, chain, address)
	if err != nil {
		return nil, err
	}

	var addressAssets []AddressAssetsStruct

	err = json.Unmarshal([]byte(addressAssetsStr), &addressAssets)
	if err != nil {
		return nil, err
	}

	for collection, assets := range assetsToCheck {
		for _, asset := range assets {
			isBelong := false
			for _, v := range addressAssets {
				if asset == v.TokenID && strings.EqualFold(collection, v.TokenAddress) {
					isBelong = true
					break
				}
			}
			if !isBelong {
				notBelongAsset[collection] = append(notBelongAsset[collection], asset)
			}
		}
	}
	return notBelongAsset, nil
}

func VerifyOrderSignature(order *pnft.Input, orderHash string, ethClient *ethclient.Client, exchangeAddress string) error {
	orderHashBytes, err := hex.DecodeString(orderHash)
	if err != nil {
		return err
	}
	pnftInst, err := pnft.NewBlurExchange(common.HexToAddress(exchangeAddress), ethClient)
	if err != nil {
		return err
	}
	domainSeparator, _ := pnftInst.DOMAINSEPARATOR(nil)

	hashToSign := crypto.Keccak256Hash([]byte("\x19\x01"), domainSeparator[:], orderHashBytes[:])

	signature := []byte{}

	signature = append(signature, order.R[:]...)
	signature = append(signature, order.S[:]...)
	signature = append(signature, order.V-27)

	pubkey, _ := crypto.SigToPub(hashToSign[:], signature)
	pubkeyBytes := crypto.FromECDSAPub(pubkey)

	if !crypto.VerifySignature(pubkeyBytes, hashToSign[:], signature[:len(signature)-1]) {
		return errors.New("invalid signature")
	}
	return nil
}

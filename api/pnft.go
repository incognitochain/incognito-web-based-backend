package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	pnftContract "github.com/incognitochain/bridge-eth/bridge/pnft"
	"github.com/incognitochain/go-incognito-sdk-v2/coin"
	"github.com/incognitochain/go-incognito-sdk-v2/common/base58"
	"github.com/incognitochain/go-incognito-sdk-v2/metadata/bridge"
	metadataCommon "github.com/incognitochain/go-incognito-sdk-v2/metadata/common"
	"github.com/incognitochain/incognito-web-based-backend/common"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/papps"
	"github.com/incognitochain/incognito-web-based-backend/papps/pnft"
	"github.com/incognitochain/incognito-web-based-backend/submitproof"
	"go.mongodb.org/mongo-driver/mongo"
)

func APIPnftGetNftsFromAddress(c *gin.Context) {

	address, _ := c.GetQuery("address")

	log.Println("address: ", address)

	if len(address) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "address is empty"})
		return
	}

	address = strings.ToLower(address)

	// get them from db:
	result, _ := database.DBPNftGetListNftCacheTableByAddress(address)
	response := ""
	var err error

	if result != nil {
		response = result.Data
	}

	if len(response) == 0 {
		// response, err = pnft.RetrieveGetNftListDeBank(config.DebankAPI, config.DebankToken, address)
		// if err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})

		// }
		response, err = pnft.RetrieveGetNftListQuickNode(config.QuickNodeAPI, address)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})

		}
		// save db:
		database.DBPNftInsertListNftCacheTable(&common.ListNftCache{
			Address: address,
			Data:    response,
		})
	}

	// var returnData interface{}
	// err = json.Unmarshal([]byte(response), &returnData)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
	// 	return
	// }
	var jsonMap []map[string]interface{}
	json.Unmarshal([]byte(response), &jsonMap)

	c.JSON(200, gin.H{"Result": jsonMap})
}

func APIPNftGetCollections(c *gin.Context) {

	filter := c.Query("filters")

	fmt.Println("filter param: ", filter)

	var filterObj common.Filter
	if len(filter) > 0 {
		filter, err := url.QueryUnescape(filter)
		fmt.Println("filter: ", filter)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "filters invalid"})
			return
		}
		err = json.Unmarshal([]byte(filter), &filterObj)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "filters invalid: can not parse filter object"})
			return
		}
	}

	list, err := database.DBPNftGetCollectionList(&filterObj)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Limit": filterObj.Limit, "Page": filterObj.Page, "Offset": filterObj.Offset, "Query": filterObj.Query, "Total": len(list), "Result": list})
}

func APIPNftGetCollectionDetail(c *gin.Context) {

	filter := c.Query("filters")

	fmt.Println("filter param: ", filter)

	var filterObj common.Filter
	if len(filter) > 0 {
		filter, err := url.QueryUnescape(filter)
		fmt.Println("filter: ", filter)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "filters invalid"})
			return
		}
		err = json.Unmarshal([]byte(filter), &filterObj)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "filters invalid: can not parse filter object"})
			return
		}
	}

	slug := c.Param("slug")

	fmt.Println("slug: ", slug)

	collection, _ := database.DBPNftGetCollectionDetail(slug)

	if collection == nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "collection invalid"})
		return
	}

	list, err := database.DBPNftGetCollectionNFTs(collection.ContractAddress, &filterObj)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Result": gin.H{"tokens": list, "collection": collection}})
}

// est fee:
func APIPNftEstimateBuyFee(c *gin.Context) {
	contract := c.Query("contract")
	nftids := c.Query("nftids")
	burnToken := c.Query("burntoken")
	burnAmount := c.Query("burnamount")
	recipient := c.Query("recipient")

	_ = contract
	// currently only supports eth
	network := "eth"
	networkID := wcommon.GetNetworkID(network)
	networkFees, err := database.DBRetrieveFeeTable()
	if err != nil {
		fmt.Println("DBRetrieveFeeTable", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	burnTokenInfo, err := getTokenInfo(burnToken)
	if err != nil {
		fmt.Println("getTokenInfo", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("not supported token")})
		return
	}
	spTkList := getSupportedTokenList()
	burnAmountInc := uint64(0)
	amount := new(big.Int)
	_, errBool := amount.SetString(burnAmount, 10)
	if !errBool {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid amount")})
		return
	}
	if burnTokenInfo.CurrencyType == wcommon.UnifiedCurrencyType {
		for _, v := range burnTokenInfo.ListUnifiedToken {
			if networkID == v.NetworkID {
				amountUint64 := ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(burnAmount, v.Decimals, int64(v.PDecimals))
				burnAmountInc = amountUint64
				isEnoughVault, err := checkEnoughVault(burnToken, v.TokenID, amountUint64)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
					return
				}
				if !isEnoughVault {
					c.JSON(http.StatusBadRequest, gin.H{"Error": "not enough token in vault"})
					return
				}
			}
		}
	} else {
		amountUint64 := ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(burnAmount, burnTokenInfo.Decimals, int64(burnTokenInfo.PDecimals))
		burnAmountInc = amountUint64
	}

	log.Println("nftids: ", nftids)

	feeAmount, err := estimateOpenSeaFee(burnAmountInc, burnTokenInfo, network, networkFees, spTkList)
	if err != nil {
		fmt.Println("estimateNftFee", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	pappList, err := database.DBRetrievePAppsByNetwork(network)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("no supported papps found").Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	proxyContract, exist := pappList.AppContracts["pnft"] // todo: Lam -> need to add config
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "blur contract not found"})
		return
	}
	log.Println("proxyContract: ", proxyContract)
	log.Println("recipient: ", recipient)

	// get list asset of the collection:
	listNFTOrder, err := database.DBPNftGetNFTSellOrder(contract, strings.Split(nftids, ","))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	var sellInputs []pnftContract.Execution

	for _, order := range listNFTOrder {
		var input pnftContract.Input
		err := json.Unmarshal([]byte(order.OrderInput), &input)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		sellInputs = append(sellInputs, pnftContract.Execution{Sell: input})
	}

	callData, err := papps.BuildpNFTCalldata(sellInputs, proxyContract, recipient)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	receiveToken := strings.ToLower("6722ec501bE09fb221bCC8a46F9660868d0a6c63")
	if config.NetworkID == "testnet" {
		receiveToken = strings.ToLower("4cB607c24Ac252A0cE4b2e987eC4413dA0F1e3Ae")
	}

	result := struct {
		Fee          *OpenSeaFee
		Calldata     string
		CallContract string
		ReceiveToken string
	}{
		Fee:          feeAmount,
		Calldata:     callData,
		CallContract: contract[2:],
		ReceiveToken: receiveToken,
	}
	c.JSON(200, gin.H{"Result": result})
}

func APIPNftGetInfoForListing(c *gin.Context) {
	collection := c.Query("collection")
	_ = collection
	//TODO: implement
	//get fee information for a collection
	//get MatchingPolicy address

	result := struct {
		Fee            map[string]uint
		MatchingPolicy string
	}{}
	c.JSON(200, gin.H{"Result": result})
}

// TODO: implement
func APIPNftListing(c *gin.Context) {
	var req PnftListingReq
	userAgent := c.Request.UserAgent()
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	_ = userAgent

	newListing := []common.PNftSellOrder{}

	for _, item := range req.Items {
		listingItem := common.PNftSellOrder{}

		itemData, err := json.Marshal(item)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}

		listingItem.OrderInput = string(itemData)
	}

	listingErr := make(map[string]map[string]string)
	for _, item := range newListing {
		err := database.DBPNftInsertSellOrder(&item)
		if err != nil {
			if len(listingErr[item.ContractAddress]) == 0 {
				listingErr[item.ContractAddress] = map[string]string{}
			}
			listingErr[item.ContractAddress][item.TokenID] = err.Error()
		}
	}

}

// TODO: implement
func APIPNftDelisting(c *gin.Context) {
	var req PnftDelistingReq
	userAgent := c.Request.UserAgent()
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	_ = userAgent
}

// TODO: implement
func APIPNftSubmitDelist(c *gin.Context) {
	var req SubmitSwapTxRequest
	userAgent := c.Request.UserAgent()
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	_ = userAgent
}

// TODO: implement
func APIPNftSubmitBuy(c *gin.Context) {
	var req SubmitSwapTxRequest
	userAgent := c.Request.UserAgent()
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	if req.FeeRefundOTA != "" && req.FeeRefundAddress != "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "FeeRefundOTA & FeeRefundAddress can't be used as the same time"})
		return
	}

	var md *bridge.BurnForCallRequest
	var mdRaw metadataCommon.Metadata
	var isPRVTx bool
	var isUnifiedToken bool
	var outCoins []coin.Coin
	var txHash string
	var rawTxBytes []byte

	if req.FeeRefundOTA == "" && req.FeeRefundAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "FeeRefundOTA/FeeRefundAddress need to be provided one of these values"})
		return
	}

	var ok bool
	rawTxBytes, _, err = base58.Base58Check{}.Decode(req.TxRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid txhash")})
		return
	}

	mdRaw, isPRVTx, outCoins, txHash, err = extractDataFromRawTx(rawTxBytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	spTkList := getSupportedTokenList()

	statusResult := checkPappTxSwapStatus(txHash, spTkList)
	if len(statusResult) > 0 {
		if er, ok := statusResult["error"]; ok {
			if er != "not found" {
				c.JSON(200, gin.H{"Result": statusResult})
				return
			}
		} else {
			c.JSON(200, gin.H{"Result": statusResult})
			return
		}
	}

	if mdRaw == nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid tx metadata type")})
		return
	}
	md, ok = mdRaw.(*bridge.BurnForCallRequest)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid tx metadata type")})
		return
	}

	burnTokenInfo, err := getTokenInfo(md.BurnTokenID.String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid tx metadata type")})
		return
	}
	if burnTokenInfo.CurrencyType == wcommon.UnifiedCurrencyType {
		isUnifiedToken = true
	}

	valid, networkList, feeToken, feeAmount, pfeeAmount, feeDiff, swapInfo, err := checkValidTxSwap(md, outCoins, spTkList, wcommon.ExternalTxTypePNFT_Buy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid tx err:" + err.Error()})
		return
	}
	// valid = true

	burntAmount, _ := md.TotalBurningAmount()
	if valid {
		status, err := submitproof.SubmitPappTx(txHash, []byte(req.TxRaw), isPRVTx, feeToken, feeAmount, pfeeAmount, md.BurnTokenID.String(), burntAmount, swapInfo, isUnifiedToken, networkList, req.FeeRefundOTA, req.FeeRefundAddress, userAgent, wcommon.ExternalTxTypePNFT_Buy)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"Result": map[string]interface{}{"inc_request_tx_status": status}, "feeDiff": feeDiff})
		return
	}
}

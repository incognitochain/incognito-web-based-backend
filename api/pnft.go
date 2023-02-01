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
	"github.com/incognitochain/incognito-web-based-backend/common"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/papps/pnft"
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
	mycontract, exist := pappList.AppContracts["pnft"] // todo: Lam -> need to add config
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "blur contract not found"})
		return
	}
	log.Println("mycontract: ", mycontract)
	log.Println("recipient: ", recipient)

	// get list asset of the collection:
	listDataAsset, _ := database.DBPNftGetNFTDetailByIDs(contract, strings.Split(nftids, ","))

	if len(listDataAsset) == 0 {
		fmt.Println("DBNftGetNFTDetailByIDs empty")
		c.JSON(http.StatusBadRequest, gin.H{"Error": "list nft empty"})
		return
	}

	// get detail only:
	var listNftDetail []pnft.NFTDetail
	for _, asset := range listDataAsset {
		listNftDetail = append(listNftDetail, asset.Detail)
	}

	// todo: get call data here...

	callData := ""

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

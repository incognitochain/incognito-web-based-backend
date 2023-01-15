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
	"github.com/incognitochain/incognito-web-based-backend/papps"
	"github.com/incognitochain/incognito-web-based-backend/papps/pblur"
)

func APIPBlurAuthChallenge(c *gin.Context) {
	var req WalletAddress
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	result, err := pblur.RetrieveAuthChallenge(config.BlurAPI, req.WalletAddress)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"Result": result})
}

//{"message":"Sign in to Blur\n\nChallenge: 59d811eb3c86b542ba95524552eb461b3713d219a6a813dd3cf9b8bfc55456bd","walletAddress":"0x8693afc2005f63b5eb9786bc3d488dae0e9cac23","expiresOn":"2023-01-13T06:10:18.659Z","hmac":"59d811eb3c86b542ba95524552eb461b3713d219a6a813dd3cf9b8bfc55456bd"}

func APIPBlurAuthLogin(c *gin.Context) {

}

func APIPBlurGetCollections(c *gin.Context) {

	// page := 1
	// var err error
	// if len(c.Query("page")) > 0 {
	// 	page, err = strconv.Atoi(c.Query("page"))
	// 	if err != nil {
	// 		c.JSON(http.StatusBadRequest, gin.H{"Error": "page invalid"})
	// 		return
	// 	}
	// }

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

	list, err := database.DBBlurGetCollectionList(&filterObj)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Limit": filterObj.Limit, "Page": filterObj.Page, "Offset": filterObj.Offset, "Query": filterObj.Query, "Total": len(list), "Result": list})
}

func APIPBlurGetCollectionDetail(c *gin.Context) {

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

	collection, _ := database.DBBlurGetCollectionDetail(slug)

	if collection == nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "collection invalid"})
		return
	}

	list, err := database.DBBlurGetCollectionNFTs(collection.ContractAddress, &filterObj)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Result": gin.H{"tokens": list, "collection": collection}})
}

// est fee:
func APIBlurEstimateBuyFee(c *gin.Context) {
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
		fmt.Println("estimateBlurFee", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	// pappList, err := database.DBRetrievePAppsByNetwork(network)
	// if err != nil {
	// 	if err == mongo.ErrNoDocuments {
	// 		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("no supported papps found").Error()})
	// 		return
	// 	}
	// 	c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
	// 	return
	// }
	// mycontract, exist := pappList.AppContracts["blur"]
	// if !exist {
	// 	c.JSON(http.StatusBadRequest, gin.H{"Error": "blur contract not found"})
	// 	return
	// }

	// get list asset of the collection:
	listDataAsset, _ := database.DBBlurGetNFTDetailByIDs(contract, strings.Split(nftids, ","))

	if len(listDataAsset) == 0 {
		fmt.Println("DBBlurGetNFTDetailByIDs empty")
		c.JSON(http.StatusBadRequest, gin.H{"Error": "list nft empty"})
		return
	}

	// get detail only:
	var listNftDetail []pblur.NFTDetail
	for _, asset := range listDataAsset {
		listNftDetail = append(listNftDetail, asset.Detail)
	}
	// get data from blur api:
	var payload pblur.BuyPayload
	payload.UserAddress = recipient
	for _, nft := range listNftDetail {
		payload.TokenPrices = append(payload.TokenPrices, pblur.TokenPrice{
			TokenID: nft.TokenID,
			Price: pblur.Price{
				Amount: nft.Price.Amount,
				Unit:   nft.Price.Unit,
			},
		})
	}

	buyDataResponse, err := pblur.RetrieveBuyToken(config.BlurAPI, config.BlurToken, config.BlurDecodeKey, contract, payload)
	if err != nil {
		fmt.Println("RetrieveBuyToken", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	fmt.Println("buyDataResponse: ", buyDataResponse)

	callData, err := papps.BuildBlurCalldata(buyDataResponse, recipient)
	if err != nil {
		fmt.Println("estimateBlurFee", err.Error())
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

// gen access token:
func APIBlurGenAccessToken(c *gin.Context) {
	accessToken, err := pblur.RetrieveAuthAuth(config.BlurAPI, config.BlurWalletAddress, config.BlurPrivateKey)
	if err != nil {
		fmt.Println("RetrieveAuthAuth err: ", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"Result": accessToken})
}

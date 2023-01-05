package api

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/incognitochain/go-incognito-sdk-v2/coin"
	"github.com/incognitochain/go-incognito-sdk-v2/common"
	"github.com/incognitochain/go-incognito-sdk-v2/common/base58"
	"github.com/incognitochain/go-incognito-sdk-v2/metadata/bridge"
	metadataCommon "github.com/incognitochain/go-incognito-sdk-v2/metadata/common"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/papps"
	"github.com/incognitochain/incognito-web-based-backend/papps/popensea"
	"github.com/incognitochain/incognito-web-based-backend/submitproof"
	"go.mongodb.org/mongo-driver/mongo"
)

func APIEstimateBuyFee(c *gin.Context) {
	contract := c.Query("contract")
	nftid := c.Query("nftid")
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

	nftDetail, err := popensea.RetrieveNFTDetail(config.OpenSeaAPI, config.OpenSeaAPIKey, contract, nftid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	if len(nftDetail.SeaportSellOrders) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "this NFT is not available for sell"})
		return
	}

	feeAmount, err := estimateOpenSeaFee(burnAmountInc, burnTokenInfo, network, networkFees, spTkList)
	if err != nil {
		fmt.Println("estimateOpenSeaFee", err.Error())
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
	contract, exist := pappList.AppContracts["opensea"]
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "opensea contract not found"})
		return
	}
	callData, err := papps.BuildOpenSeaCalldata(nftDetail, recipient)
	if err != nil {
		fmt.Println("estimateOpenSeaFee", err.Error())
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
func APISubmitBuyTx(c *gin.Context) {
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

	valid, networkList, feeToken, feeAmount, pfeeAmount, feeDiff, swapInfo, err := checkValidTxSwap(md, outCoins, spTkList, wcommon.ExternalTxTypeOpensea)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid tx err:" + err.Error()})
		return
	}
	// valid = true

	burntAmount, _ := md.TotalBurningAmount()
	if valid {
		status, err := submitproof.SubmitPappTx(txHash, []byte(req.TxRaw), isPRVTx, feeToken, feeAmount, pfeeAmount, md.BurnTokenID.String(), burntAmount, swapInfo, isUnifiedToken, networkList, req.FeeRefundOTA, req.FeeRefundAddress, userAgent, wcommon.ExternalTxTypeOpensea)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"Result": map[string]interface{}{"inc_request_tx_status": status}, "feeDiff": feeDiff})
		return
	}
}

func APIGetCollections(c *gin.Context) {
	defaultList, err := database.DBGetDefaultCollectionList()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	result := []popensea.CollectionDetail{}
	dupCollection := make(map[string]struct{})
	for _, coll := range defaultList {
		data, err := database.DBGetCollectionsInfo(coll.Address)
		if err != nil {
			log.Printf("DBGetCollectionsInfo err %v \n", coll.Slug)
			continue
		}
		if _, ok := dupCollection[strings.ToLower(coll.Address)]; ok {
			continue
		}
		dupCollection[strings.ToLower(coll.Address)] = struct{}{}
		result = append(result, data.Detail)
	}
	// collections, err := popensea.RetrieveCollectionList(config.OpenSeaAPI, config.OpenSeaAPIKey, 20, 0)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
	// 	return
	// }
	c.JSON(http.StatusOK, gin.H{"Result": result})
}
func APINFTDetail(c *gin.Context) {
	contract := c.Query("contract")
	nftid := c.Query("nftid")
	var nftDetail *popensea.NFTDetail
	var err error
	if config.NetworkID == "mainnet" {
		nftDetailDB, err := database.DBGetNFTDetail(contract, nftid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		nftDetail = &nftDetailDB.Detail
	} else {
		nftDetail, err = popensea.RetrieveNFTDetail(config.OpenSeaAPI, config.OpenSeaAPIKey, contract, nftid)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"Result": nftDetail})
}

func APICollectionAssets(c *gin.Context) {
	contract := c.Query("contract")
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	limit = 9999
	var result []popensea.NFTDetail
	if config.NetworkID == "mainnet" {
		assetList, err := database.DBGetCollectionNFTs(contract, int64(limit), int64(offset))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		for _, asset := range assetList {
			if len(asset.Detail.SeaportSellOrders) > 0 {
				if len(asset.Detail.SeaportSellOrders[0].ProtocolData.Parameters.Consideration) > 0 {
					if asset.Detail.SeaportSellOrders[0].ProtocolData.Parameters.Consideration[0].ItemType == 0 {
						result = append(result, asset.Detail)
					}
				}
			}
		}
		c.JSON(http.StatusOK, gin.H{"Result": result})
		return
	}
	assetList, err := popensea.RetrieveCollectionAssets(config.OpenSeaAPI, config.OpenSeaAPIKey, contract, limit, offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	//force only return buy-able assets
	for _, asset := range assetList {
		if len(asset.SeaportSellOrders) > 0 {
			result = append(result, asset)
		}
	}
	c.JSON(http.StatusOK, gin.H{"Result": result})
}
func APICollectionDetail(c *gin.Context) {
	contract := c.Query("contract")
	data, err := database.DBGetCollectionsInfo(contract)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"Result": data})
}

func estimateOpenSeaFee(amount uint64, burnTokenInfo *wcommon.TokenInfo, network string, networkFees *wcommon.ExternalNetworksFeeData, spTkList []PappSupportedTokenData) (*OpenSeaFee, error) {

	networkID := wcommon.GetNetworkID(network)
	isSupportedOutNetwork := false
	if burnTokenInfo.CurrencyType == wcommon.UnifiedCurrencyType {
		for _, childToken := range burnTokenInfo.ListUnifiedToken {
			childNetID, err := wcommon.GetNetworkIDFromCurrencyType(childToken.CurrencyType)
			if err != nil {
				return nil, err
			}
			if childNetID == networkID {
				isSupportedOutNetwork = true
				break
			}
		}

	} else {
		netID, err := getNetworkIDFromCurrencyType(burnTokenInfo.CurrencyType)
		if err != nil {
			return nil, err
		}
		if netID == networkID {
			isSupportedOutNetwork = true
		}
	}
	if !isSupportedOutNetwork {
		return nil, errors.New("unsupported network")
	}

	feeTokenWhiteList, err := retrieveFeeTokenWhiteList()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	isFeeWhitelist := false
	if _, ok := feeTokenWhiteList[burnTokenInfo.TokenID]; ok {
		isFeeWhitelist = true
	}

	if _, ok := networkFees.GasPrice[network]; !ok {
		return nil, errors.New("network gasPrice not found")
	}
	gasPrice := networkFees.GasPrice[network]

	nativeCurrentType := wcommon.GetNativeNetworkCurrencyType(network)
	nativeToken, err := getNativeTokenData(spTkList, nativeCurrentType)
	if err != nil {
		return nil, err
	}

	isUnifiedNativeToken := false

	if burnTokenInfo.CurrencyType == nativeCurrentType {
		isUnifiedNativeToken = true
	}
	if burnTokenInfo.CurrencyType == wcommon.UnifiedCurrencyType {
		for _, v := range burnTokenInfo.ListUnifiedToken {
			if v.CurrencyType == nativeCurrentType {
				isUnifiedNativeToken = true
				break
			}
		}
	}
	gasFee := (wcommon.EVMGasLimitOpensea * gasPrice)
	amountInBig0 := new(big.Float).SetUint64(amount)

	additionalTokenInFee := amountInBig0.Mul(amountInBig0, new(big.Float).SetFloat64(0.003))
	additionalTokenInFee = additionalTokenInFee.Mul(additionalTokenInFee, new(big.Float).SetFloat64(math.Pow10(-burnTokenInfo.PDecimals)))
	fees := getFee(isFeeWhitelist, isUnifiedNativeToken, nativeToken, new(big.Float).SetInt64(1), gasFee, burnTokenInfo.TokenID, burnTokenInfo, &PappSupportedTokenData{
		CurrencyType: burnTokenInfo.CurrencyType,
	}, new(big.Float).SetInt64(1), additionalTokenInFee, true)
	if len(fees) == 0 {
		return nil, errors.New("can't get fee")
	}
	// burntAmount := uint64(0)
	// protocolFee := uint64(0)
	// if burnTokenInfo.CurrencyType == wcommon.UnifiedCurrencyType {
	// 	var tokenID string

	// 	for _, childToken := range burnTokenInfo.ListUnifiedToken {
	// 		childNetID, err := wcommon.GetNetworkIDFromCurrencyType(childToken.CurrencyType)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		if childNetID == networkID {
	// 			tokenID = childToken.TokenID
	// 			break
	// 		}
	// 	}
	// 	burntAmount, protocolFee, err = getUnifiedUnshieldFee(tokenID, burnTokenInfo.TokenID, amount)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// } else {
	// 	burntAmount = amount
	// }

	feeAddress := ""
	feeAddressShardID := byte(0)
	if incFeeKeySet != nil {
		feeAddress, err = incFeeKeySet.GetPaymentAddress()
		if err != nil {
			return nil, err
		}
		_, feeAddressShardID = common.GetShardIDsFromPublicKey(incFeeKeySet.KeySet.PaymentAddress.Pk)
	}
	result := OpenSeaFee{
		FeeAddress:        feeAddress,
		FeeAddressShardID: int(feeAddressShardID),
		TokenID:           fees[0].TokenID,
		Amount:            fees[0].Amount,
		PrivacyFee:        fees[0].PrivacyFee,
		// ProtocolFee:       protocolFee,
		FeeInUSD: fees[0].FeeInUSD,
	}

	return &result, nil
}

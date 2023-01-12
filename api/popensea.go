package api

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
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

	nftDetail, err := popensea.RetrieveNFTDetail(config.OpenSeaAPI, config.OpenSeaAPIKey, contract, nftid, false)
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

	valid, networkList, feeToken, feeAmount, pfeeAmount, feeDiff, swapInfo, err := checkValidTxSwap(md, outCoins, spTkList, wcommon.ExternalTxTypeOpenseaBuy)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid tx err:" + err.Error()})
		return
	}
	// valid = true

	burntAmount, _ := md.TotalBurningAmount()
	if valid {
		status, err := submitproof.SubmitPappTx(txHash, []byte(req.TxRaw), isPRVTx, feeToken, feeAmount, pfeeAmount, md.BurnTokenID.String(), burntAmount, swapInfo, isUnifiedToken, networkList, req.FeeRefundOTA, req.FeeRefundAddress, userAgent, wcommon.ExternalTxTypeOpenseaBuy)
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
		nftDetail, err = popensea.RetrieveNFTDetail(config.OpenSeaAPI, config.OpenSeaAPIKey, contract, nftid, false)
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
	var result []popensea.NFTDetail
	if config.NetworkID == "mainnet" {
		limit = 9999
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
	if limit > 200 {
		limit = 200
	}

	assetList, err := popensea.RetrieveCollectionAssets(config.OpenSeaAPI, config.OpenSeaAPIKey, contract, limit, offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	//force only return buy-able assets
	for _, asset := range assetList {
		// if len(asset.SeaportSellOrders) > 0 {
		result = append(result, asset)
		// }
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
	gasFee := (wcommon.EVMGasLimitOpenseaOffer * gasPrice)
	amountInBig0 := new(big.Float).SetUint64(amount)

	additionalTokenInFee := amountInBig0.Mul(amountInBig0, new(big.Float).SetFloat64(0.003))
	additionalTokenInFee = additionalTokenInFee.Mul(additionalTokenInFee, new(big.Float).SetFloat64(math.Pow10(-burnTokenInfo.PDecimals)))
	fees := getFee(isFeeWhitelist, isUnifiedNativeToken, nativeToken, new(big.Float).SetInt64(1), gasFee, burnTokenInfo.TokenID, burnTokenInfo, &PappSupportedTokenData{
		CurrencyType: burnTokenInfo.CurrencyType,
	}, new(big.Float).SetInt64(1), additionalTokenInFee, true)
	if len(fees) == 0 {
		return nil, errors.New("can't get fee")
	}

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

func APIOpenSeaListOffer(c *gin.Context) {
	// walletAddress := c.Query("wallet")
}

func APIOpenSeaOfferStatus(c *gin.Context) {
	var req SubmitTxListRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	result := make(map[string]interface{})

	// spTkList := getSupportedTokenList()
	var wg sync.WaitGroup
	var lock sync.Mutex
	for _, txHash := range req.TxList {
		wg.Add(1)
		go func(txh string) {
			statusResult, err := checkOpenseaOfferStatus(txh)
			if err != nil {
				statusResult = make(map[string]interface{})
				statusResult["error"] = err.Error()
				result[txh] = statusResult
				wg.Done()
				return
			}
			lock.Lock()
			if len(statusResult) == 0 {
				statusResult["error"] = "tx not found"
				result[txh] = statusResult
			} else {
				result[txh] = statusResult
			}
			lock.Unlock()
			wg.Done()
		}(txHash)
	}
	wg.Wait()
	c.JSON(200, gin.H{"Result": result})
}

func APIOpenSeaSubmitOffer(c *gin.Context) {
	var req SubmitOpenseaOfferTxRequest
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
	if req.Offer == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid offer"})
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

	valid, networkList, feeToken, feeAmount, pfeeAmount, feeDiff, swapInfo, err := checkValidTxSwap(md, outCoins, spTkList, wcommon.ExternalTxTypeOpenseaOffer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid tx err:" + err.Error()})
		return
	}
	// valid = true

	burntAmount, _ := md.TotalBurningAmount()
	if valid {
		swapInfo.AdditionalData = req.Offer
		status, err := submitproof.SubmitPappTx(txHash, []byte(req.TxRaw), isPRVTx, feeToken, feeAmount, pfeeAmount, md.BurnTokenID.String(), burntAmount, swapInfo, isUnifiedToken, networkList, req.FeeRefundOTA, req.FeeRefundAddress, userAgent, wcommon.ExternalTxTypeOpenseaOffer)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"Result": map[string]interface{}{"inc_request_tx_status": status}, "feeDiff": feeDiff})
		return
	}
}

func APIOpenSeaCancelOffer(c *gin.Context) {
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

	valid, networkList, feeToken, feeAmount, pfeeAmount, feeDiff, swapInfo, err := checkValidTxSwap(md, outCoins, spTkList, wcommon.ExternalTxTypeOpenseaOfferCancel)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid tx err:" + err.Error()})
		return
	}
	// valid = true

	burntAmount, _ := md.TotalBurningAmount()
	if valid {
		status, err := submitproof.SubmitPappTx(txHash, []byte(req.TxRaw), isPRVTx, feeToken, feeAmount, pfeeAmount, md.BurnTokenID.String(), burntAmount, swapInfo, isUnifiedToken, networkList, req.FeeRefundOTA, req.FeeRefundAddress, userAgent, wcommon.ExternalTxTypeOpenseaOfferCancel)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"Result": map[string]interface{}{"inc_request_tx_status": status}, "feeDiff": feeDiff})
		return
	}
}

func APIOpenSeaGenOffer(c *gin.Context) {
	var req OpenSeaGenOfferRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	var nftDetail *popensea.NFTDetail
	if config.NetworkID == "mainnet" {
		nftDetailDB, err := database.DBGetNFTDetail(req.CollectionAddress, req.NFTID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		nftDetail = &nftDetailDB.Detail
	} else {
		nftDetail, err = popensea.RetrieveNFTDetail(config.OpenSeaAPI, config.OpenSeaAPIKey, req.CollectionAddress, req.NFTID, true)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
	}

	papps, err := database.DBRetrievePAppsByNetwork("eth")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	offerAdapterAddr, exist := papps.AppContracts["opensea-offer-proxy"]
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "opensea offer adapter not found"})
		return
	}
	seaportAddress, exist := papps.AppContracts["seaport"]
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "opensea seaport not found"})
		return
	}

	offer, err := popensea.GenOfferOrder("0x0000000000000000000000000000000000000000", offerAdapterAddr, offerAdapterAddr, req.Amount, req.StartTime, req.EndTime, *nftDetail)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	offerJson, err := json.Marshal(offer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	networknInfo, err := database.DBGetBridgeNetworkInfo("eth")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	offerSignData := ""
	for _, endpoint := range networknInfo.Endpoints {
		evmClient, err := ethclient.Dial(endpoint)
		if err != nil {
			log.Println("ethclient.Dial err:", err)
			continue
		}

		seaport, err := popensea.NewIopensea(ethcommon.HexToAddress(seaportAddress), evmClient)
		if err != nil {
			log.Println("popensea.NewIopensea err:", err)
			continue
		}
		orderHash, err := seaport.GetOrderHash(nil, *offer)
		if err != nil {
			log.Println("seaport.GetOrderHash err:", err)
			continue
		}

		openseaOfferAddr := ethcommon.HexToAddress(offerAdapterAddr)
		offerAdapter, err := popensea.NewOpenseaOffer(openseaOfferAddr, evmClient)
		if err != nil {
			log.Println("popensea.NewOpenseaOffer err:", err)
			continue
		}
		domainSeparator, _ := offerAdapter.DomainSeparator(nil)
		signData, _ := offerAdapter.ToTypedDataHash(nil, domainSeparator, orderHash)

		offerSignData = hex.EncodeToString(signData[:])
		break
	}
	if offerSignData == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "can't get offer hash"})
		return
	}

	result := struct {
		Offer         string `json:"offer"`
		OfferSignData string `json:"offer_sign_data"`
	}{
		OfferSignData: offerSignData,
		Offer:         hex.EncodeToString(offerJson),
	}
	c.JSON(200, gin.H{"Result": result})
}

// TODO: opensea
func APIEstimateOfferFee(c *gin.Context) {
	var req OpenSeaOfferFeeEstimateRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	offer := popensea.OrderComponents{}
	offerBytes, err := hex.DecodeString(req.Offer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "offerBytes, err := hex.DecodeString(req.Offer) " + err.Error()})
		return
	}

	err = json.Unmarshal(offerBytes, &offer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "err = json.Unmarshal(offerBytes, &offer) " + err.Error()})
		return
	}

	// contract := offer.Consideration[0].Token.Hex()
	// nftid := offer.Consideration[0].IdentifierOrCriteria.String()
	burnToken := req.BurnToken

	burnAmountBig := new(big.Int).SetInt64(0)

	burnAmountBig.Add(offer.Offer[0].StartAmount, burnAmountBig)
	for idx, cn := range offer.Consideration {
		if idx == 0 { // 1st consideration is nft
			continue
		}
		burnAmountBig.Add(cn.StartAmount, burnAmountBig)
	}
	burnAmountBig.Div(burnAmountBig, new(big.Int).SetInt64(int64(math.Pow10(9))))
	burnAmount := burnAmountBig.String()

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

	// nftDetail, err := popensea.RetrieveNFTDetail(config.OpenSeaAPI, config.OpenSeaAPIKey, contract, nftid)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
	// 	return
	// }

	// if len(nftDetail.SeaportSellOrders) == 0 {
	// 	c.JSON(http.StatusBadRequest, gin.H{"Error": "this NFT is not available for sell"})
	// 	return
	// }

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
	callProxy, exist := pappList.AppContracts["opensea-offer-forward"]
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "opensea-proxy contract not found"})
		return
	}

	openseaOffer, exist := pappList.AppContracts["opensea-offer-proxy"]
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "opensea-proxy contract not found"})
		return
	}
	proxyOfferAddress := ethcommon.HexToAddress(openseaOffer)
	recipient := ethcommon.HexToAddress(req.Recipient)
	openseaProxyAbi, _ := abi.JSON(strings.NewReader(popensea.OpenseaMetaData.ABI))
	openseaOfferAbi, _ := abi.JSON(strings.NewReader(popensea.OpenseaOfferMetaData.ABI))
	ota := req.Ota
	signBytes, err := hex.DecodeString(req.Signature)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if signBytes[64] <= 1 {
		signBytes[64] += 27
	}
	conduit := ethcommon.BytesToAddress(offer.ConduitKey[:])

	tempData, err := openseaOfferAbi.Pack("offer", offer, ota, signBytes, conduit, recipient)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "openseaOfferAbi.Pack" + err.Error()})
		return
	}
	//TODO: opensea add recipient to calldata
	callData, err := openseaProxyAbi.Pack("forward", proxyOfferAddress, tempData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "openseaProxyAbi.Pack" + err.Error()})
		return
	}

	callDataHex := hex.EncodeToString(callData)
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
		Calldata:     callDataHex,
		CallContract: callProxy[2:],
		ReceiveToken: receiveToken,
	}
	c.JSON(200, gin.H{"Result": result})
}

func APIEstimateCancelFee(c *gin.Context) {
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

	nftDetail, err := popensea.RetrieveNFTDetail(config.OpenSeaAPI, config.OpenSeaAPIKey, contract, nftid, false)
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
	contract, exist := pappList.AppContracts["opensea-offer-proxy"]
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "opensea-offer contract not found"})
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

func checkOpenseaOfferStatus(txhash string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	offerData, err := database.DBGetOpenseaOfferByOfferTx(txhash)
	if err != nil {
		return nil, err
	}

	result["timeout_at"] = offerData.TimeoutAt
	result["nftid"] = offerData.NFTID
	result["collection"] = offerData.NFTCollection
	result["receiver"] = offerData.Receiver
	result["status"] = offerData.Status

	result["offer_inc_tx"] = txhash
	offerTx, err := database.DBGetPappTxData(txhash)
	if err != nil {
		return nil, err
	}

	result["offer_inc_tx_status"] = offerTx.Status
	if offerTx.Status == wcommon.StatusAccepted {
		offerExtTx, err := database.DBGetExternalTxByIncTx(txhash, "eth")
		if err != nil {
			return nil, err
		}
		result["offer_external_tx"] = offerExtTx.Txhash
		result["offer_external_tx_status"] = offerExtTx.Status

		if offerExtTx.WillRedeposit {
			result["reshield_tx_status"] = wcommon.StatusSubmitting
			if offerExtTx.RedepositSubmitted {
				reshieldData, err := database.DBGetShieldTxByExternalTx(offerExtTx.Txhash, 1)
				if err != nil {
					return nil, err
				}
				result["reshield_tx"] = reshieldData.IncTx
				result["reshield_tx_status"] = reshieldData.Status
			}
		} else {
			if offerData.CancelTxInc != "" {
				result["cancel_inc_tx"] = offerData.CancelTxInc
				result["cancel_external_tx"] = ""
				// result["cancel_seaport_tx"] = offerData.CancelOpenseaTx
				result["cancel_inc_tx_status"] = offerData.CancelTxInc
				result["cancel_external_tx_status"] = ""
				result["reshield_tx"] = ""
				result["reshield_tx_status"] = ""
				// result["cancel_seaport_tx_status"] = offerData.CancelOpenseaTx
			}
			if offerData.Status == popensea.OfferStatusFilled || offerData.Status == popensea.OfferStatusClaiming || offerData.Status == popensea.OfferStatusClaimed {
				result["claim_tx"] = offerData.ClaimNFTTx
			}
		}

	}
	return result, nil
}

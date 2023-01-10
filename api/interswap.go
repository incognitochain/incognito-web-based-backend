package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/incognitochain/go-incognito-sdk-v2/coin"
	"github.com/incognitochain/go-incognito-sdk-v2/common"
	"github.com/incognitochain/go-incognito-sdk-v2/common/base58"
	"github.com/incognitochain/go-incognito-sdk-v2/metadata/bridge"
	metadataCommon "github.com/incognitochain/go-incognito-sdk-v2/metadata/common"
	"github.com/incognitochain/go-incognito-sdk-v2/metadata/pdexv3"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gin-gonic/gin"
	beCommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/interswap"
)

const EmptyExternalAddress = "0000000000000000000000000000000000000000"

type AddOnSwapInfo struct {
	ToToken        string
	Path           []string
	MinAcceptedAmt uint64
	CallContract   string
	AppName        string
}

type SubmitInterSwapTxRequest struct {
	TxRaw  string
	TxHash string

	FromToken           string
	ToToken             string
	MidToken            string
	FinalMinExpectedAmt uint64
	Slippage            string

	PAppName    string
	PAppNetwork string
	ShardID     string

	OTARefundFee    string // user's OTA to receive refunded swap papp fee (sender is BE, receiver is user)
	OTARefund       string // user's OTA to receive fund from InterswapBE only in case PappPdexType and the first tx is reverted (sender is InterswapBE, receiver is user)
	OTAFromToken    string // user's OTA to receive refunded swap amount (sell token || mid token)
	OTAToToken      string // user's OTA to receive buy token
	WithdrawAddress string // only withdraw when PathType = pDexToPapp
}

// validate sanity request
func ValidateSubmitInterSwapTxRequest(req SubmitInterSwapTxRequest, network string) (bool, string, error) {
	if interswap.IsMidTokens(req.ToToken) || interswap.IsMidTokens(req.FromToken) {
		return false, "", errors.New("FromToken and ToToken should be diff from midToken")
	}

	if !interswap.IsMidTokens(req.MidToken) {
		return false, "", errors.New("MidToken is invalid")
	}

	if req.FinalMinExpectedAmt == 0 {
		return false, "", errors.New("FinalMinExpectedAmt must be greater than 0")
	}

	if req.PAppName == "" || req.PAppNetwork == "" {
		return false, "", errors.New("PAppName, PAppNetwork must not empty")
	}

	// validate papp info
	pappEndpint := beCommon.MainnetPappsEndpointData
	if network == "testnet" {
		pappEndpint = beCommon.TestnetPappsEndpointData
	}

	pAppContract := ""
	for _, data := range pappEndpint {
		if data.Network == req.PAppNetwork {
			pAppContract = interswap.Remove0xPrefix(data.AppContracts[req.PAppName])
		}
	}
	if pAppContract == "" {
		return false, "", errors.New("PAppContract not found with PAppName, PAppNetwork")
	}

	// validate OTA keys
	if req.OTAFromToken == "" || req.OTAToToken == "" || req.OTARefund == "" || req.OTARefundFee == "" {
		return false, "", errors.New("OTA keys must not empty")
	}

	otas := []string{req.OTAFromToken, req.OTAToToken, req.OTARefund, req.OTARefundFee}

	isValidOTA := interswap.IsUniqueSlices(otas)
	if !isValidOTA {
		return false, "", errors.New("OTA keys must not be duplicated")
	}
	isValid, err := IsValidOTAs(otas)
	if err != nil || !isValid {
		return false, "", errors.New("OTA keys is invalid")
	}

	return true, pAppContract, nil
}

func IsValidOTAs(otas []string) (bool, error) {
	for _, ota := range otas {
		coin := &coin.OTAReceiver{}
		err := coin.FromString(ota)
		if err != nil {
			return false, err
		}
		if !coin.IsValid() {
			return false, nil
		}
	}
	return true, nil
}

func APISubmitInterSwapTx(c *gin.Context) {
	log.Println("Processing APISubmitInterSwapTx")
	var req SubmitInterSwapTxRequest
	userAgent := c.Request.UserAgent()
	log.Println("Processing APISubmitInterSwapTx 1")
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("Can not parse submit request %v", err.Error())})
		return
	}
	log.Println("Processing APISubmitInterSwapTx 2 - Req ", req)

	// validate sanity req
	isValidSanity, pAppContract, err := ValidateSubmitInterSwapTxRequest(req, config.NetworkID)
	if err != nil || !isValidSanity {

		c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("invalid submit interswap req %v", err)})
		return
	}
	log.Println("Processing APISubmitInterSwapTx 3")

	var mdRaw metadataCommon.Metadata
	var isPRVTx bool
	var txHash string
	var rawTxBytes []byte

	// decode raw tx 1
	rawTxBytes, _, err = base58.Base58Check{}.Decode(req.TxRaw)
	if err != nil {
		log.Printf("APISubmitInterSwapTx Decode raw tx error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("decode raw tx error: %v", err)})
		return
	}
	log.Println("Processing APISubmitInterSwapTx 4")

	mdRaw, isPRVTx, _, txHash, err = extractDataFromRawTx(rawTxBytes)
	if err != nil {
		log.Printf("APISubmitInterSwapTx Extract raw tx error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if txHash != req.TxHash {
		log.Printf("APISubmitInterSwapTx TxID Mismatched: Expected %v, Got %v\n", txHash, req.TxHash)
		c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("TxID Mismatched: Expected %v, Got %v\n", txHash, req.TxHash)})
		return
	}

	mdType := mdRaw.GetType()
	if mdType != metadataCommon.BurnForCallRequestMeta && mdType != metadataCommon.Pdexv3TradeRequestMeta {
		c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("invalid metadata %v", mdType)})
		return
	}

	shardIDByte, err := strconv.Atoi(req.ShardID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("invalid ShardID %v", err)})
		return
	}

	log.Println("Processing APISubmitInterSwapTx 5")

	switch mdType {
	case metadataCommon.BurnForCallRequestMeta:
		{
			md, ok := mdRaw.(*bridge.BurnForCallRequest)
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid tx metadata type"})
				return
			}
			if len(md.Data) != 1 {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid metadata burn for call: md.Data"})
				return
			}

			// validate sellToken
			if md.BurnTokenID.String() != req.FromToken {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid metadata BurnForCallRequest BurnTokenID must be FromToken"})
				return
			}

			// validate buyToken must be midToken
			// Note: It is contractId of token
			childMidToken, err := interswap.GetChildTokenUnified(req.MidToken, int(md.Data[0].ExternalNetworkID), config)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("cannot get child token of midToken %v - ExternalTokenID %v", req.MidToken, md.Data[0].ExternalNetworkID)})
				return
			}

			tokenInfo, err := getTokenInfo(childMidToken)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "cannot get token info of md.Data[0].IncTokenID"})
				return
			}
			log.Printf("APISubmitInterSwapTx tokenInfo of ReceiveToken: %v\n", tokenInfo)
			if strings.ToLower(md.Data[0].ReceiveToken) != strings.ToLower(interswap.Remove0xPrefix(tokenInfo.ContractID)) {
				c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("invalid metadata BurnForCallRequest Data ReceiveToken must be %v", tokenInfo.ContractID)})
				return
			}

			// the first papp tx must be reshield
			if md.Data[0].WithdrawAddress != EmptyExternalAddress {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid metadata BurnForCallRequest Data WithdrawAddress"})
				return
			}
			log.Println("Processing APISubmitInterSwapTx 6")

			fromAmount, err := md.TotalBurningAmount()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "cannot get total burning amount"})
				return
			}

			// NOTE: can't verify RedepositReceiver belong to ISIncPrivKey or not
			// only check UTXOs in response tx

			// store DB to InterSwap
			// to avoid missing db record since the storing db error occurs in while the tx swap was submitted
			status := interswap.FirstPending

			interswapInfo := beCommon.InterSwapTxData{
				TxID:  txHash,
				TxRaw: req.TxRaw,

				OTARefundFee:    req.OTARefundFee,
				OTARefund:       req.OTARefund,
				OTAFromToken:    req.OTAFromToken,
				OTAToToken:      req.OTAToToken,
				WithdrawAddress: req.WithdrawAddress,

				FromAmount:          fromAmount,
				FromToken:           md.BurnTokenID.String(),
				ToToken:             req.ToToken,
				MidToken:            req.MidToken,
				PathType:            interswap.PAppToPdex,
				FinalMinExpectedAmt: req.FinalMinExpectedAmt,
				Slippage:            req.Slippage,
				ShardID:             byte(shardIDByte),

				PAppName:     req.PAppName,
				PAppNetwork:  req.PAppNetwork,
				PAppContract: pAppContract,

				Status:    status,
				StatusStr: interswap.StatusStr[status],
				UserAgent: userAgent,
				Error:     "",
			}

			err = database.DBInsertInterswapTxData(interswapInfo)
			if err != nil {
				writeErr, ok := err.(mongo.WriteException)
				if !ok {
					c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("DBInsertInterswapTxData err %v", err)})
					return
				}
				if !writeErr.HasErrorCode(11000) {
					c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("DBInsertInterswapTxData err %v", err)})
					return
				}
			}

			log.Println("Processing APISubmitInterSwapTx 7")
			// call api submitswaptx to broadcast papp swap tx to BE
			_, err = interswap.CallSubmitPappSwapTx(req.TxRaw, txHash, req.OTARefundFee, config, "")
			if err != nil {
				status := interswap.SubmitFailed
				err = database.DBUpdateInterswapTxStatus(txHash, status, interswap.StatusStr[status], err.Error())
				if err != nil {
					log.Printf("InterswapID %v DBUpdateInterswapTxStatus err: %v", txHash, err)
				}
				c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("submit papp tx failed %v", err.Error())})
				return
			}
			err = interswap.SendSlackSwapInfo(txHash, userAgent, "submiting", fromAmount, req.FromToken, req.FinalMinExpectedAmt, req.ToToken, 0, "", config)
			if err != nil {
				log.Printf("InterswapID %v SendSlackSwapInfo err %v\n", txHash, err)
			}

			c.JSON(200, gin.H{"Result": "success"})
			return
		}
	case metadataCommon.Pdexv3TradeRequestMeta:
		{
			md, ok := mdRaw.(*pdexv3.TradeRequest)
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid tx metadata type"})
				return
			}

			buyTokenID := common.Hash{}
			for tokenID, _ := range md.Receiver {
				if tokenID != md.TokenToSell && tokenID != common.PRVCoinID {
					buyTokenID = tokenID
					break
				}
			}

			// validate sellToken
			if md.TokenToSell.String() != req.FromToken {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid metadata TradeRequest TokenToSell must be FromToken"})
				return
			}

			// validate buyToken must be midToken
			if buyTokenID.String() != req.MidToken {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid metadata TradeRequest buyTokenID must be MidToken"})
				return
			}

			log.Println("Processing APISubmitInterSwapTx 8")
			withdrawAddress := ""
			if req.WithdrawAddress == "" {
				withdrawAddress = EmptyExternalAddress
			} else {
				withdrawAddress = interswap.Remove0xPrefix(req.WithdrawAddress)
			}
			// store DB to InterSwap before broadcast tx
			status := interswap.FirstPending
			interswapInfo := beCommon.InterSwapTxData{
				TxID:  txHash,
				TxRaw: req.TxRaw,

				OTARefundFee:    req.OTARefundFee,
				OTARefund:       req.OTARefund,
				OTAFromToken:    req.OTAFromToken,
				OTAToToken:      req.OTAToToken,
				WithdrawAddress: withdrawAddress,

				FromAmount:          md.SellAmount,
				FromToken:           req.FromToken,
				ToToken:             req.ToToken,
				MidToken:            req.MidToken,
				PathType:            interswap.PdexToPApp,
				FinalMinExpectedAmt: req.FinalMinExpectedAmt,
				Slippage:            req.Slippage,
				ShardID:             byte(shardIDByte),

				PAppName:     req.PAppName,
				PAppNetwork:  req.PAppNetwork,
				PAppContract: pAppContract,

				Status:    interswap.FirstPending,
				StatusStr: interswap.StatusStr[status],
				UserAgent: userAgent,
				Error:     "",
			}
			// if error occurs when saving DB in while tx swap was submitted
			err = database.DBInsertInterswapTxData(interswapInfo)
			if err != nil {
				writeErr, ok := err.(mongo.WriteException)
				if !ok {
					log.Println("DBSaveInterSwapTxData err", err)
					c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("DBSaveInterSwapTxData err %v", err)})
					return
				}
				if !writeErr.HasErrorCode(11000) {
					log.Println("DBSaveInterSwapTxData err", err)
					c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("DBSaveInterSwapTxData err %v", err)})
					return
				}
			}
			log.Println("Processing APISubmitInterSwapTx 9")

			// send raw tx
			if isPRVTx {
				err = incClient.SendRawTx([]byte(req.TxRaw))
			} else {
				err = incClient.SendRawTokenTx([]byte(req.TxRaw))
			}
			log.Printf("InterswapID %v isPRVTx %v Send raw tx error %v", txHash, isPRVTx, err)
			if err != nil {
				status := interswap.SubmitFailed
				err = database.DBUpdateInterswapTxStatus(txHash, status, interswap.StatusStr[status], err.Error())
				if err != nil {
					log.Printf("InterswapID %v DBUpdateInterswapTxStatus err: %v", txHash, err)
				}
				c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("broadcast pdex raw tx failed %v", err.Error())})
				return
			}
			log.Println("Processing APISubmitInterSwapTx 10")

			err = interswap.SendSlackSwapInfo(txHash, userAgent, "submiting", md.SellAmount, req.FromToken, req.FinalMinExpectedAmt, req.ToToken, 0, "", config)
			if err != nil {
				log.Printf("InterswapID %v SendSlackSwapInfo err %v\n", txHash, err)
			}

			c.JSON(200, gin.H{"Result": "success"})
			return
		}
	default:
		{
			c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("invalid metadata %v", mdType)})
			return
		}
	}
}

func APIGetInterswapTxStatus(c *gin.Context) {
	var req SubmitTxListRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	result := make(map[string]interface{})

	var wg sync.WaitGroup
	var lock sync.Mutex
	for _, txHash := range req.TxList {
		wg.Add(1)
		go func(txh string) {
			statusResult, err := getInterswapTxStatus(txh)
			lock.Lock()
			if err != nil {
				result[txh] = map[string]string{"error": "tx not found"}
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

type InterswapStatus struct {
	Status       string
	PdexTxID     string
	PappTxID     string
	FromAmount   uint64
	FromToken    string
	ToAmount     uint64
	ToToken      string
	ResponseTxID string
	RefundAmount uint64
	RefundToken  string
	RefundTxID   string
	TxIDOutchain string
}

func getInterswapTxStatus(txID string) (*InterswapStatus, error) {
	data, err := database.DBRetrieveInterswapTxByTxID(txID)
	if err != nil {
		log.Printf("DBRetrieveInterswapTxByTxID %v %v", txID, err)
		return nil, err
	}
	pdexTxID := ""
	pappTxID := ""
	if data.PathType == interswap.PdexToPApp {
		pdexTxID = data.TxID
		pappTxID = data.AddOnTxID
	} else {
		pappTxID = data.TxID
		pdexTxID = data.AddOnTxID
	}

	toAmt := uint64(0)
	refundAmt := uint64(0)
	tokenRefund := ""
	if data.AmountResponse > 0 {
		if data.StatusStr == interswap.InterswapRefundedStr || data.StatusStr == interswap.InterswapRefundingStr {
			refundAmt = data.AmountResponse
			tokenRefund = data.TokenResponse
		} else {
			toAmt = data.AmountResponse
		}
	}

	res := &InterswapStatus{
		Status:       data.StatusStr,
		PdexTxID:     pdexTxID,
		PappTxID:     pappTxID,
		FromAmount:   data.FromAmount,
		FromToken:    data.FromToken,
		ToAmount:     toAmt,
		ToToken:      data.ToToken,
		ResponseTxID: data.TxIDResponse,
		RefundAmount: refundAmt,
		RefundToken:  tokenRefund,
		RefundTxID:   data.TxIDRefund,
		TxIDOutchain: data.TxIDOutchain,
	}

	return res, nil

}

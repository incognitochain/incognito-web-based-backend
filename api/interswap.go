package api

import (
	"errors"
	"fmt"
	"log"
	"net/http"

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
	"github.com/incognitochain/incognito-web-based-backend/submitproof"
)

const PdexSwapType = 1
const PappSwapType = 2

const EmptyExternalAddress = "0x0000000000000000000000000000000000000000"

type AddOnSwapInfo struct {
	ToToken        string
	Path           []string
	MinAcceptedAmt uint64
	CallContract   string
	AppName        string
}

type SubmitInterSwapTxRequest struct {
	//
	// pdex => pApp: broadcast chain pdex, worker create tx pApp, call func submitproof.SubmitPappTx; refund pApp fee
	// pApp => pDEx: call API submitswaptx (privacyFee = privacyFee + pDEXTRadingFee), worker create tx pDEX
	//
	// pApp: call API submitswaptx
	// store DB InterSwap
	// worker: get status by load from DB (Lam)

	TxRaw  string
	TxHash string

	FromToken           string
	ToToken             string
	MidToken            string
	FinalMinExpectedAmt uint64
	Slippage            string

	PAppName     string
	PAppNetwork  string
	PAppContract string

	OTARefundFee    string // user's OTA to receive refunded swap papp fee (sender is BE, receiver is user)
	OTARefund       string // user's OTA to receive fund from InterswapBE only in case PappPdexType and the first tx is reverted (sender is InterswapBE, receiver is user)
	OTAFromToken    string // user's OTA to receive refunded swap amount (sell token || mid token)
	OTAToToken      string // user's OTA to receive buy token
	WithdrawAddress string // only withdraw when PathType = pDexToPapp
}

// TODO: 0xkraken
// validate sanity request
func ValidateSubmitInterSwapTxRequest(req SubmitInterSwapTxRequest) (bool, error) {
	if interswap.IsMidTokens(req.ToToken) || interswap.IsMidTokens(req.FromToken) {
		return false, errors.New("FromToken and ToToken should be diff from midToken")
	}

	if !interswap.IsMidTokens(req.MidToken) {
		return false, errors.New("MidToken is invalid")
	}

	return true, nil

	// if addOnSwapInfo.FromToken != expectedFromToken {
	// 	return false, fmt.Errorf("From token in addon swap info is invalid: expected %v - got %v", expectedFromToken, addOnSwapInfo.FromToken)
	// }

	// if addOnSwapInfo.ToToken == "" {
	// 	return false, errors.New("The add on swap info ToToken is required")
	// }

	// if addOnSwapInfo.MinExpectedAmt == 0 {
	// 	return false, errors.New("The add on swap info MinExpectedAmt must be greater than 0")
	// }

	// switch expectSwapType {
	// case PdexSwapType:
	// 	{
	// 		if addOnSwapInfo.AppName != "pdex" {
	// 			return false, errors.New("The add on swap info must be pdex")
	// 		}
	// 		return true, nil
	// 	}
	// case PappSwapType:
	// 	{
	// 		if addOnSwapInfo.AppName == "pdex" {
	// 			return false, errors.New("The add on swap info must be papp")
	// 		}
	// 		if addOnSwapInfo.CallContract == "" {
	// 			return false, errors.New("The add on swap info CallContract is required")
	// 		}
	// 		return true, nil
	// 	}
	// default:
	// 	{
	// 		return false, errors.New("Invalid swap type")
	// 	}
	// }
}

// TODO: 0xkraken
// IsValidOTA returns true if ota belongs to privKey
func IsValidOTA(ota coin.OTAReceiver, privKey string) (bool, error) {
	// NOTE: only can check when receive the response tx

	return true, nil
}

func APISubmitInterSwapTx(c *gin.Context) {
	var req SubmitInterSwapTxRequest
	userAgent := c.Request.UserAgent()
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	// var md *bridge.BurnForCallRequest
	var mdRaw metadataCommon.Metadata
	var isPRVTx bool
	// var isUnifiedToken bool
	// var outCoins []coin.Coin
	var txHash string
	var rawTxBytes []byte

	// validate sanity req
	isValidSanity, err := ValidateSubmitInterSwapTxRequest(req)
	if err != nil || !isValidSanity {
		c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Errorf("invalid submit interswap req %v", err)})
		return
	}

	// decode raw tx 1
	rawTxBytes, _, err = base58.Base58Check{}.Decode(req.TxRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid raw tx")})
		return
	}

	mdRaw, isPRVTx, _, txHash, err = extractDataFromRawTx(rawTxBytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	if txHash != req.TxHash {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid txhash - mismatched")})
		return
	}

	mdType := mdRaw.GetType()
	if mdType != metadataCommon.BurnForCallRequestMeta && mdType != metadataCommon.Pdexv3TradeRequestMeta {
		c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Errorf("invalid metadata %v", mdType)})
		return
	}

	switch mdType {
	case metadataCommon.BurnForCallRequestMeta:
		{
			md, ok := mdRaw.(*bridge.BurnForCallRequest)
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid tx metadata type")})
				return
			}
			if len(md.Data) != 1 {
				c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid metadata burn for call: md.Data")})
				return
			}

			// the first papp tx must be reshield
			if md.Data[0].WithdrawAddress != EmptyExternalAddress {
				c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid metadata BurnForCallRequest Data")})
				return
			}

			// redeposit address must belong to ISIncPrivKey
			isValid, err := IsValidOTA(md.Data[0].RedepositReceiver, config.ISIncPrivKey)
			if err != nil || !isValid {
				c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid metadata BurnForCallRequest RedepositReceiver")})
				return
			}

			// TODO: validate tx info with req (fromToken, midToken,...)

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

				FromToken:           md.BurnTokenID.String(),
				ToToken:             req.ToToken,
				MidToken:            req.MidToken,
				PathType:            interswap.PAppToPdex,
				FinalMinExpectedAmt: req.FinalMinExpectedAmt,
				Slippage:            req.Slippage,

				PAppName:     req.PAppName,
				PAppNetwork:  req.PAppNetwork,
				PAppContract: req.PAppContract,

				Status:    status,
				StatusStr: interswap.StatusStr[status],
				UserAgent: userAgent,
				Error:     "",
			}

			_, err = database.DBSaveInterSwapTxData(interswapInfo)
			if err != nil {
				writeErr, ok := err.(mongo.WriteException)
				if !ok {
					c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Errorf("DBSaveInterSwapTxData err %v", err)})
					return
				}
				if !writeErr.HasErrorCode(11000) {
					c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Errorf("DBSaveInterSwapTxData err %v", err)})
					return
				}
			}

			// call api submitswaptx to broadcast papp swap tx to BE
			_, err = interswap.CallSubmitPappSwapTx(req.TxRaw, txHash, req.OTARefundFee, config)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Errorf("submit papp tx failed %v", err)})
				//
				// TODO: update DB status failed
				return
			}
			msgParam := interswap.InterswapSubmitTxTask{
				TxID:       txHash,
				TxRawBytes: rawTxBytes,

				OTARefundFee: req.OTARefundFee,
				OTARefund:    req.OTARefund,
				OTAFromToken: req.OTAFromToken,
				OTAToToken:   req.OTAToToken,

				FromToken:           md.BurnTokenID.String(),
				ToToken:             req.ToToken,
				MidToken:            req.MidToken,
				PathType:            interswap.PAppToPdex,
				FinalMinExpectedAmt: req.FinalMinExpectedAmt,

				PAppName:     req.PAppName,
				PAppNetwork:  req.PAppNetwork,
				PAppContract: req.PAppContract,

				Status:    status,
				StatusStr: interswap.StatusStr[status],
				UserAgent: userAgent,
				Error:     "",
			}

			// call assigner to publish msg to Interswap Worker
			_, err = submitproof.PublishMsgInterswapTx(msgParam)
			if err != nil {
				// TODO: Update DB status failed
				c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
				return
			}

			c.JSON(200, gin.H{"Result": map[string]interface{}{"inc_request_tx_status": interswap.FirstPending}})
			return

		}
	case metadataCommon.Pdexv3TradeRequestMeta:
		{
			md, ok := mdRaw.(*pdexv3.TradeRequest)
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid tx metadata type")})
				return
			}

			// receiver address of buyToken must belong to ISIncPrivKey (if swap success)
			// receiver address of sellToken must belong to user (if swap fail, don't need to refund the swap amount)
			buyTokenID := common.Hash{}
			for tokenID, _ := range md.Receiver {
				// must belong users
				if tokenID == md.TokenToSell || tokenID == common.PRVCoinID {
					// TODO: 0xkraken
				} else {
					buyTokenID = tokenID
				}
			}
			isValid, err := IsValidOTA(md.Receiver[buyTokenID], config.ISIncPrivKey)
			if err != nil || !isValid {
				c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid metadata BurnForCallRequest RedepositReceiver")})
				return
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
				WithdrawAddress: req.WithdrawAddress,

				FromToken:           req.FromToken,
				ToToken:             req.ToToken,
				MidToken:            req.MidToken,
				PathType:            interswap.PdexToPApp,
				FinalMinExpectedAmt: req.FinalMinExpectedAmt,
				Slippage:            req.Slippage,

				PAppName:     req.PAppName,
				PAppNetwork:  req.PAppNetwork,
				PAppContract: req.PAppContract,

				Status:    interswap.FirstPending,
				StatusStr: interswap.StatusStr[status],
				UserAgent: userAgent,
				Error:     "",
			}
			// if error occurs when saving DB in while tx swap was submitted
			_, err = database.DBSaveInterSwapTxData(interswapInfo)
			if err != nil {
				writeErr, ok := err.(mongo.WriteException)
				if !ok {
					log.Println("DBSaveInterSwapTxData err", err)
					return
				}
				if !writeErr.HasErrorCode(11000) {
					log.Println("DBSaveInterSwapTxData err", err)
					return
				}
			}

			// send raw tx
			if isPRVTx {
				err = incClient.SendRawTx(rawTxBytes)
			} else {
				err = incClient.SendRawTokenTx(rawTxBytes)
			}
			if err != nil {
				// TODO: Update DB status failed
				c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Errorf("broadcast pdex raw tx failed %v", err)})
				return
			}

			msgParam := interswap.InterswapSubmitTxTask{
				TxID:       txHash,
				TxRawBytes: rawTxBytes,

				OTARefundFee: req.OTARefundFee,
				OTARefund:    req.OTARefund,
				OTAFromToken: req.OTAFromToken,
				OTAToToken:   req.OTAToToken,

				FromToken:           req.FromToken,
				ToToken:             req.ToToken,
				MidToken:            req.MidToken,
				PathType:            interswap.PdexToPApp,
				FinalMinExpectedAmt: req.FinalMinExpectedAmt,

				PAppName:     req.PAppName,
				PAppNetwork:  req.PAppNetwork,
				PAppContract: req.PAppContract,

				Status:    interswap.FirstPending,
				StatusStr: interswap.StatusStr[status],
				UserAgent: userAgent,
				Error:     "",
			}

			// call assigner to publish msg to Interswap Worker
			_, err = submitproof.PublishMsgInterswapTx(msgParam)
			if err != nil {
				// TODO: Update DB status failed
				c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
				return
			}

			c.JSON(200, gin.H{"Result": map[string]interface{}{"inc_request_tx_status": interswap.FirstPending}})
			return
		}
	default:
		{
			c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Errorf("invalid metadata %v", mdType)})
			return
		}
	}

	// spTkList := getSupportedTokenList()

	// statusResult := checkPappTxSwapStatus(txHash, spTkList)
	// if len(statusResult) > 0 {
	// 	if er, ok := statusResult["error"]; ok {
	// 		if er != "not found" {
	// 			c.JSON(200, gin.H{"Result": statusResult})
	// 			return
	// 		}
	// 	} else {
	// 		c.JSON(200, gin.H{"Result": statusResult})
	// 		return
	// 	}
	// }

	// if mdRaw == nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid tx metadata type")})
	// 	return
	// }

	// md, ok := mdRaw.(*bridge.BurnForCallRequest)
	// if !ok {
	// 	c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid tx metadata type")})
	// 	return
	// }

	// burnTokenInfo, err := getTokenInfo(md.BurnTokenID.String())
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid tx metadata type")})
	// 	return
	// }
	// if burnTokenInfo.CurrencyType == wcommon.UnifiedCurrencyType {
	// 	isUnifiedToken = true
	// }

	// valid, networkList, feeToken, feeAmount, pfeeAmount, feeDiff, swapInfo, err := checkValidTxSwap(md, outCoins, spTkList)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid tx err:" + err.Error()})
	// 	return
	// }
	// // valid = true

	// burntAmount, _ := md.TotalBurningAmount()
	// if valid {
	// 	status, err := submitproof.SubmitPappTx(txHash, []byte(req.TxRaw), isPRVTx, feeToken, feeAmount, pfeeAmount, md.BurnTokenID.String(), burntAmount, swapInfo, isUnifiedToken, networkList, req.FeeRefundOTA, req.FeeRefundAddress, userAgent)
	// 	if err != nil {
	// 		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
	// 		return
	// 	}
	// 	c.JSON(200, gin.H{"Result": map[string]interface{}{"inc_request_tx_status": status}, "feeDiff": feeDiff})
	// 	return
	// }

}

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

// validate sanity request
func ValidateSubmitInterSwapTxRequest(req SubmitInterSwapTxRequest, network string) (bool, error) {
	if interswap.IsMidTokens(req.ToToken) || interswap.IsMidTokens(req.FromToken) {
		return false, errors.New("FromToken and ToToken should be diff from midToken")
	}

	if !interswap.IsMidTokens(req.MidToken) {
		return false, errors.New("MidToken is invalid")
	}

	if req.FinalMinExpectedAmt == 0 {
		return false, errors.New("FinalMinExpectedAmt must be greater than 0")
	}

	if req.PAppName == "" || req.PAppNetwork == "" || req.PAppContract == "" {
		return false, errors.New("PAppName, PAppNetwork, PAppContract must not empty")
	}

	// validate papp info
	pappEndpint := beCommon.MainnetPappsEndpointData
	if network == "testnet" {
		pappEndpint = beCommon.TestnetPappsEndpointData
	}

	isValid := false
	for _, data := range pappEndpint {
		if data.Network == req.PAppNetwork {
			if data.AppContracts[req.PAppName] == req.PAppContract {
				isValid = true
			}
		}
	}
	if isValid {
		return false, errors.New("PAppName, PAppNetwork, PAppContract not matched")
	}

	// validate OTA keys
	if req.OTAFromToken == "" || req.OTAToToken == "" || req.OTARefund == "" || req.OTARefundFee == "" {
		return false, errors.New("OTA keys must not empty")
	}

	isValidOTA := interswap.IsUniqueSlices([]string{req.OTAFromToken, req.OTAToToken, req.OTARefund, req.OTARefundFee})
	if !isValidOTA {
		return false, errors.New("OTA keys must not be duplicated")
	}

	return true, nil
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

	// validate sanity req
	isValidSanity, err := ValidateSubmitInterSwapTxRequest(req, config.NetworkID)
	if err != nil || !isValidSanity {
		c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Errorf("invalid submit interswap req %v", err)})
		return
	}

	var mdRaw metadataCommon.Metadata
	var isPRVTx bool
	var txHash string
	var rawTxBytes []byte

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

	// TODO: 0xkraken check database is exist or not

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

			// validate sellToken
			if md.BurnTokenID.String() != req.FromToken {
				c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid metadata BurnForCallRequest BurnTokenID must be FromToken")})
				return
			}

			// validate buyToken must be midToken
			if md.Data[0].ReceiveToken != req.MidToken {
				c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid metadata BurnForCallRequest Data ReceiveToken must be MidToken")})
				return
			}

			// the first papp tx must be reshield
			if md.Data[0].WithdrawAddress != EmptyExternalAddress {
				c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid metadata BurnForCallRequest Data WithdrawAddress")})
				return
			}

			// can't verify RedepositReceiver belong to ISIncPrivKey or not
			// only check UTXOs in response tx
			// isValid, err := IsValidOTA(md.Data[0].RedepositReceiver, config.ISIncPrivKey)
			// if err != nil || !isValid {
			// 	c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid metadata BurnForCallRequest RedepositReceiver")})
			// 	return
			// }

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
				return
			}
			// Note: Don't use worker

			// msgParam := interswap.InterswapSubmitTxTask{
			// 	TxID:       txHash,
			// 	TxRawBytes: rawTxBytes,

			// 	OTARefundFee: req.OTARefundFee,
			// 	OTARefund:    req.OTARefund,
			// 	OTAFromToken: req.OTAFromToken,
			// 	OTAToToken:   req.OTAToToken,

			// 	FromToken:           md.BurnTokenID.String(),
			// 	ToToken:             req.ToToken,
			// 	MidToken:            req.MidToken,
			// 	PathType:            interswap.PAppToPdex,
			// 	FinalMinExpectedAmt: req.FinalMinExpectedAmt,

			// 	PAppName:     req.PAppName,
			// 	PAppNetwork:  req.PAppNetwork,
			// 	PAppContract: req.PAppContract,

			// 	Status:    status,
			// 	StatusStr: interswap.StatusStr[status],
			// 	UserAgent: userAgent,
			// 	Error:     "",
			// }

			// // call assigner to publish msg to Interswap Worker
			// _, err = submitproof.PublishMsgInterswapTx(msgParam)
			// if err != nil {
			// 	// TODO: Update DB status failed
			// 	c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			// 	return
			// }

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

			// validate sellToken
			if md.TokenToSell.String() != req.FromToken {
				c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid metadata TradeRequest TokenToSell must be FromToken")})
				return
			}

			// validate buyToken must be midToken
			if buyTokenID.String() != req.MidToken {
				c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid metadata TradeRequest buyTokenID must be MidToken")})
				return
			}
			// isValid, err := IsValidOTA(md.Receiver[buyTokenID], config.ISIncPrivKey)
			// if err != nil || !isValid {
			// 	c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid metadata TradeRequest RedepositReceiver")})
			// 	return
			// }

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

			// Note: Don't use worker

			// msgParam := interswap.InterswapSubmitTxTask{
			// 	TxID:       txHash,
			// 	TxRawBytes: rawTxBytes,

			// 	OTARefundFee: req.OTARefundFee,
			// 	OTARefund:    req.OTARefund,
			// 	OTAFromToken: req.OTAFromToken,
			// 	OTAToToken:   req.OTAToToken,

			// 	FromToken:           req.FromToken,
			// 	ToToken:             req.ToToken,
			// 	MidToken:            req.MidToken,
			// 	PathType:            interswap.PdexToPApp,
			// 	FinalMinExpectedAmt: req.FinalMinExpectedAmt,

			// 	PAppName:     req.PAppName,
			// 	PAppNetwork:  req.PAppNetwork,
			// 	PAppContract: req.PAppContract,

			// 	Status:    interswap.FirstPending,
			// 	StatusStr: interswap.StatusStr[status],
			// 	UserAgent: userAgent,
			// 	Error:     "",
			// }

			// // call assigner to publish msg to Interswap Worker
			// _, err = submitproof.PublishMsgInterswapTx(msgParam)
			// if err != nil {
			// 	// TODO: Update DB status failed
			// 	c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			// 	return
			// }

			c.JSON(200, gin.H{"Result": map[string]interface{}{"inc_request_tx_status": interswap.FirstPending}})
			return
		}
	default:
		{
			c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Errorf("invalid metadata %v", mdType)})
			return
		}
	}
}

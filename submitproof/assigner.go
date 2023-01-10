package submitproof

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"strconv"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/interswap"
	"github.com/incognitochain/incognito-web-based-backend/slacknoti"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

func StartAssigner(cfg common.Config, serviceID uuid.UUID) error {
	config = cfg

	err := startPubsubClient(cfg.GGCProject, cfg.GGCAuth)
	if err != nil {
		return err
	}

	shieldTxTopic, err = startPubsubTopic(cfg.NetworkID + "_" + SHIELD_TX_TOPIC)
	if err != nil {
		panic(err)
	}

	unshieldTxTopic, err = startPubsubTopic(cfg.NetworkID + "_" + UNSHIELD_TX_TOPIC)
	if err != nil {
		panic(err)
	}

	pappTxTopic, err = startPubsubTopic(cfg.NetworkID + "_" + PAPP_TX_TOPIC)
	if err != nil {
		panic(err)
	}

	interSwapTxTopic, err = startPubsubTopic(cfg.NetworkID + "_" + INTERSWAP_TX_TOPIC)
	if err != nil {
		panic(err)
	}

	return nil
}

func SubmitShieldProof(txhash string, networkID int, tokenID string, txtype string, retry bool) (interface{}, error) {
	if networkID == 0 {
		return "", errors.New("unsupported network")
	}

	currentStatus, err := database.DBGetShieldTxStatusByExternalTx(txhash, networkID)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return "", err
		}
	}
	if currentStatus != "" {
		if currentStatus != common.StatusSubmitFailed || !retry {
			return currentStatus, nil
		}
	}

	task := SubmitProofShieldTask{
		TxHash:    txhash,
		NetworkID: networkID,
		TokenID:   tokenID,
		Metatype:  txtype,
	}
	taskBytes, _ := json.Marshal(task)

	ctx := context.Background()
	msg := &pubsub.Message{
		Attributes: map[string]string{
			"txhash":    txhash,
			"networkid": strconv.Itoa(networkID),
		},
		Data: taskBytes,
	}
	msgID, err := shieldTxTopic.Publish(ctx, msg).Get(ctx)
	if err != nil {
		return nil, err
	}
	log.Println("publish msgID:", msgID)

	return "submitting", nil
}

func SubmitPappTx(txhash string, rawTxData []byte, isPRVTx bool, feeToken string, feeAmount uint64, pfeeAmount uint64, burntToken string, burntAmount uint64, swapInfo *common.PappSwapInfo, isUnifiedToken bool, networks []string, refundFeeOTA string, refundFeeAddress string, userAgent string, txType int) (interface{}, error) {
	currentStatus, err := database.DBGetPappTxStatus(txhash)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return "", err
		}
	}
	if currentStatus != "" {
		return currentStatus, nil
	}

	task := SubmitPappTxTask{
		TxHash:           txhash,
		TxRawData:        rawTxData,
		IsPRVTx:          isPRVTx,
		IsUnifiedToken:   isUnifiedToken,
		FeeToken:         feeToken,
		FeeAmount:        feeAmount,
		PFeeAmount:       pfeeAmount,
		FeeRefundOTA:     refundFeeOTA,
		FeeRefundAddress: refundFeeAddress,
		BurntToken:       burntToken,
		BurntAmount:      burntAmount,
		PappSwapInfo:     swapInfo,
		Networks:         networks,
		Time:             time.Now(),
		UserAgent:        userAgent,
		TxType:           txType,
	}
	taskBytes, _ := json.Marshal(task)

	ctx := context.Background()
	msg := &pubsub.Message{
		Attributes: map[string]string{
			"txhash": txhash,
			"task":   PappSubmitIncTask,
		},
		Data: taskBytes,
	}
	msgID, err := pappTxTopic.Publish(ctx, msg).Get(ctx)
	if err != nil {
		return nil, err
	}
	log.Println("publish msgID:", msgID)
	txTypeStr := "unknown"
	switch txType {
	case common.ExternalTxTypeSwap:
		txTypeStr = "txswap"
	case common.ExternalTxTypeOpenseaBuy:
		txTypeStr = "opensea"
	case common.ExternalTxTypeOpenseaOffer:
		txTypeStr = "opensea-offer"
	case common.ExternalTxTypeOpenseaOfferCancel:
		txTypeStr = "opensea-cancel"
	}
	go func() {
		tkInfo, _ := getTokenInfo(feeToken)
		amount := new(big.Float).SetUint64(feeAmount)
		decimal := new(big.Float).SetFloat64(math.Pow10(-tkInfo.PDecimals))
		afl64, _ := amount.Mul(amount, decimal).Float64()
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[%v]` tx `%v` approved with fee `%f %v` 👌", txTypeStr, txhash, afl64, tkInfo.Symbol))
	}()

	return "submitting", nil
}

func SubmitTxFeeRefund(incReqTx, refundOTA, paymentAddress, token string, amount uint64, isPrivacyFeeRefund bool) (interface{}, error) {
	task := SubmitRefundFeeTask{
		IncReqTx:           incReqTx,
		OTA:                refundOTA,
		PaymentAddress:     paymentAddress,
		Amount:             amount,
		Token:              token,
		IsPrivacyFeeRefund: isPrivacyFeeRefund,
		Time:               time.Now(),
	}
	taskBytes, _ := json.Marshal(task)

	ctx := context.Background()
	msg := &pubsub.Message{
		Attributes: map[string]string{
			"txhash": incReqTx,
			"task":   PappSubmitFeeRefundTask,
		},
		Data: taskBytes,
	}
	msgID, err := pappTxTopic.Publish(ctx, msg).Get(ctx)
	if err != nil {
		return nil, err
	}
	log.Println("publish msgID:", msgID)

	return "submitting", nil
}

func SubmitOutChainTx(incTxHash string, network string, isUnifiedToken bool, retry bool, txType int) (interface{}, error) {
	currentStatus, err := database.DBGetExternalTxStatusByIncTx(incTxHash, network)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return "", err
		}
	}
	if currentStatus != "" {
		if currentStatus != common.StatusSubmitFailed || !retry {
			return currentStatus, nil
		}
	}

	task := SubmitProofOutChainTask{
		IncTxhash:      incTxHash,
		Network:        network,
		IsUnifiedToken: isUnifiedToken,
		IsRetry:        retry,
		Type:           txType,
		Time:           time.Now(),
	}
	taskBytes, _ := json.Marshal(task)

	ctx := context.Background()
	switch txType {
	case common.ExternalTxTypeSwap, common.ExternalTxTypeOpenseaBuy, common.ExternalTxTypeOpenseaOffer, common.ExternalTxTypeOpenseaOfferCancel:
		msg := &pubsub.Message{
			Attributes: map[string]string{
				"txhash": incTxHash,
				"task":   PappSubmitExtTask,
			},
			Data: taskBytes,
		}
		msgID, err := pappTxTopic.Publish(ctx, msg).Get(ctx)
		if err != nil {
			return nil, err
		}
		log.Println("publish msgID:", msgID)
		break
	case common.ExternalTxTypeUnshield:
		msg := &pubsub.Message{
			Attributes: map[string]string{
				"txhash": incTxHash,
				"task":   UnshieldSubmitExtTask,
			},
			Data: taskBytes,
		}
		msgID, err := unshieldTxTopic.Publish(ctx, msg).Get(ctx)
		if err != nil {
			return nil, err
		}
		log.Println("publish msgID:", msgID)
		break
	default:
		log.Println("unknown txType")
	}

	return "submitting", nil
}

func SubmitUnshieldTx(txhash string, rawTxData []byte, isPRVTx bool, feeToken string, feeAmount uint64, pfeeAmount uint64, tokenID, uTokenID string, burntAmount uint64, isUnifiedToken bool, externalAddress string, networks []string, refundFeeOTA string, refundFeeAddress string, userAgent string) (interface{}, error) {
	currentStatus, err := database.DBGetUnshieldTxStatusByIncTx(txhash)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return "", err
		}
	}
	if currentStatus != "" {
		return currentStatus, nil
	}

	task := SubmitUnshieldTxTask{
		TxHash:           txhash,
		TxRawData:        rawTxData,
		IsPRVTx:          isPRVTx,
		IsUnifiedToken:   isUnifiedToken,
		FeeToken:         feeToken,
		FeeAmount:        feeAmount,
		PFeeAmount:       pfeeAmount,
		FeeRefundOTA:     refundFeeOTA,
		FeeRefundAddress: refundFeeAddress,
		Token:            tokenID,
		UToken:           uTokenID,
		BurntAmount:      burntAmount,
		ExternalAddress:  externalAddress,
		Networks:         networks,
		Time:             time.Now(),
		UserAgent:        userAgent,
	}
	taskBytes, _ := json.Marshal(task)

	ctx := context.Background()
	msg := &pubsub.Message{
		Attributes: map[string]string{
			"txhash": txhash,
			"task":   UnshieldSubmitIncTask,
		},
		Data: taskBytes,
	}
	msgID, err := unshieldTxTopic.Publish(ctx, msg).Get(ctx)
	if err != nil {
		return nil, err
	}
	log.Println("publish msgID:", msgID)
	go func() {
		tkInfo, _ := getTokenInfo(feeToken)
		amount := new(big.Float).SetUint64(feeAmount)
		decimal := new(big.Float).SetFloat64(math.Pow10(-tkInfo.PDecimals))
		afl64, _ := amount.Mul(amount, decimal).Float64()
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[unshield]` unshield `%v` approved with fee `%f %v` 👌", txhash, afl64, tkInfo.Symbol))
	}()

	return "submitting", nil
}

func PublishMsgInterswapTx(
	task interswap.InterswapSubmitTxTask,
) (interface{}, error) {
	taskBytes, _ := json.Marshal(task)

	taskType := interswap.InterswapPdexPappTxTask
	if task.PathType == interswap.PAppToPdex {
		taskType = interswap.InterswapPappPdexTask
	}

	ctx := context.Background()
	msg := &pubsub.Message{
		Attributes: map[string]string{
			"txhash": task.TxID,
			"task":   taskType,
		},
		Data: taskBytes,
	}
	msgID, errPub := interSwapTxTopic.Publish(ctx, msg).Get(ctx)
	if errPub != nil {
		return nil, errPub
	}
	log.Println("publish msgID:", msgID)
	go func() {
		// tkInfo, _ := getTokenInfo(feeToken)
		// amount := new(big.Float).SetUint64(feeAmount)
		// decimal := new(big.Float).SetFloat64(math.Pow10(-tkInfo.PDecimals))
		// afl64, _ := amount.Mul(amount, decimal).Float64()
		// go slacknoti.SendSlackNoti(fmt.Sprintf("`[swaptx]` swaptx `%v` approved with fee `%f %v` 👌", txhash, afl64, tkInfo.Symbol))
	}()

	return "submitting", nil
}

func SubmitPdaoOutchainTx(incTxHash string, network string, payload []byte, retry bool, reqType, txType int) (interface{}, error) {
	incTxHash = incTxHash + "_" + strconv.Itoa(reqType)
	currentStatus, err := database.DBGetExternalTxStatusByIncTx(incTxHash, network)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return "", err
		}
	}
	if currentStatus != "" {
		if currentStatus != common.StatusSubmitFailed || !retry {
			return currentStatus, nil
		}
	}

	task := SubmitPDaoTask{
		IncTxhash: incTxHash,
		Network:   network,
		ReqType:   reqType,
		IsRetry:   retry,
		Type:      txType,
		Payload:   payload,
		Time:      time.Now(),
	}
	taskBytes, _ := json.Marshal(task)

	ctx := context.Background()
	switch txType {
	case common.ExternalTxTypePdaoProposal:
		msg := &pubsub.Message{
			Attributes: map[string]string{
				"txhash": incTxHash,
				"task":   PdaoSubmitProposalExtTask,
			},
			Data: taskBytes,
		}
		msgID, err := pappTxTopic.Publish(ctx, msg).Get(ctx)
		if err != nil {
			return nil, err
		}
		log.Println("publish msgID:", msgID)
		break
	case common.ExternalTxTypePdaoVote:
		msg := &pubsub.Message{
			Attributes: map[string]string{
				"txhash": incTxHash,
				"task":   PdaoSubmitVoteExtTask,
			},
			Data: taskBytes,
		}
		msgID, err := pappTxTopic.Publish(ctx, msg).Get(ctx)
		if err != nil {
			return nil, err
		}
		log.Println("publish msgID:", msgID)
		break
	case common.ExternalTxTypePdaoCancel:
		//TODO: @phuong
		msg := &pubsub.Message{
			Attributes: map[string]string{
				"txhash": incTxHash,
				"task":   PdaoSubmitCancelExtTask,
			},
			Data: taskBytes,
		}
		msgID, err := pappTxTopic.Publish(ctx, msg).Get(ctx)
		if err != nil {
			return nil, err
		}
		log.Println("publish msgID:", msgID)
		break

	case common.ExternalTxTypePdaoReShieldPRV:
		msg := &pubsub.Message{
			Attributes: map[string]string{
				"txhash": incTxHash,
				"task":   PdaoSubmitReShieldPRVExtTask,
			},
			Data: taskBytes,
		}
		msgID, err := pappTxTopic.Publish(ctx, msg).Get(ctx)
		if err != nil {
			return nil, err
		}
		log.Println("publish msgID:", msgID)
		break
	default:
		log.Println("unknown txType")
	}

	return "submitting", nil
}

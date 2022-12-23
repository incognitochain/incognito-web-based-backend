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

	pappTxTopic, err = startPubsubTopic(cfg.NetworkID + "_" + PAPP_TX_TOPIC)
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

// TODO: 0xkraken write another one for IS to publish msg
func SubmitPappTx(txhash string, rawTxData []byte, isPRVTx bool, feeToken string, feeAmount uint64, pfeeAmount uint64, burntToken string, burntAmount uint64, swapInfo *common.PappSwapInfo, isUnifiedToken bool, networks []string, refundFeeOTA string, refundFeeAddress string, userAgent string) (interface{}, error) {
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

		// add more

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
	go func() {
		tkInfo, _ := getTokenInfo(feeToken)
		amount := new(big.Float).SetUint64(feeAmount)
		decimal := new(big.Float).SetFloat64(math.Pow10(-tkInfo.PDecimals))
		afl64, _ := amount.Mul(amount, decimal).Float64()
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[swaptx]` swaptx `%v` approved with fee `%f %v` 👌", txhash, afl64, tkInfo.Symbol))
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

func SubmitOutChainPappTx(incTxHash string, network string, isUnifiedToken bool, retry bool) (interface{}, error) {
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

	task := SubmitPappProofOutChainTask{
		IncTxhash:      incTxHash,
		Network:        network,
		IsUnifiedToken: isUnifiedToken,
		IsRetry:        retry,
		Time:           time.Now(),
	}
	taskBytes, _ := json.Marshal(task)

	ctx := context.Background()
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

	return "submitting", nil
}

package submitproof

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

func StartAssigner(cfg common.Config, serviceID uuid.UUID) error {
	config = cfg

	err := startPubsubClient(cfg.GGCProject, cfg.GGCAuth)
	if err != nil {
		return err
	}

	shieldTxTopic, err = startPubsubTopic(SHIELD_TX_TOPIC)
	if err != nil {
		panic(err)
	}

	pappTxTopic, err = startPubsubTopic(PAPP_TX_TOPIC)
	if err != nil {
		panic(err)
	}

	return nil
}

func SubmitShieldProof(txhash string, networkID int, tokenID string, txtype string) (interface{}, error) {
	if networkID == 0 {
		return "", errors.New("unsported network")
	}

	currentStatus, err := database.DBGetShieldTxStatusByExternalTx(txhash, networkID)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return "", err
		}
	}
	if currentStatus != "" {
		return currentStatus, nil
	}

	task := SubmitProofShieldTask{
		TxHash:    txhash,
		NetworkID: networkID,
		TokenID:   tokenID,
		Metatype:  txtype,
		Time:      time.Now(),
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

func SubmitPappTx(txhash string, rawTxData []byte, isPRVTx bool, feeToken string, feeAmount uint64, burntToken string, burntAmount uint64, isUnifiedToken bool, networks []string, refundFeeOTA string, refundFeeAddress string) (interface{}, error) {
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
		FeeRefundOTA:     refundFeeOTA,
		FeeRefundAddress: refundFeeAddress,
		BurntToken:       burntToken,
		BurntAmount:      burntAmount,
		Networks:         networks,
		Time:             time.Now(),
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

	return "submitting", nil
}

func SubmitTxFeeRefund(incReqTx, refundOTA, refundOTASS, paymentAddress, token string, amount uint64) (interface{}, error) {
	task := SubmitRefundFeeTask{
		IncReqTx:       incReqTx,
		OTA:            refundOTA,
		OTASS:          refundOTASS,
		PaymentAddress: paymentAddress,
		Amount:         amount,
		Token:          token,
		Time:           time.Now(),
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

func SendOutChainPappTx(incTxHash string, network string, isUnifiedToken bool) (interface{}, error) {
	currentStatus, err := database.DBGetExternalTxStatusByIncTx(incTxHash, network)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return "", err
		}
	}
	if currentStatus != "" {
		return currentStatus, nil
	}

	task := SubmitPappProofOutChainTask{
		IncTxhash:      incTxHash,
		Network:        network,
		IsUnifiedToken: isUnifiedToken,
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

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

	swapTxTopic, err = startPubsubTopic(SWAP_TX_TOPIC)
	if err != nil {
		panic(err)
	}

	return nil
}

func SubmitShieldProof(txhash string, networkID int, tokenID string) (interface{}, error) {
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
		Metatype:  TxTypeShielding,
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

func SubmitSwapTx(txhash string, rawTxData []byte, isPRVTx bool, feeToken string, feeAmount uint64) (interface{}, error) {
	currentStatus, err := database.DBGetPappTxStatus(txhash)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return "", err
		}
	}
	if currentStatus != "" {
		return currentStatus, nil
	}

	task := SubmitPappSwapTask{
		TxHash:    txhash,
		TxRawData: rawTxData,
		IsPRVTx:   isPRVTx,
		FeeToken:  feeToken,
		FeeAmount: feeAmount,
		Time:      time.Now(),
	}
	taskBytes, _ := json.Marshal(task)

	ctx := context.Background()
	msg := &pubsub.Message{
		Attributes: map[string]string{
			"txhash": txhash,
		},
		Data: taskBytes,
	}
	msgID, err := swapTxTopic.Publish(ctx, msg).Get(ctx)
	if err != nil {
		return nil, err
	}
	log.Println("publish msgID:", msgID)

	return "submitting", nil
}

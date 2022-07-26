package submitproof

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/pkg/errors"
)

func StartAssigner(cfg common.Config, serviceID uuid.UUID) error {
	config = cfg

	err := connectDB(cfg.DatabaseURLs)
	if err != nil {
		return err
	}

	err = connectMQ(serviceID, cfg.DatabaseURLs)
	if err != nil {
		return err
	}
	return nil
}

func SubmitShieldProof(txhash string, networkID int, tokenID string) (interface{}, error) {
	if networkID == 0 {
		return "", errors.New("unsported network")
	}

	currentStatus, err := getShieldTxStatus(txhash, networkID, tokenID)
	if err != nil {
		return "", err
	}
	if currentStatus != ShieldStatusUnknown {
		return ShieldStatusMap[currentStatus], nil
	}

	task := SubmitProofShieldTask{
		Txhash:    txhash,
		NetworkID: networkID,
		TokenID:   tokenID,
		Metatype:  TxTypeShielding,
		Time:      time.Now(),
	}
	taskBytes, _ := json.Marshal(task)

	taskQueue, err := rdmq.OpenQueue(MqSubmitTx)
	if err != nil {
		return nil, err
	}

	err = taskQueue.PublishBytes(taskBytes)
	if err != nil {
		return nil, err
	}
	// go submitProof(txhash, tokenID, networkID)
	return "submitting", nil
}

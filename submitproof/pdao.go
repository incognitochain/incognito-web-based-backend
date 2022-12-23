package submitproof

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/pdao"
	"github.com/incognitochain/incognito-web-based-backend/slacknoti"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

func ProcessPdaoRequest(ctx context.Context, m *pubsub.Message) {
	taskDesc := m.Attributes["task"]
	switch taskDesc {
	case PdaoSubmitExtTask:
		processSubmitPdaoRequest(ctx, m)
	case PrvSubmitExtTask:
		// shield by signature
		// re-shield from pdao
		processSubmitPrvRequest(ctx, m)
	}
}

func processSubmitPdaoRequest(ctx context.Context, m *pubsub.Message) {
	//TODO
	task := SubmitPDaoTask{}
	err := json.Unmarshal(m.Data, &task)
	if err != nil {
		log.Println("processSubmitPdaoRequest error decoding message", err)
		m.Ack()
		return
	}

	if time.Since(m.PublishTime) > time.Hour {
		status := wcommon.ExternalTxStatus{
			IncRequestTx: task.IncTxhash,
			Type:         wcommon.ExternalTxTypeUnshield,
			Status:       wcommon.StatusSubmitFailed,
			Network:      task.Network,
			Error:        "timeout",
		}
		// todo thachtb: update here
		err = database.DBSaveExternalTxStatus(&status)
		if err != nil {
			writeErr, ok := err.(mongo.WriteException)
			if !ok {
				log.Println("DBSaveExternalTxStatus err", err)
				m.Nack()
				return
			}
			if !writeErr.HasErrorCode(11000) {
				log.Println("DBSaveExternalTxStatus err", err)
				m.Nack()
				return
			}
		}
		// todo: update here
		err = database.DBUpdatePappTxSubmitOutchainStatus(task.IncTxhash, wcommon.StatusSubmitFailed)
		if err != nil {
			writeErr, ok := err.(mongo.WriteException)
			if !ok {
				log.Println("DBSaveExternalTxStatus err", err)
				m.Nack()
				return
			}
			if !writeErr.HasErrorCode(11000) {
				log.Println("DBSaveExternalTxStatus err", err)
				m.Nack()
				return
			}
		}
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[unshield]` submitProofTx timeout 😵 inctx `%v` network `%v`\n", task.IncTxhash, task.Network))
		return
	}
	status, err := pdao.CreateGovernanceOutChainTx(task.Network, task.IncTxhash, task.Payload, uint8(task.Type), config)
	if err != nil {
		log.Println("createOutChainUnshieldTx error", err)
		time.Sleep(15 * time.Second)
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[unshield]` submitProofTx `%v` for network `%v` failed 😵 err: %v", task.IncTxhash, task.Network, err))
		m.Ack()
		return
	}
	go slacknoti.SendSlackNoti(fmt.Sprintf("`[unshield]` submitProofTx `%v` for network `%v` success 👌 txhash `%v`", task.IncTxhash, task.Network, status.Txhash))

	// todo thachtb: update here
	err = database.DBSaveExternalTxStatus(status)
	if err != nil {
		writeErr, ok := err.(mongo.WriteException)
		if !ok {
			log.Println("DBSaveExternalTxStatus err", err)
			m.Ack()
			return
		}
		if !writeErr.HasErrorCode(11000) {
			log.Println("DBSaveExternalTxStatus err", err)
			m.Ack()
			return
		}
	}

	// todo thachtb: update here
	err = database.DBUpdateUnshieldTxSubmitOutchainStatus(task.IncTxhash, wcommon.StatusPending)
	if err != nil {
		log.Println("DBUpdateUnshieldTxSubmitOutchainStatus err", err)
		m.Ack()
		return
	}

	m.Ack()
}

func processSubmitPrvRequest(ctx context.Context, m *pubsub.Message) {
	//TODO
	task := SubmitPDaoTask{}
	err := json.Unmarshal(m.Data, &task)
	if err != nil {
		log.Println("processSubmitPrvRequest error decoding message", err)
		m.Ack()
		return
	}

	if time.Since(m.PublishTime) > time.Hour {
		status := wcommon.ExternalTxStatus{
			IncRequestTx: task.IncTxhash,
			Type:         wcommon.ExternalTxTypeUnshield,
			Status:       wcommon.StatusSubmitFailed,
			Network:      task.Network,
			Error:        "timeout",
		}
		// todo: update here
		err = database.DBSaveExternalTxStatus(&status)
		if err != nil {
			writeErr, ok := err.(mongo.WriteException)
			if !ok {
				log.Println("DBSaveExternalTxStatus err", err)
				m.Nack()
				return
			}
			if !writeErr.HasErrorCode(11000) {
				log.Println("DBSaveExternalTxStatus err", err)
				m.Nack()
				return
			}
		}
		// todo: update here
		err = database.DBUpdatePappTxSubmitOutchainStatus(task.IncTxhash, wcommon.StatusSubmitFailed)
		if err != nil {
			writeErr, ok := err.(mongo.WriteException)
			if !ok {
				log.Println("DBSaveExternalTxStatus err", err)
				m.Nack()
				return
			}
			if !writeErr.HasErrorCode(11000) {
				log.Println("DBSaveExternalTxStatus err", err)
				m.Nack()
				return
			}
		}
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[unshield]` submitProofTx timeout 😵 inctx `%v` network `%v`\n", task.IncTxhash, task.Network))
		return
	}

	status, err := pdao.CreatePRVOutChainTx(task.Network, task.IncTxhash, task.Payload, uint8(task.Type), config)
	if err != nil {
		log.Println("createOutChainUnshieldTx error", err)
		time.Sleep(15 * time.Second)
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[unshield]` submitProofTx `%v` for network `%v` failed 😵 err: %v", task.IncTxhash, task.Network, err))
		m.Ack()
		return
	}
	go slacknoti.SendSlackNoti(fmt.Sprintf("`[unshield]` submitProofTx `%v` for network `%v` success 👌 txhash `%v`", task.IncTxhash, task.Network, status.Txhash))

	// todo: update here
	err = database.DBSaveExternalTxStatus(status)
	if err != nil {
		writeErr, ok := err.(mongo.WriteException)
		if !ok {
			log.Println("DBSaveExternalTxStatus err", err)
			m.Ack()
			return
		}
		if !writeErr.HasErrorCode(11000) {
			log.Println("DBSaveExternalTxStatus err", err)
			m.Ack()
			return
		}
	}

	// todo: update here
	err = database.DBUpdateUnshieldTxSubmitOutchainStatus(task.IncTxhash, wcommon.StatusPending)
	if err != nil {
		log.Println("DBUpdateUnshieldTxSubmitOutchainStatus err", err)
		m.Ack()
		return
	}

	m.Ack()
}

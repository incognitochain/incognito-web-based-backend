package submitproof

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/pubsub"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/pdao"
	"github.com/incognitochain/incognito-web-based-backend/slacknoti"
	"go.mongodb.org/mongo-driver/mongo"
)

func processSubmitPdaoRequest(ctx context.Context, m *pubsub.Message) {
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
			Type:         task.Type,
			Status:       wcommon.StatusSubmitFailed,
			Network:      task.Network,
			Error:        "timeout",
		}
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
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[pdao]` submitProofTx timeout 😵 inctx `%v` network `%v`\n", task.IncTxhash, task.Network))
		return
	}
	status, err := pdao.CreateGovernanceOutChainTx(task.Network, task.IncTxhash, task.Payload, uint8(task.ReqType), config, task.Type)
	if err != nil {
		log.Println("CreateGovernanceOutChainTx error", err)
		time.Sleep(15 * time.Second)
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[pdao]` submitProofTx `%v` for network `%v` failed 😵 err: %v", task.IncTxhash, task.Network, err))
		m.Ack()
		return
	}
	go slacknoti.SendSlackNoti(fmt.Sprintf("`[pdao]` submitProofTx `%v` for network `%v` success 👌 txhash `%v`", task.IncTxhash, task.Network, status.Txhash))

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

	proposal, err := database.DBGetProposalByIncTx(task.IncTxhash)
	if err != nil {
		log.Println("DBGetProposalByIncTx error", err)
		m.Ack()
		return
	}

	proposal.SubmitProposalTx = status.Txhash
	proposal.Status = wcommon.StatusPdaOutchainTxPending
	err = database.DBUpdateProposalTable(proposal)
	if err != nil {
		log.Println("DBUpdateProposalTable err:", err)
		m.Ack()
		return
	}

	m.Ack()
}

func processSubmitPrvRequest(ctx context.Context, m *pubsub.Message) {
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
			Type:         task.Type,
			Status:       wcommon.StatusSubmitFailed,
			Network:      task.Network,
			Error:        "timeout",
		}
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

	status, err := pdao.CreatePRVOutChainTx(task.Network, task.IncTxhash, task.Payload, uint8(task.ReqType), config, task.Type)
	if err != nil {
		log.Println("CreatePRVOutChainTx error", err)
		time.Sleep(15 * time.Second)
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[unshield]` submitProofTx `%v` for network `%v` failed 😵 err: %v", task.IncTxhash, task.Network, err))
		m.Ack()
		return
	}
	go slacknoti.SendSlackNoti(fmt.Sprintf("`[unshield]` submitProofTx `%v` for network `%v` success 👌 txhash `%v`", task.IncTxhash, task.Network, status.Txhash))

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

	// todo: @phuong update here Vote?

	// proposal, err := database.DBGetProposalByIncTx(task.IncTxhash)
	// if err != nil {
	// 	log.Println("DBGetProposalByIncTx error", err)
	// 	m.Ack()
	// 	return
	// }

	// proposal.SubmitProposalTx = status.Txhash
	// proposal.Status = wcommon.StatusPdaOutchainTxPending
	// err = database.DBUpdateProposalTable(proposal)
	// if err != nil {
	// 	log.Println("DBUpdateProposalTable err:", err)
	// 	continue
	// }

	m.Ack()
}

func watchPendingProposal() {
	for {
		time.Sleep(10 * time.Second)
		proposals, err := database.DBGetPendingProposal()
		if err != nil {
			log.Println("DBRetrievePendingShieldTxs err:", err)
		}
		for _, p := range proposals {
			unshieldStatus, err := database.DBGetUnshieldTxByIncTx(p.SubmitBurnTx)
			if err != nil {
				log.Println("DBGetUnshieldTxStatusByIncTx err:", err)
				continue
			}
			if unshieldStatus.Status == wcommon.StatusSubmitFailed {
				p.Status = unshieldStatus.Status
				err = database.DBUpdateProposalTable(&p)
				if err != nil {
					log.Println("DBUpdateProposalTable err:", err)
					continue
				}
			}
			if unshieldStatus.OutchainStatus == wcommon.StatusAccepted {
				p.Status = wcommon.StatusPdaOutchainTxSubmitting
				err = database.DBUpdateProposalTable(&p)
				if err != nil {
					log.Println("DBUpdateProposalTable err:", err)
					continue
				}
				_, err := SubmitPdaoOutchainTx(p.SubmitBurnTx, unshieldStatus.Networks[0], false, pdao.CREATE_PROP, wcommon.ExternalTxTypePdaoProposal)
				if err != nil {
					log.Println("SubmitPdaoOutchainTx err:", err)
					continue
				}
			}
		}
	}
}

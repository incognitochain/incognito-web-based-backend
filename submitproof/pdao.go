package submitproof

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/pdao"
	"github.com/incognitochain/incognito-web-based-backend/pdao/governance"
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
	inctx := strings.Split(task.IncTxhash, "_")
	proposal, err := database.DBGetProposalByIncTx(inctx[0])
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

func processSubmitVoteRequest(ctx context.Context, m *pubsub.Message) {
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
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[pdao-vote]` submitProofTx timeout 😵 inctx `%v` network `%v`\n", task.IncTxhash, task.Network))
		return
	}

	status, err := pdao.CreateGovernanceOutChainTx(task.Network, task.IncTxhash, task.Payload, uint8(task.ReqType), config, task.Type)
	if err != nil {
		log.Println("CreateGovernanceOutChainTx error", err)
		time.Sleep(15 * time.Second)
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[pdao-vote]` submitProofTx `%v` for network `%v` failed 😵 err: %v", task.IncTxhash, task.Network, err))
		m.Ack()
		return
	}
	go slacknoti.SendSlackNoti(fmt.Sprintf("`[pdao-vote]` submitProofTx `%v` for network `%v` success 👌 txhash `%v`", task.IncTxhash, task.Network, status.Txhash))

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
	inctx := strings.Split(task.IncTxhash, "_")

	vote, err := database.DBGetVoteByIncTx(inctx[0])
	if err != nil {
		log.Println("DBGetVoteByIncTx error", err)
		m.Ack()
		return
	}

	vote.SubmitVoteTx = status.Txhash
	vote.Status = wcommon.StatusPdaOutchainTxPending
	err = database.DBUpdateVoteTable(vote)
	if err != nil {
		log.Println("DBUpdateVoteTable err:", err)
		m.Ack()
		return
	}

	m.Ack()
}

func watchPendingProposal() {
	for {
		time.Sleep(10 * time.Second)
		proposals, err := database.DBGetPendingProposal()
		if err != nil {
			log.Println("DBGetPendingProposal err:", err)
		}
		for _, p := range proposals {
			unshieldStatus, err := database.DBGetUnshieldTxByIncTx(p.SubmitBurnTx)
			if err != nil {
				log.Println("DBGetUnshieldTxByIncTx err:", err)
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

				proposalJson, err := json.Marshal(p)
				if err != nil {
					log.Println("marshal proposal err:", err)
					continue
				}
				p.Status = wcommon.StatusPdaOutchainTxSubmitting
				err = database.DBUpdateProposalTable(&p)
				if err != nil {
					log.Println("DBUpdateProposalTable err:", err)
					continue
				}

				_, err = SubmitPdaoOutchainTx(p.SubmitBurnTx, unshieldStatus.Networks[0], proposalJson, false, pdao.CREATE_PROP, wcommon.ExternalTxTypePdaoProposal)
				if err != nil {
					log.Println("SubmitPdaoOutchainTx err:", err)
					continue
				}
			}
		}
	}
}

// this watcher for checking proposal success and check vote date to auto vote:
func watchSuccessProposal() {
	for {

		time.Sleep(10 * time.Second)

		go slacknoti.SendSlackNoti("checking proposal auto vote....")

		proposals, err := database.DBGetSuccessProposalNoVoted()
		if err != nil {
			go slacknoti.SendSlackNoti("watchSuccessProposal DBGetSuccessProposalNoVoted err:" + err.Error())
		}
		log.Println("there are ", len(proposals), "records!")

		networkInfo, err := database.DBGetBridgeNetworkInfo(wcommon.NETWORK_ETH)

		evmClient, err := ethclient.Dial(networkInfo.Endpoints[1])
		if err != nil {
			go slacknoti.SendSlackNoti("watchSuccessProposal DBGetSuccessProposalNoVoted err:" + err.Error())
			continue
		}

		gv, err := governance.NewGovernance(common.HexToAddress(wcommon.GOVERNANCE_CONTRACT_ADDRESS), evmClient)
		if err != nil {
			continue
		}

		for _, p := range proposals {

			proposalID, ok := big.NewInt(0).SetString(p.ProposalID, 10)

			if !ok {
				go slacknoti.SendSlackNoti("watchSuccessProposal parse  ProposalID no ok")
				continue
			}

			prop, err := gv.Proposals(nil, proposalID)
			if err != nil {
				go slacknoti.SendSlackNoti("watchSuccessProposal Proposals err:" + err.Error() + ", proposalID: " + p.ProposalID + networkInfo.Endpoints[1])
				continue
			}

			header, err := evmClient.HeaderByNumber(context.Background(), nil)
			if err != nil {
				log.Println("watchSuccessProposal HeaderByNumber err:", err)
				continue
			}

			log.Println("watchSuccessProposal prop.StartBlock:", prop.StartBlock, "header.Number: ", header.Number)

			if prop.StartBlock.Cmp(header.Number) == -1 {

				// auto vote now (insert to vote):
				vote := &wcommon.Vote{
					ProposalID:        p.ProposalID,
					Status:            wcommon.StatusPdaOutchainTxSubmitting,
					Vote:              1,
					PropVoteSignature: p.PropVoteSignature,
					ReShieldSignature: p.ReShieldSignature,
					AutoVoted:         true,           // auto vote for owner of proposal.
					SubmitBurnTx:      p.SubmitBurnTx, // use proposal burn prv tx for tracking
				}

				voteJson, err := json.Marshal(vote)
				if err != nil {
					log.Println("marshal proposal err:", err)
					continue
				}

				_, err = SubmitPdaoOutchainTx(p.SubmitBurnTx, wcommon.NETWORK_ETH, voteJson, false, pdao.VOTE_PROP, wcommon.ExternalTxTypePdaoVote)
				if err != nil {
					go slacknoti.SendSlackNoti("watchSuccessProposal SubmitPdaoOutchainTx err:" + err.Error())
					continue
				}

				err = database.DBInsertVoteTable(vote)

				if err != nil {
					go slacknoti.SendSlackNoti("watchSuccessProposal DBInsertVoteTable err:" + err.Error())
					continue
				}
				// update proposal:
				p.Voted = true
				database.DBUpdateProposalTable(&p)

			}

		}
	}
}

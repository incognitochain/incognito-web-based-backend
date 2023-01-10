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

func processSubmitReShieldPRVRequest(ctx context.Context, m *pubsub.Message) {
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
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[pdao-reshield]` submitProofTx timeout 😵 inctx `%v` network `%v`\n", task.IncTxhash, task.Network))
		return
	}

	status, err := pdao.CreatePRVOutChainTx(task.Network, task.IncTxhash, task.Payload, uint8(task.ReqType), config, task.Type)
	if err != nil {
		log.Println("CreatePRVOutChainTx error", err)
		time.Sleep(15 * time.Second)
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[pdao-reshield]` submitProofTx `%v` for network `%v` failed 😵 err: %v", task.IncTxhash, task.Network, err))
		m.Ack()
		return
	}
	go slacknoti.SendSlackNoti(fmt.Sprintf("`[pdao-reshield]` submitProofTx `%v` for network `%v` success 👌 txhash `%v`", task.IncTxhash, task.Network, status.Txhash))

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

	//TODO update status reshield?
	vote, err := database.DBGetVoteByIncTx(inctx[0])
	if err != nil {
		log.Println("DBGetVoteByIncTx error", err)
		m.Ack()
		return
	}

	vote.SubmitReShieldTx = status.Txhash
	vote.ReShieldStatus = wcommon.StatusPdaOutchainTxPending
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
				log.Println("watchPendingProposal DBGetUnshieldTxByIncTx err:", err)
				continue
			}
			if unshieldStatus.Status == wcommon.StatusSubmitFailed {
				p.Status = unshieldStatus.Status
				err = database.DBUpdateProposalTable(&p)
				if err != nil {
					log.Println("watchPendingProposal DBUpdateProposalTable err:", err)
					continue
				}
			}
			if unshieldStatus.OutchainStatus == wcommon.StatusAccepted {
				proposalJson, err := json.Marshal(p)
				if err != nil {
					log.Println("marshal proposal err:", err)
					continue
				}
				// update status
				p.Status = wcommon.StatusPdaOutchainTxSubmitting
				err = database.DBUpdateProposalTable(&p)
				if err != nil {
					log.Println("watchPendingProposal DBUpdateProposalTable err:", err)
					continue
				}

				_, err = SubmitPdaoOutchainTx(p.SubmitBurnTx, unshieldStatus.Networks[0], proposalJson, false, pdao.CREATE_PROP, wcommon.ExternalTxTypePdaoProposal)
				if err != nil {
					log.Println("watchPendingProposal SubmitPdaoOutchainTx err:", err)
					continue
				}
			}
		}
	}
}

func watchPendingVoting() {
	for {
		time.Sleep(10 * time.Second)
		proposals, err := database.DBGetPendingVote()
		if err != nil {
			log.Println("watchPendingVoting DBGetPendingVote err:", err)
		}
		for _, p := range proposals {
			unshieldStatus, err := database.DBGetUnshieldTxByIncTx(p.SubmitBurnTx)
			if err != nil {
				log.Println("watchPendingVoting DBGetUnshieldTxByIncTx err:", err)
				continue
			}
			if unshieldStatus.Status == wcommon.StatusSubmitFailed {
				p.Status = unshieldStatus.Status
				err = database.DBUpdateVoteTable(&p)
				if err != nil {
					log.Println("watchPendingVoting DBUpdateVoteTable err:", err)
					continue
				}
			}
			if unshieldStatus.OutchainStatus == wcommon.StatusAccepted {
				// update status ready to vote
				p.Status = wcommon.StatusPdaoReadyForVote
				err = database.DBUpdateVoteTable(&p)
				if err != nil {
					log.Println("watchPendingVoting DBUpdateVoteTable err:", err)
					continue
				}
			}
		}
	}
}

func watchReadyToVote() {
	for {

		time.Sleep(60 * time.Second)

		proposals, err := database.DBGetReadyToVote()
		if err != nil {
			log.Println("watchReadyToVote DBGetReadyToVote err:" + err.Error())
		}
		log.Println("watchReadyToVote there are ", len(proposals), "records!")

		networkInfo, err := database.DBGetBridgeNetworkInfo(wcommon.NETWORK_ETH)
		if err != nil {
			log.Println("watchVotedToReshield DBGetBridgeNetworkInfo err:" + err.Error())
			continue
		}

		evmClient, err := ethclient.Dial(networkInfo.Endpoints[1])
		if err != nil {
			log.Println("watchReadyToVote DBGetSuccessProposalNoVoted err:" + err.Error())
			continue
		}

		papps, err := database.DBRetrievePAppsByNetwork("eth")
		if err != nil {
			log.Println("watchVotedToReshield DBRetrievePAppsByNetwork err:" + err.Error())
			continue
		}

		contract := papps.AppContracts["pdao"]

		gv, err := governance.NewGovernance(common.HexToAddress(contract), evmClient)
		if err != nil {
			continue
		}

		for _, vote := range proposals {

			proposalID, ok := big.NewInt(0).SetString(vote.ProposalID, 10)

			if !ok {
				go slacknoti.SendSlackNoti("watchReadyToVote parse  ProposalID no ok")
				continue
			}

			prop, err := gv.Proposals(nil, proposalID)
			if err != nil {
				log.Println("watchReadyToVote Proposals err:" + err.Error() + ", proposalID: " + vote.ProposalID + networkInfo.Endpoints[1])
				continue
			}

			header, err := evmClient.HeaderByNumber(context.Background(), nil)
			if err != nil {
				log.Println("watchReadyToVote HeaderByNumber err:", err)
				continue
			}

			log.Println("watchReadyToVote prop.StartBlock:", prop.StartBlock, "header.Number: ", header.Number)

			if prop.StartBlock.Cmp(header.Number) == -1 {

				voteJson, err := json.Marshal(vote)
				if err != nil {
					log.Println("marshal proposal err:", err)
					continue
				}

				_, err = SubmitPdaoOutchainTx(vote.SubmitBurnTx, wcommon.NETWORK_ETH, voteJson, false, pdao.VOTE_PROP, wcommon.ExternalTxTypePdaoVote)
				if err != nil {
					go slacknoti.SendSlackNoti("watchReadyToVote SubmitPdaoOutchainTx err:" + err.Error())
					continue
				}
				vote.Status = wcommon.StatusPdaOutchainTxSubmitting
				err = database.DBUpdateVoteTable(&vote)

				if err != nil {
					log.Println("watchReadyToVote DBUpdateVoteTable err:", err)
					continue
				}
			}

		}
	}
}

func watchVotedToReshield() {

	for {

		time.Sleep(20 * time.Second)

		votedRequests, err := database.DBGetVotingToReShield()
		if err != nil {
			log.Println("watchVotedToReshield DBGetVotingToReShield err:" + err.Error())
		}
		log.Println("watchVotedToReshield there are ", len(votedRequests), "records!")

		networkInfo, err := database.DBGetBridgeNetworkInfo(wcommon.NETWORK_ETH)
		if err != nil {
			log.Println("watchVotedToReshield DBGetBridgeNetworkInfo err:" + err.Error())
			continue
		}

		evmClient, err := ethclient.Dial(networkInfo.Endpoints[0])
		if err != nil {
			log.Println("watchVotedToReshield DBGetVoteSuccess err:" + err.Error())
			continue
		}

		papps, err := database.DBRetrievePAppsByNetwork("eth")
		if err != nil {
			log.Println("watchVotedToReshield DBRetrievePAppsByNetwork err:" + err.Error())
			continue
		}

		contract := papps.AppContracts["pdao"]

		gv, err := governance.NewGovernance(common.HexToAddress(contract), evmClient)
		if err != nil {
			log.Println("watchVotedToReshield NewGovernance err:" + err.Error())
			continue
		}

		for _, vote := range votedRequests {

			proposalID, ok := big.NewInt(0).SetString(vote.ProposalID, 10)

			if !ok {
				go slacknoti.SendSlackNoti("watchVotedToReshield parse ProposalID no ok")
				continue
			}

			state, err := gv.State(nil, proposalID)
			if err != nil {
				log.Println("watchVotedToReshield Votes err:" + err.Error() + ", proposalID: " + vote.ProposalID + networkInfo.Endpoints[1])
				continue
			}

			log.Println("watchVotedToReshield state prop:", state, "with prop id: ", proposalID)

			// if auto vote and state in Pending/Active then continue
			if state < 2 && vote.AutoVoted {
				continue
			}

			voteJson, err := json.Marshal(vote)
			if err != nil {
				log.Println("marshal proposal err:", err)
				continue
			}

			// todo: Call PRV contract reshield

			_, err = SubmitPdaoOutchainTx(vote.SubmitBurnTx, wcommon.NETWORK_ETH, voteJson, false, pdao.RESHIELD_BY_SIGN, wcommon.ExternalTxTypePdaoReShieldPRV)
			if err != nil {
				go slacknoti.SendSlackNoti("watchVotedToReshield SubmitPdaoOutchainTx err:" + err.Error())
				continue
			}

			vote.ReShieldStatus = wcommon.StatusPdaOutchainTxSubmitting
			err = database.DBUpdateVoteTable(&vote)

			if err != nil {
				log.Println("DBUpdateVoteTable err:", err)
				continue
			}
		}
	}
}

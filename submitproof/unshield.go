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
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/incognitochain/bridge-eth/bridge/vault"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/evmproof"
	"github.com/incognitochain/incognito-web-based-backend/slacknoti"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

func ProcessUnshieldTxRequest(ctx context.Context, m *pubsub.Message) {
	taskDesc := m.Attributes["task"]
	switch taskDesc {
	case UnshieldSubmitIncTask:
		processSubmitUnshieldRequest(ctx, m)
	case UnshieldSubmitExtTask:
		processSubmitUnshieldExtTask(ctx, m)
	case UnshieldSubmitFeeRefundTask:
		processSubmitRefundFeeTask(ctx, m)
	}
}

func processSubmitUnshieldRequest(ctx context.Context, m *pubsub.Message) {
	task := SubmitUnshieldTxTask{}
	err := json.Unmarshal(m.Data, &task)
	if err != nil {
		log.Println("processSubmitUnshieldRequest error decoding message", err)
		m.Ack()
		return
	}

	data := wcommon.UnshieldTxData{
		IncTx:            task.TxHash,
		IncTxData:        string(task.TxRawData),
		Status:           wcommon.StatusSubmitting,
		IsUnifiedToken:   task.IsUnifiedToken,
		FeeToken:         task.FeeToken,
		FeeAmount:        task.FeeAmount,
		PFeeAmount:       task.PFeeAmount,
		TokenID:          task.Token,
		UTokenID:         task.UToken,
		Amount:           task.BurntAmount,
		Networks:         task.Networks,
		FeeRefundOTA:     task.FeeRefundOTA,
		FeeRefundAddress: task.FeeRefundAddress,
		OutchainStatus:   wcommon.StatusWaiting,
		UserAgent:        task.UserAgent,
	}
	_, err = database.DBSaveUnshieldTxData(data)
	if err != nil {
		writeErr, ok := err.(mongo.WriteException)
		if !ok {
			log.Println("DBSaveUnshieldTxData err", err)
			m.Nack()
			return
		}
		if !writeErr.HasErrorCode(11000) {
			log.Println("DBSaveUnshieldTxData err", err)
			m.Nack()
			return
		}
	}

	txDetail, err := incClient.GetTxDetail(task.TxHash)
	if err != nil {
		log.Println("GetTxDetail err", err)
	} else {
		if txDetail.IsInMempool {
			err = database.DBUpdateUnshieldTxStatus(task.TxHash, wcommon.StatusPending, "")
			if err != nil {
				log.Println(err)
				m.Nack()
				return
			}
		}
		if txDetail.IsInBlock {
			err = database.DBUpdateUnshieldTxStatus(task.TxHash, wcommon.StatusExecuting, "")
			if err != nil {
				log.Println(err)
				m.Nack()
				return
			}
		}
		m.Ack()
		return
	}

	var errSubmit error

	if task.IsPRVTx {
		errSubmit = incClient.SendRawTx(task.TxRawData)
	} else {
		errSubmit = incClient.SendRawTokenTx(task.TxRawData)
	}

	if errSubmit != nil {
		err = database.DBUpdateUnshieldTxStatus(task.TxHash, wcommon.StatusSubmitFailed, errSubmit.Error())
		if err != nil {
			log.Println(err)
			m.Nack()
			return
		}
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[unshield]` submit unshield failed 😵 `%v`", task.TxHash))
		m.Ack()
		return
	} else {
		err = database.DBUpdateUnshieldTxStatus(task.TxHash, wcommon.StatusPending, "")
		if err != nil {
			log.Println(err)
			m.Nack()
			return
		}
	}

	m.Ack()
}

func processSubmitUnshieldExtTask(ctx context.Context, m *pubsub.Message) {
	//TODO
	task := SubmitProofOutChainTask{}
	err := json.Unmarshal(m.Data, &task)
	if err != nil {
		log.Println("processSubmitUnshieldExtTask error decoding message", err)
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

	_, err = database.DBGetExternalTxStatusByIncTx(task.IncTxhash, task.Network)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Println("DBGetExternalTxStatusByIncTx err", err)
			m.Nack()
			return
		}
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[unshield]` submitProofTx `%v` for network `%v`", task.IncTxhash, task.Network))
	} else {
		if task.IsRetry {
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[unshield]` retry submitProofTx `%v` for network `%v` 🫡", task.IncTxhash, task.Network))
		}
	}

	status, err := createOutChainUnshieldTx(task.Network, task.IncTxhash, task.IsUnifiedToken)
	if err != nil {
		log.Println("createOutChainUnshieldTx error", err)
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

	err = database.DBUpdateUnshieldTxSubmitOutchainStatus(task.IncTxhash, wcommon.StatusPending)
	if err != nil {
		log.Println("DBUpdateUnshieldTxSubmitOutchainStatus err", err)
		m.Ack()
		return
	}

	m.Ack()
}

func createOutChainUnshieldTx(network string, incTxHash string, isUnifiedToken bool) (*wcommon.ExternalTxStatus, error) {
	var result wcommon.ExternalTxStatus

	// networkID := wcommon.GetNetworkID(network)
	networkInfo, err := database.DBGetBridgeNetworkInfo(network)
	if err != nil {
		return nil, err
	}

	pappAddress, err := database.DBGetPappVaultData(network, wcommon.ExternalTxTypeSwap)
	if err != nil {
		return nil, err
	}

	networkChainId := networkInfo.ChainID

	networkChainIdInt := new(big.Int)
	networkChainIdInt.SetString(networkChainId, 10)

	var proof *evmproof.DecodedProof
	if isUnifiedToken {
		proof, err = evmproof.GetAndDecodeBurnProofUnifiedToken(config.FullnodeURL, incTxHash, 0)
	} else {
		switch network {
		case wcommon.NETWORK_ETH:
			proof, err = evmproof.GetAndDecodeBurnProofV2(config.FullnodeURL, incTxHash, "getburnprooffordeposittosc")
		case wcommon.NETWORK_BSC:
			proof, err = evmproof.GetAndDecodeBurnProofV2(config.FullnodeURL, incTxHash, "getburnpbscprooffordeposittosc")
		case wcommon.NETWORK_PLG:
			proof, err = evmproof.GetAndDecodeBurnProofV2(config.FullnodeURL, incTxHash, "getburnplgprooffordeposittosc")
		case wcommon.NETWORK_FTM:
			proof, err = evmproof.GetAndDecodeBurnProofV2(config.FullnodeURL, incTxHash, "getburnftmprooffordeposittosc")
		case wcommon.NETWORK_AVAX:
			proof, err = evmproof.GetAndDecodeBurnProofV2(config.FullnodeURL, incTxHash, "getburnavaxprooffordeposittosc")
		case wcommon.NETWORK_AURORA:
			proof, err = evmproof.GetAndDecodeBurnProofV2(config.FullnodeURL, incTxHash, "getburnauroraprooffordeposittosc")
		}
	}
	if err != nil {
		return nil, err
	}
	if proof == nil {
		return nil, fmt.Errorf("could not get proof for network %s", networkChainId)
	}

	if len(proof.InstRoots) == 0 {
		return nil, fmt.Errorf("could not get proof for network %s", networkChainId)
	}

	privKey, _ := crypto.HexToECDSA(config.EVMKey)
	i := 0
retry:
	if i == 10 {
		return nil, errors.New("submit tx outchain failed")
	}
	for _, endpoint := range networkInfo.Endpoints {
		evmClient, err := ethclient.Dial(endpoint)
		if err != nil {
			log.Println(err)
			continue
		}

		c, err := vault.NewVault(common.HexToAddress(pappAddress.ContractAddress), evmClient)
		if err != nil {
			log.Println(err)
			continue
		}

		gasPrice, err := evmClient.SuggestGasPrice(context.Background())
		if err != nil {
			log.Println(err)
			continue
		}

		auth, err := bind.NewKeyedTransactorWithChainID(privKey, networkChainIdInt)
		if err != nil {
			log.Println(err)
			continue
		}

		gasPrice = gasPrice.Mul(gasPrice, big.NewInt(11))
		gasPrice = gasPrice.Div(gasPrice, big.NewInt(10))

		auth.GasPrice = gasPrice
		if network == "eth" {
			auth.GasLimit = wcommon.EVMGasLimitETH
		} else {
			if network == "bsc" {
				auth.GasLimit = wcommon.EVMGasLimitPancake
			}
			auth.GasLimit = wcommon.EVMGasLimit
		}

		result.Type = wcommon.ExternalTxTypeUnshield
		result.Network = network
		result.IncRequestTx = incTxHash

		tx, err := evmproof.ExecuteWithBurnProof(c, auth, proof)
		if err != nil {
			log.Println(err)
			if strings.Contains(err.Error(), "insufficient funds") {
				return nil, errors.New("submit tx outchain failed err insufficient funds")
			}
			continue
		}
		result.Txhash = tx.Hash().String()
		result.Status = wcommon.StatusPending
		result.Nonce = tx.Nonce()
		break
	}

	if result.Txhash == "" {
		i++
		time.Sleep(2 * time.Second)
		goto retry
	}

	return &result, nil
}

package submitproof

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/incognitochain/bridge-eth/bridge/vault"
	inccommon "github.com/incognitochain/go-incognito-sdk-v2/common"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/evmproof"
	"github.com/incognitochain/incognito-web-based-backend/papps/popensea"
	"github.com/incognitochain/incognito-web-based-backend/slacknoti"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

func ProcessPappTxRequest(ctx context.Context, m *pubsub.Message) {
	taskDesc := m.Attributes["task"]
	switch taskDesc {
	case PappSubmitIncTask:
		processSubmitPappIncTask(ctx, m)
	case PappSubmitExtTask:
		processSubmitPappExtTask(ctx, m)
	case PappSubmitFeeRefundTask:
		processSubmitRefundFeeTask(ctx, m)
	case PdaoSubmitProposalExtTask:
		processSubmitPdaoRequest(ctx, m)
	case PdaoSubmitVoteExtTask:
		processSubmitVoteRequest(ctx, m)
	case PdaoSubmitReShieldPRVExtTask:
		processSubmitReShieldPRVRequest(ctx, m)
	}
}

func processSubmitPappExtTask(ctx context.Context, m *pubsub.Message) {
	task := SubmitProofOutChainTask{}
	err := json.Unmarshal(m.Data, &task)
	if err != nil {
		log.Println("processSubmitPappExtTask error decoding message", err)
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
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[swaptx]` submitProofTx timeout 😵 inctx `%v` network `%v`\n", task.IncTxhash, task.Network))
		return
	}

	_, err = database.DBGetExternalTxStatusByIncTx(task.IncTxhash, task.Network)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Println("DBGetExternalTxStatusByIncTx err", err)
			m.Nack()
			return
		}
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[swaptx]` submitProofTx `%v` for network `%v`", task.IncTxhash, task.Network))
	} else {
		if task.IsRetry {
			go slacknoti.SendSlackNoti(fmt.Sprintf("`[swaptx]` retry submitProofTx `%v` for network `%v` 🫡", task.IncTxhash, task.Network))
		}
	}

	status, err := createOutChainSwapTx(task.Network, task.IncTxhash, task.IsUnifiedToken, task.Type)
	if err != nil {
		log.Println("createOutChainSwapTx error", err)
		time.Sleep(15 * time.Second)
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[swaptx]` submitProofTx `%v` for network `%v` failed 😵 err: %v", task.IncTxhash, task.Network, err))
		m.Ack()
		return
	}
	go slacknoti.SendSlackNoti(fmt.Sprintf("`[swaptx]` submitProofTx `%v` for network `%v` success 👌 txhash `%v`", task.IncTxhash, task.Network, status.Txhash))

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

	err = database.DBUpdatePappTxSubmitOutchainStatus(task.IncTxhash, wcommon.StatusPending)
	if err != nil {
		log.Println("DBUpdatePappTxSubmitOutchainStatus err", err)
		m.Ack()
		return
	}

	m.Ack()
}

func processSubmitPappIncTask(ctx context.Context, m *pubsub.Message) {
	task := SubmitPappTxTask{}
	err := json.Unmarshal(m.Data, &task)
	if err != nil {
		log.Println("processSubmitPappIncTask error decoding message", err)
		m.Ack()
		return
	}
	pappSwapInfoStr, _ := json.MarshalIndent(task.PappSwapInfo, "", "\t")
	data := wcommon.PappTxData{
		IncTx:            task.TxHash,
		IncTxData:        string(task.TxRawData),
		Type:             task.TxType,
		Status:           wcommon.StatusSubmitting,
		IsUnifiedToken:   task.IsUnifiedToken,
		FeeToken:         task.FeeToken,
		FeeAmount:        task.FeeAmount,
		PFeeAmount:       task.PFeeAmount,
		BurntToken:       task.BurntToken,
		BurntAmount:      task.BurntAmount,
		PappSwapInfo:     string(pappSwapInfoStr),
		Networks:         task.Networks,
		FeeRefundOTA:     task.FeeRefundOTA,
		FeeRefundAddress: task.FeeRefundAddress,
		OutchainStatus:   wcommon.StatusWaiting,
		UserAgent:        task.UserAgent,
	}
	docID, err := database.DBSavePappTxData(data)
	if err != nil {
		writeErr, ok := err.(mongo.WriteException)
		if !ok {
			log.Println("DBAddPappTxData err", err)
			m.Nack()
			return
		}
		if !writeErr.HasErrorCode(11000) {
			log.Println("DBAddPappTxData err", err)
			m.Nack()
			return
		}
	}

	//TODO: popensea parse calldata
	switch task.TxType {
	case wcommon.ExternalTxTypeOpenseaOffer:
		offerStr := task.PappSwapInfo.AdditionalData
		networknInfo, err := database.DBGetBridgeNetworkInfo("eth")
		if err != nil {
			log.Println("DBGetBridgeNetworkInfo err", err)
			m.Nack()
			return
		}
		offer := popensea.OrderComponents{}
		offerBytes, err := hex.DecodeString(offerStr)
		if err != nil {
			log.Println(" hex.DecodeString(offerStr) err", err)
			m.Nack()
			return
		}

		err = json.Unmarshal(offerBytes, &offer)
		if err != nil {
			log.Println("json.Unmarshal(offerBytes, &offer) err", err)
			m.Nack()
			return
		}
		papps, err := database.DBRetrievePAppsByNetwork("eth")
		if err != nil {
			log.Println("json.Unmarshal(offerBytes, &offer) err", err)
			m.Nack()
			return
		}

		seaportAddress, exist := papps.AppContracts["seaport"]
		if !exist {
			log.Println("opensea seaport not found")
			m.Nack()
			return
		}

		offerHash := ""
		for _, endpoint := range networknInfo.Endpoints {
			evmClient, err := ethclient.Dial(endpoint)
			if err != nil {
				log.Println("ethclient.Dial err:", err)
				continue
			}

			seaport, err := popensea.NewIopensea(common.HexToAddress(seaportAddress), evmClient)
			if err != nil {
				log.Println("popensea.NewIopensea err:", err)
				continue
			}
			orderHash, err := seaport.GetOrderHash(nil, offer)
			if err != nil {
				log.Println("seaport.GetOrderHash err:", err)
				continue
			}
			offerHash = hex.EncodeToString(orderHash[:])
			break
		}

		timeoutAt := time.Unix(offer.EndTime.Int64(), 0)
		nftID := offer.Consideration[0].IdentifierOrCriteria.String()
		nftCollection := offer.Consideration[0].Token.Hex()
		receiver := offer.Consideration[0].Recipient.Hex()
		offerInfo := wcommon.OpenseaOfferData{
			Receiver:      receiver,
			Status:        popensea.OfferStatusSubmitting,
			NFTCollection: nftCollection,
			NFTID:         nftID,
			TimeoutAt:     timeoutAt,
			OfferTxInc:    task.TxHash,
			OfferHash:     offerHash,
		}
		err = database.DBInsertOpenseaOfferData(&offerInfo)
		if err != nil {
			log.Println("seaport.GetOrderHash err:", err)
			m.Nack()
			return
		}
	case wcommon.ExternalTxTypeOpenseaOfferCancel:

	}

	txDetail, err := incClient.GetTxDetail(task.TxHash)
	if err != nil {
		log.Println("GetTxDetail err", err)
	} else {
		if txDetail.IsInMempool {
			err = database.DBUpdatePappTxStatus(task.TxHash, wcommon.StatusPending, "")
			if err != nil {
				log.Println(err)
				m.Nack()
				return
			}
		}
		if txDetail.IsInBlock {
			err = database.DBUpdatePappTxStatus(task.TxHash, wcommon.StatusExecuting, "")
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
		err = database.DBUpdatePappTxStatus(task.TxHash, wcommon.StatusSubmitFailed, errSubmit.Error())
		if err != nil {
			log.Println(err)
			m.Nack()
			return
		}
		go slacknoti.SendSlackNoti(fmt.Sprintf("`[swaptx]` submit swaptx failed 😵 `%v`", task.TxHash))
		m.Ack()
		return
	} else {
		err = database.DBUpdatePappTxStatus(task.TxHash, wcommon.StatusPending, "")
		if err != nil {
			log.Println(err)
			m.Nack()
			return
		}
	}
	switch task.TxType {
	case wcommon.ExternalTxTypeSwap:
		go func() {
			slackep := os.Getenv("SLACK_SWAP_ALERT")
			if slackep != "" {
				swapAlert := ""
				pappTxData := data
				if pappTxData.PappSwapInfo != "" {
					networkID := wcommon.GetNetworkID(task.Networks[0])
					tkInInfo, _ := getTokenInfo(task.PappSwapInfo.TokenIn)
					amount := new(big.Float).SetInt(task.PappSwapInfo.TokenInAmount)
					decimal := new(big.Float)
					decimalInt, err := getTokenDecimalOnNetwork(tkInInfo, networkID)
					if err != nil {
						log.Println("getTokenDecimalOnNetwork1", err)
						return
					}
					decimal.SetFloat64(math.Pow10(int(-decimalInt)))

					amountInFloat := amount.Mul(amount, decimal).Text('f', -1)
					tokenInSymbol := tkInInfo.Symbol

					tkOutInfo, _ := getTokenInfo(task.PappSwapInfo.TokenOut)
					amount = new(big.Float).SetInt(task.PappSwapInfo.MinOutAmount)

					decimalInt, err = getTokenDecimalOnNetwork(tkOutInfo, networkID)
					if err != nil {
						log.Println("getTokenDecimalOnNetwork2", err)
						return
					}
					decimal.SetFloat64(math.Pow10(int(-decimalInt)))
					amountOutFloat := amount.Mul(amount, decimal).Text('f', -1)
					tokenOutSymbol := tkOutInfo.Symbol

					uaStr := parseUserAgent(task.UserAgent)

					swapAlert = fmt.Sprintf("`[%v(%v) | %v]` swap submitting 🛰\n SwapID: `%v`\n Requested: `%v %v` to `%v %v`\n--------------------------------------------------------", task.PappSwapInfo.DappName, pappTxData.Networks[0], uaStr, docID.Hex(), amountInFloat, tokenInSymbol, amountOutFloat, tokenOutSymbol)
					log.Println(swapAlert)
					slacknoti.SendWithCustomChannel(swapAlert, slackep)
				}
			}
		}()
	case wcommon.ExternalTxTypeOpenseaBuy:
	case wcommon.ExternalTxTypeOpenseaOffer:
	case wcommon.ExternalTxTypeOpenseaOfferCancel:
	}

	m.Ack()
}

func createOutChainSwapTx(network string, incTxHash string, isUnifiedToken bool, txType int) (*wcommon.ExternalTxStatus, error) {
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
	// if isUnifiedToken {
	proof, err = evmproof.GetAndDecodeBurnProofUnifiedToken(config.FullnodeURL, incTxHash, 0)
	// } else {
	// 	proof, err = evmproof.GetAndDecodeBurnProofV2(config.FullnodeURL, incTxHash, "getburnplgprooffordeposittosc")
	// }
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

		gasPrice = gasPrice.Mul(gasPrice, big.NewInt(12))
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

		result.Type = txType
		result.Network = network
		result.IncRequestTx = incTxHash

		// address, err := wcommon.GetEVMAddress(config.EVMKey)
		// if err != nil {
		// 	log.Println(err)
		// 	continue
		// }
		// account := common.HexToAddress(address)
		// pendingNonce, _ := evmClient.PendingNonceAt(context.Background(), account)
		// auth.Nonce = new(big.Int).SetUint64(pendingNonce)

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

func processSubmitRefundFeeTask(ctx context.Context, m *pubsub.Message) {
	task := SubmitRefundFeeTask{}
	err := json.Unmarshal(m.Data, &task)
	if err != nil {
		log.Println("processSubmitRefundFeeTask error decoding message", err)
		m.Ack()
		return
	}
	i := 0
	defer m.Ack()
	go slacknoti.SendSlackNoti(fmt.Sprintf("`[refundfee]` Need refund fee for tx `%v`\n", task.IncReqTx))
retry:
	i++
	var errSubmit error
	var txhash string
	var txRaw []byte
	if i == 10 {
		errStr := ""
		if errSubmit != nil {
			errStr = errSubmit.Error()
		}
		err = database.DBUpdateRefundFeeRefundTx(txhash, task.IncReqTx, wcommon.StatusSubmitFailed, errStr)
		if err != nil {
			log.Println("DBUpdateRefundFeeRefundTx error ", err)
			return
		}
	}

	if task.Token != inccommon.PRVCoinID.String() {
		var tokenParam *incclient.TxTokenParam
		if task.PaymentAddress != "" {
			tokenParam = incclient.NewTxTokenParam(task.Token, 1, []string{task.PaymentAddress}, []uint64{task.Amount}, false, 0, nil)
		} else {
			tokenParam = incclient.NewTxTokenParam(task.Token, 1, []string{task.OTA}, []uint64{task.Amount}, false, 0, nil)
		}

		txParam := incclient.NewTxParam(config.IncKey, []string{}, []uint64{}, 100, tokenParam, nil, nil)

		txRaw, txhash, err = incClient.CreateRawTokenTransactionVer2(txParam)
		if err != nil {
			log.Println("CreateRawTokenTransactionVer2 error ", err)
			errSubmit = err
			goto retry
		}
		err = incClient.SendRawTokenTx(txRaw)
		if err != nil {
			log.Println("SendRawTokenTx error ", err)
			errSubmit = err
			goto retry
		}
	} else {
		var txParam *incclient.TxParam

		if task.PaymentAddress != "" {
			txParam = incclient.NewTxParam(config.IncKey, []string{task.PaymentAddress}, []uint64{task.Amount}, 0, nil, nil, nil)
		} else {
			txParam = incclient.NewTxParam(config.IncKey, []string{task.OTA}, []uint64{task.Amount}, 0, nil, nil, nil)
		}

		txRaw, txhash, err = incClient.CreateRawTransactionVer2(txParam)
		if err != nil {
			log.Println("CreateRawTransactionVer2 error ", err)
			errSubmit = err
			goto retry
		}
		err = incClient.SendRawTx(txRaw)
		if err != nil {
			log.Println("SendRawTx error ", err)
			errSubmit = err
			goto retry
		}
	}

	if errSubmit != nil {
		log.Println("processSubmitRefundFeeTask error ", errSubmit)
		time.Sleep(5 * time.Second)
		goto retry
	} else {
	retrySaved:
		err = database.DBUpdateRefundFeeRefundTx(txhash, task.IncReqTx, wcommon.StatusPending, "")
		if err != nil {
			time.Sleep(5 * time.Second)
			goto retrySaved
		}
	}
	go slacknoti.SendSlackNoti(fmt.Sprintf("`[refundfee]` refund fee tx submitted, `%v`, requestTx `%v`, isPrivacyFee `%v`\n", txhash, task.IncReqTx, task.IsPrivacyFeeRefund))

}
func speedupOutChainSwapTx(network int, evmTxHash string) error {
	return nil
}

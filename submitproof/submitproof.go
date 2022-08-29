package submitproof

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	"github.com/incognitochain/go-incognito-sdk-v2/rpchandler"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
)

var config wcommon.Config
var incClient *incclient.IncClient
var keyList []string

func getProof(txhash string, networkID int) (*wcommon.EVMProofRecordData, *incclient.EVMDepositProof, error) {
	networkName := wcommon.GetNetworkName(networkID)
	networkInfo, err := database.DBGetBridgeNetworkInfo(networkName)
	if err != nil {
		return nil, nil, err
	}

	for _, endpoint := range networkInfo.Endpoints {
		evmClient, err := ethclient.Dial(endpoint)
		if err != nil {
			log.Println(err)
			continue
		}
		blockNumber, blockHash, txIdx, proof, contractID, paymentAddr, isRedeposit, otaStr, amount, _, err := getETHDepositProof(evmClient, txhash)
		if err != nil {
			log.Println(err)
			continue
		}
		if len(proof) == 0 {
			return nil, nil, fmt.Errorf("invalid proof or tx not found")
		}
		depositProof := incclient.NewETHDepositProof(uint(blockNumber.Int64()), common.HexToHash(blockHash), txIdx, proof)

		proofBytes, _ := json.Marshal(proof)
		if len(proof) == 0 {
			return nil, nil, fmt.Errorf("invalid proof or tx not found")
		}
		result := wcommon.EVMProofRecordData{
			Proof:       string(proofBytes),
			BlockNumber: blockNumber.Uint64(),
			BlockHash:   blockHash,
			TxIndex:     txIdx,
			ContractID:  contractID,
			PaymentAddr: paymentAddr,
			IsRedeposit: isRedeposit,
			OTAStr:      otaStr,
			Amount:      amount,
			Network:     networkName,
		}

		return &result, depositProof, nil
	}

	return nil, nil, errors.New("cant retrieve proof")
}

func submitProof(txhash, tokenID string, networkID int, key string) (string, string, string, string, error) {
	var linkedTokenID string
	if tokenID != "" {
		linkedTokenID = getLinkedTokenID(tokenID, networkID)
		fmt.Println("used tokenID: ", linkedTokenID)
	}
	i := 0
	var finalErr string
retry:
	if i == 10 {
		return "", "", "", "", errors.New(finalErr)
	}
	if i > 0 {
		time.Sleep(1 * time.Second)
	}
	i++
	proofRecord, depositProof, err := getProof(txhash, networkID)
	if err != nil {
		log.Println("error:", err)
		finalErr = fmt.Sprintf("getProof %v %v ", txhash, networkID) + err.Error()
		goto retry
	}
	isSubmitted, err := checkProofSubmitted(proofRecord.BlockHash, proofRecord.TxIndex, networkID)
	if err != nil {
		log.Println("checkProofSubmitted error:", err)
	}
	if isSubmitted {
		return "", "", "", "", errors.New(ProofAlreadySubmitError)
	}
	if linkedTokenID == "" && tokenID == "" {
		tokenID, linkedTokenID, err = findTokenByContractID(proofRecord.ContractID, networkID)
		if err != nil {
			log.Println("error:", err)
			goto retry
		}
	}
	incTx, err := submitProofTx(depositProof, linkedTokenID, tokenID, networkID, key)
	if err != nil {
		log.Println("error:", err)
		fmt.Println("linkedTokenID, tokenID", linkedTokenID, tokenID)
		fmt.Println("incTx", incTx)
		finalErr = "submitProof " + err.Error()
		goto retry
	}
	fmt.Println("done submit proof", incTx)

	if proofRecord.IsRedeposit {
		return incTx, proofRecord.OTAStr, tokenID, linkedTokenID, nil
	}
	return incTx, proofRecord.PaymentAddr, tokenID, linkedTokenID, nil
}

func submitProofTx(proof *incclient.EVMDepositProof, tokenID string, pUTokenID string, networkID int, key string) (string, error) {
	if tokenID == pUTokenID {
		result, err := incClient.CreateAndSendIssuingEVMRequestTransaction(key, tokenID, *proof, networkID-1)
		if err != nil {
			return result, err
		}
		return result, nil
	}

	result, err := incClient.CreateAndSendIssuingpUnifiedRequestTransaction(key, tokenID, pUTokenID, *proof, networkID)
	if err != nil {
		return result, err
	}
	return result, err
}

func checkProofSubmitted(blockHash string, txIdx uint, networkID int) (bool, error) {
	var result bool
	method := "checkethhashissued"
	switch networkID {
	case 1:
		method = "checkethhashissued"
	case 2:
		method = "checkbschashissued"
	case 3, 4:
		return false, nil
	}
	mapParams := make(map[string]interface{})
	mapParams["BlockHash"] = blockHash
	mapParams["TxIndex"] = txIdx
	responseInBytes, err := incClient.NewRPCCall("1.0", method, []interface{}{mapParams}, 1)
	if err != nil {
		return false, err
	}
	err = rpchandler.ParseResponse(responseInBytes, &result)
	if err != nil {
		return false, err
	}
	return result, nil
}

func checkBeaconBridgeAggUnshieldStatus(txhash string) (int, error) {
	var result struct {
		Status int
	}
	method := "bridgeaggGetStatusUnshield"
	responseInBytes, err := incClient.NewRPCCall("1.0", method, []interface{}{txhash}, 1)
	if err != nil {
		return -1, err
	}
	err = rpchandler.ParseResponse(responseInBytes, &result)
	if err != nil {
		return -1, err
	}
	return result.Status, nil
}

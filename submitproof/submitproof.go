package submitproof

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	"github.com/incognitochain/go-incognito-sdk-v2/rpchandler"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
)

var config wcommon.Config
var incClient *incclient.IncClient
var keyList []string

func getProof(txhash string, networkID int) (*incclient.EVMDepositProof, string, string, string, uint, error) {
	blockNumber, blockHash, txIdx, proof, contractID, paymentAddr, err := getETHDepositProof(incClient, networkID, txhash)
	if err != nil {
		return nil, "", "", blockHash, txIdx, err
	}
	if len(proof) == 0 {
		return nil, "", "", blockHash, txIdx, fmt.Errorf("invalid proof or tx not found")
	}
	result := incclient.NewETHDepositProof(uint(blockNumber.Int64()), common.HexToHash(blockHash), txIdx, proof)
	return result, contractID, paymentAddr, blockHash, txIdx, nil
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
	proof, contractID, paymentAddr, blockHash, txIdx, err := getProof(txhash, networkID-1)
	if err != nil {
		log.Println("error:", err)
		finalErr = "getProof " + err.Error()
		goto retry
	}
	isSubmitted, err := checkProofSubmitted(blockHash, txIdx, networkID)
	if err != nil {
		log.Println("checkProofSubmitted error:", err)
	}
	if isSubmitted {
		return "", "", "", "", errors.New(ProofAlreadySubmitError)
	}
	if linkedTokenID == "" && tokenID == "" {
		tokenID, linkedTokenID, err = findTokenByContractID(contractID, networkID)
		if err != nil {
			log.Println("error:", err)
			goto retry
		}
	}
	incTx, err := submitProofTx(proof, linkedTokenID, tokenID, networkID, key)
	if err != nil {
		log.Println("error:", err)
		fmt.Println("linkedTokenID, tokenID", linkedTokenID, tokenID)
		fmt.Println("incTx", incTx)
		finalErr = "submitProof " + err.Error()
		goto retry
	}
	fmt.Println("done submit proof", incTx)

	return incTx, paymentAddr, tokenID, linkedTokenID, nil
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

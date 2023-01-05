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
	_, blockHash, txIdx, proof, contractID, paymentAddr, err := getETHDepositProof(incClient, networkID, txhash)
	if err != nil {
		return nil, "", "", blockHash, txIdx, err
	}
	if len(proof) == 0 {
		return nil, "", "", blockHash, txIdx, fmt.Errorf("invalid proof or tx not found")
	}
	result := incclient.NewETHDepositProof(0, common.HexToHash(blockHash), txIdx, proof)
	return result, contractID, paymentAddr, blockHash, txIdx, nil
}

func submitProof(txhash, tokenID string, networkID int, key string) (string, string, error) {
	err := updateShieldTxStatus(txhash, networkID, ShieldStatusSubmitting)
	if err != nil {
		log.Println("error:", err)
		return "", "", err
	}
	var linkedTokenID string
	if tokenID != "" {
		linkedTokenID = getLinkedTokenID(tokenID, networkID)
		fmt.Println("used tokenID: ", linkedTokenID)
	}
	i := 0
	var finalErr string
retry:
	if i == 10 {
		err = updateShieldTxStatus(txhash, networkID, ShieldStatusSubmitting)
		if err != nil {
			log.Println("updateShieldTxStatus error:", err)
			return "", "", err
		}
		err = setShieldTxStatusError(txhash, networkID, finalErr)
		if err != nil {
			log.Println("setShieldTxStatusError error:", err)
			return "", "", err
		}
		return "", "", errors.New(finalErr)
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
		err = updateShieldTxStatus(txhash, networkID, ShieldStatusAccepted)
		if err != nil {
			log.Println("error123:", err)
			return "", "", err
		}
		return "", "", errors.New(ProofAlreadySubmitError)
	}
	if linkedTokenID == "" && tokenID == "" {
		tokenID, linkedTokenID, err = findTokenByContractID(contractID, networkID)
		if err != nil {
			log.Println("error:", err)
			goto retry
		}
	}
	result, err := submitProofTx(proof, linkedTokenID, tokenID, networkID, key)
	if err != nil {
		log.Println("error:", err)
		finalErr = "submitProof " + err.Error()
		goto retry
	}
	fmt.Println("done submit proof", result)
	err = updateShieldTxStatus(txhash, networkID, ShieldStatusPending)
	if err != nil {
		log.Println("error123:", err)
		return "", "", err
	}
	return result, paymentAddr, nil
}

func submitProofTx(proof *incclient.EVMDepositProof, tokenID string, pUTokenID string, networkID int, key string) (string, error) {
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

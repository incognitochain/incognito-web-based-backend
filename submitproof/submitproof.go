package submitproof

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
)

var config wcommon.Config
var incClient *incclient.IncClient
var keyList []string

func getProof(txhash string, networkID int) (*incclient.EVMDepositProof, string, error) {
	_, blockHash, txIdx, proof, contractID, err := getETHDepositProof(incClient, networkID, txhash)
	if err != nil {
		return nil, "", err
	}
	if len(proof) == 0 {
		return nil, "", fmt.Errorf("invalid proof or tx not found")
	}
	result := incclient.NewETHDepositProof(0, common.HexToHash(blockHash), txIdx, proof)
	return result, contractID, nil
}

func submitProof(txhash, tokenID string, networkID int, key string) (string, error) {
	err := updateShieldTxStatus(txhash, networkID, ShieldStatusSubmitting)
	if err != nil {
		log.Println("error:", err)
		return "", err
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
		err = updateShieldTxStatus(txhash, networkID, ShieldStatusSubmitFailed)
		if err != nil {
			log.Println("updateShieldTxStatus error:", err)
			return "", err
		}
		err = setShieldTxStatusError(txhash, networkID, finalErr)
		if err != nil {
			log.Println("setShieldTxStatusError error:", err)
			return "", err
		}
		return "", errors.New(finalErr)
	}
	if i > 0 {
		time.Sleep(1 * time.Second)
	}
	i++
	proof, contractID, err := getProof(txhash, networkID-1)
	if err != nil {
		log.Println("error:", err)
		finalErr = "getProof " + err.Error()
		goto retry
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
	// _ = proof
	// result := "sdfgsdfds"
	fmt.Println("done submit proof", result)
	err = updateShieldTxStatus(txhash, networkID, ShieldStatusPending)
	if err != nil {
		log.Println("error123:", err)
		return "", err
	}
	return result, nil
}

func submitProofTx(proof *incclient.EVMDepositProof, tokenID string, pUTokenID string, networkID int, key string) (string, error) {
	result, err := incClient.CreateAndSendIssuingpUnifiedRequestTransaction(key, tokenID, pUTokenID, *proof, networkID)
	if err != nil {
		return result, err
	}
	return result, err
}

func checkProofSubmitted(proof string) (bool, error) {
	var result bool

	return result, nil
}

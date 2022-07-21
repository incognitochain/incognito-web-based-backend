package submitproof

import (
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	"github.com/incognitochain/go-incognito-sdk-v2/wallet"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/redb"
	"github.com/pkg/errors"
	"github.com/rueian/rueidis"
)

var config wcommon.Config
var incClient *incclient.IncClient
var keyList []string

var db rueidis.Client

func connectDB(endpoint []string) error {
	var err error
	fmt.Println("endpoint: ", endpoint)
	db, err = redb.NewClient(endpoint)
	return err
}

func Start(keylist []string, cfg wcommon.Config) error {
	config = cfg
	keyList = keylist

	err := connectDB(cfg.DatabaseURLs)
	if err != nil {
		return err
	}

	network := cfg.NetworkID
	switch network {
	case "mainnet":
		incClient, err = incclient.NewMainNetClient()
	case "testnet": // testnet2
		incClient, err = incclient.NewTestNetClient()
	case "testnet1":
		incClient, err = incclient.NewTestNet1Client()
	case "devnet":
		return errors.New("unsupported network")
	}
	if err != nil {
		return err
	}

	for _, v := range keyList {
		wl, err := wallet.Base58CheckDeserialize(v)
		if err != nil {
			panic(err)
		}
		err = incClient.SubmitKey(wl.Base58CheckSerialize(wallet.OTAKeyType))
		if err != nil {
			return err
		}
	}
	incclient.Logger = incclient.NewLogger(true)
	log.Println("Done submit keys")

	return nil
}

func SubmitShieldProof(txhash string, networkID int, tokenID string) (interface{}, error) {
	if networkID == 0 {
		return "", errors.New("unsported network")
	}

	currentStatus, err := getShieldTxStatus(txhash, networkID, tokenID)
	if err != nil {
		return "", err
	}
	if currentStatus != ShieldStatusUnknown {
		return ShieldStatusMap[currentStatus], nil
	}
	go submitProof(txhash, tokenID, networkID)
	return "submitting", nil
}

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

func submitProof(txhash, tokenID string, networkID int) {
	err := updateShieldTxStatus(txhash, networkID, tokenID, ShieldStatusSubmitting)
	if err != nil {
		log.Println("error:", err)
	}
	var linkedTokenID string
	if tokenID != "" {
		linkedTokenID = getLinkedTokenID(tokenID, networkID)
		fmt.Println("used tokenID: ", linkedTokenID)
	}
	i := 0
	var finalErr string
retry:
	if i == 120 {
		err = updateShieldTxStatus(txhash, networkID, tokenID, ShieldStatusSubmitFailed)
		if err != nil {
			log.Println("updateShieldTxStatus error:", err)
		}
		err = setShieldTxStatusError(txhash, networkID, tokenID, finalErr)
		if err != nil {
			log.Println("setShieldTxStatusError error:", err)
		}
		panic(fmt.Sprintln("failed to shield txhash:", txhash))
	}
	if i > 0 {
		time.Sleep(15 * time.Second)
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
	// result, err := submitProofTx(proof, linkedTokenID, tokenID, networkID)
	// if err != nil {
	// 	log.Println("error:", err)
	// 	finalErr = "submitProof " + err.Error()
	// 	goto retry
	// }
	_ = proof
	fmt.Println("done submit proof")
	err = updateShieldTxStatus(txhash, networkID, tokenID, ShieldStatusSubmitted)
	if err != nil {
		log.Println("error123:", err)
	}
	// log.Println(result)
}

func submitProofTx(proof *incclient.EVMDepositProof, tokenID string, pUTokenID string, networkID int) (string, error) {
	t := time.Now().Unix()
	key := keyList[t%int64(len(keyList))]
	result, err := incClient.CreateAndSendIssuingpUnifiedRequestTransaction(key, tokenID, pUTokenID, *proof, networkID)
	if err != nil {
		return result, err
	}
	return result, err
}

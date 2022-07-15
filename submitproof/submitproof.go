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
	db, err = redb.NewClient(endpoint)
	return err
}

func Start(keylist []string, network string, cfg wcommon.Config) error {
	config = cfg
	keyList = keylist

	err := connectDB(cfg.DatabaseURLs)
	if err != nil {
		return err
	}

	switch network {
	case "mainnet":
		incClient, err = incclient.NewMainNetClient()
	case "testnet": // testnet2
		incClient, err = incclient.NewTestNetClient()
	case "testnet1":
		incClient, err = incclient.NewTestNet1Client()
	case "devnet":
		// incclient.NewIncClient()
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

	}
	go submitProof(txhash, tokenID, networkID)
	return "submitting", nil
}

func getProof(txhash string, networkID int) (*incclient.EVMDepositProof, error) {
	_, blockHash, txIdx, proof, err := getETHDepositProof(incClient, networkID, txhash)
	if err != nil {
		return nil, err
	}
	if len(proof) == 0 {
		return nil, fmt.Errorf("invalid proof or tx not found")
	}
	result := incclient.NewETHDepositProof(0, common.HexToHash(blockHash), txIdx, proof)
	return result, nil
}

func submitProof(txhash, tokenID string, networkID int) {
	err := updateShieldTxStatus(txhash, networkID, tokenID, ShieldStatusSubmitting)
	if err != nil {
		log.Println("error:", err)
	}
	linkedTokenID := getLinkedTokenID(tokenID, networkID)
	fmt.Println("used tokenID: ", linkedTokenID)
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

	time.Sleep(15 * time.Second)
	i++
	proof, err := getProof(txhash, networkID-1)
	if err != nil {
		log.Println("error:", err)
		finalErr = "getProof " + err.Error()
		goto retry
	}
	result, err := submitProofTx(proof, linkedTokenID, tokenID, networkID)
	if err != nil {
		log.Println("error:", err)
		finalErr = "submitProof " + err.Error()
		goto retry
	}
	fmt.Println("done submit proof")
	err = updateShieldTxStatus(txhash, networkID, tokenID, ShieldStatusSubmitted)
	if err != nil {
		log.Println("error:", err)
	}
	log.Println(result)
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

func getTokenInfo(pUTokenID string) (*TokenInfo, error) {

	type APIRespond struct {
		Result []TokenInfo
		Error  *string
	}

	reqBody := struct {
		TokenIDs []string
	}{
		TokenIDs: []string{pUTokenID},
	}

	var responseBodyData APIRespond
	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).SetBody(reqBody).
		Post(config.CoinserviceURL + "/coins/tokeninfo")
	if err != nil {
		return nil, err
	}

	if len(responseBodyData.Result) == 1 {
		return &responseBodyData.Result[0], nil
	}
	return nil, errors.New(fmt.Sprintf("token not found"))
}

func getLinkedTokenID(pUTokenID string, networkID int) string {
	tokenInfo, err := getTokenInfo(pUTokenID)
	if err != nil {
		log.Println("getLinkedTokenID", err)
		return pUTokenID
	}
	for _, v := range tokenInfo.ListUnifiedToken {
		if v.NetworkID == networkID {
			return v.TokenID
		}
	}
	return pUTokenID
}

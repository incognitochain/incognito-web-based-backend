package submitproof

import (
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	"github.com/incognitochain/go-incognito-sdk-v2/wallet"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/pkg/errors"
)

var config wcommon.Config
var incClient *incclient.IncClient
var keyList []string

func Start(keylist []string, cfg wcommon.Config) error {
	config = cfg
	keyList = keylist
	network := cfg.NetworkID
	var err error
	switch network {
	case "mainnet":
		incClient, err = incclient.NewMainNetClient()
	case "testnet-2": // testnet2
		incClient, err = incclient.NewTestNetClient()
	case "testnet-1":
		incClient, err = incclient.NewTestNet1Client()
	case "devnet":
		return errors.New("unsupported network")
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

func SubmitShieldProof(txhash string, networkID int, tokenID string) error {
	if networkID == 0 {
		return errors.New("unsported network")
	}
	go func() {
		var linkedTokenID string
		if tokenID != "" {
			linkedTokenID = getLinkedTokenID(tokenID, networkID)
			fmt.Println("used tokenID: ", linkedTokenID)
		}

		i := 0
	retry:
		if i == 120 {
			panic(fmt.Sprintln("failed to shield txhash:", txhash))
		}
		if i > 0 {
			time.Sleep(15 * time.Second)
		}
		i++
		proof, contractID, err := getProof(txhash, networkID-1)
		if err != nil {
			log.Println("error:", err)
			goto retry
		}
		if linkedTokenID == "" && tokenID == "" {
			tokenID, linkedTokenID, err = findTokenByContractID(contractID, networkID)
			if err != nil {
				log.Println("error:", err)
				goto retry
			}
		}
		result, err := submitProof(proof, linkedTokenID, tokenID, networkID)
		if err != nil {
			log.Println("error:", err)
			goto retry
		}
		fmt.Println("done submit proof")
		log.Println(result)
	}()
	return nil
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

func submitProof(proof *incclient.EVMDepositProof, tokenID string, pUTokenID string, networkID int) (string, error) {
	t := time.Now().Unix()
	key := keyList[t%int64(len(keyList))]
	result, err := incClient.CreateAndSendIssuingpUnifiedRequestTransaction(key, tokenID, pUTokenID, *proof, networkID)
	if err != nil {
		return result, err
	}
	return result, err
}

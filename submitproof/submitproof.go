package submitproof

import (
	"errors"
	"fmt"
	"time"

	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	"github.com/incognitochain/go-incognito-sdk-v2/wallet"
)

var incClient *incclient.IncClient
var keyList []string

func Start(keylist []string, network string) error {
	keyList = keylist
	var err error
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

	fmt.Println("Done submit keys")

	return nil
}

func SubmitShieldProof(txhash string, networkID int, tokenID string) error {
	// networkID := 0
	// switch network {
	// case "eth":
	// 	networkID = rpc.ETHNetworkID
	// case "bsc":
	// 	networkID = rpc.BSCNetworkID
	// case "plg":
	// 	networkID = rpc.PLGNetworkID
	// case "ftm":
	// 	networkID = rpc.FTMNetworkID
	// }
	status, err := incClient.GetEVMTransactionStatus(txhash, networkID)
	if err != nil {
		return err
	}
	if status != 1 {
		return errors.New("tx failed")
	}
	proof, err := getProof(txhash, networkID)
	if err != nil {
		return err
	}
	submitProof(proof, tokenID, networkID)
	return nil
}

func getProof(txhash string, networkID int) (*incclient.EVMDepositProof, error) {
	proof, _, err := incClient.GetEVMDepositProof(txhash, networkID)
	if err != nil {
		return nil, err
	}
	return proof, nil
}

func submitProof(proof *incclient.EVMDepositProof, tokenID string, networkID int) (string, error) {
	t := time.Now().Unix()
	result, err := incClient.CreateAndSendIssuingEVMRequestTransaction(keyList[t%int64(len(keyList))], tokenID, *proof, networkID)
	if err != nil {
		return result, err
	}
	return result, err
}

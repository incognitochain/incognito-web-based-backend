package utxomanager

import (
	"fmt"

	"github.com/incognitochain/go-incognito-sdk-v2/common"
	"github.com/incognitochain/go-incognito-sdk-v2/common/base58"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	"github.com/incognitochain/go-incognito-sdk-v2/wallet"
)

func getListUTXOs(incClient *incclient.IncClient, privateKey string) ([]UTXO, error) {
	inputCoins, idxCoins, err := incClient.GetUnspentOutputCoins(privateKey, PRVIDStr, 0)

	if err != nil {
		return []UTXO{}, err
	}

	utxos := []UTXO{}
	for idx := range inputCoins {
		utxos = append(utxos, UTXO{
			Coin:  inputCoins[idx],
			Index: idxCoins[idx],
		})
	}
	return utxos, nil
}

func getPublicKeyStr(privateKey string) (string, error) {
	keyWallet, err := wallet.Base58CheckDeserialize(privateKey)
	if err != nil {
		return "", fmt.Errorf("Can not deserialize private key %v\n", err)
	}
	err = keyWallet.KeySet.InitFromPrivateKey(&keyWallet.KeySet.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("sender private key is invalid")
	}
	publicKeyBytes := keyWallet.KeySet.PaymentAddress.Pk
	publicKey := base58.Base58Check{}.Encode(publicKeyBytes, common.ZeroByte)
	return publicKey, nil
}

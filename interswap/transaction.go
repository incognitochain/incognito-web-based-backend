package interswap

import (
	"fmt"
	"log"

	"github.com/incognitochain/go-incognito-sdk-v2/coin"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
)

func createTxTokenWithInputCoins(
	senderPrivKey, otaReceiver, tokenID string, amount uint64,
	tokenUtxos []coin.PlainCoin, tokenUtxoIndices []uint64,
) (string, error) {
	prvUTXOs, _, err := UtxoManager.GetUTXOsByAmount(senderPrivKey, incclient.DefaultPRVFee)
	if err != nil {
		log.Printf("Error get PRV coin: %v\n", err)
		return "", fmt.Errorf("Error get PRV coin: %v\n", err)
	}

	prvCoins := []coin.PlainCoin{}
	prvCoinIndices := []uint64{}
	for _, u := range prvUTXOs {
		prvCoins = append(prvCoins, u.Coin)
		prvCoinIndices = append(prvCoinIndices, u.Index.Uint64())
	}

	txTokenParams := incclient.NewTxTokenParam(tokenID, 1, []string{otaReceiver}, []uint64{amount}, false, 0, nil)
	txParams := incclient.NewTxParam(senderPrivKey, nil, nil, incclient.DefaultPRVFee, txTokenParams, nil, nil)
	rawTx, txID, err := incClient.CreateRawTokenTransactionWithInputCoins(txParams, tokenUtxos, tokenUtxoIndices, prvCoins, prvCoinIndices)
	if err != nil {
		log.Printf("Error create tx: %v\n", err)
		return "", fmt.Errorf("Error create tx: %v\n", err)
	}
	err = incClient.SendRawTokenTx(rawTx)
	if err != nil {
		log.Printf("Error send tx: %v\n", err)
		return "", fmt.Errorf("Error send tx: %v\n", err)
	}

	return txID, nil
}

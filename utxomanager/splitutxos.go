package utxomanager

import (
	"fmt"
	"sync"
	"time"

	"github.com/incognitochain/go-incognito-sdk-v2/coin"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
)

const (
	MaxLoopTime   = 100
	MaxReceiver   = 2
	MinUTXOAmount = 1e9 // 1 PRV
	PRVIDStr      = "0000000000000000000000000000000000000000000000000000000000000004"
)

func SplitUTXOs(privateKey string, paymentAddress string, minNumUTXOs int, utxoManager *UTXOManager) error {
	cntLoop := 0

	for {
		utxos, err := utxoManager.GetListUnspentUTXO(privateKey)
		if err != nil {
			return err
		}

		fmt.Printf("Number of UTXOs: %v\n", len(utxos))

		if len(utxos) >= minNumUTXOs {
			fmt.Printf("Split UTXOs succeed.\n")
			break
		}
		// if len(utxos) == 0 {
		// 	return fmt.Errorf("Could not get any UTXO from this account")
		// }

		var wg sync.WaitGroup
		for idx := range utxos {
			utxo := utxos[idx]
			fmt.Printf("UTXO Value: %v\n", utxo.Coin.GetValue())
			if utxo.Coin.GetValue() < MinUTXOAmount*MaxReceiver {
				fmt.Printf("Skipped\n")
				continue
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				receiverList := []string{paymentAddress}
				amountList := []uint64{utxo.Coin.GetValue() / 2}

				txParam := incclient.NewTxParam(privateKey, receiverList, amountList, 0, nil, nil, nil)

				encodedTx, txID, err := utxoManager.IncClient.CreateRawTransactionWithInputCoins(
					txParam, []coin.PlainCoin{utxo.Coin}, []uint64{utxo.Index.Uint64()},
				)
				if err != nil {
					fmt.Printf("CreateRawTransactionWithInputCoins error: %v\n", err)
					return
				}
				err = utxoManager.IncClient.SendRawTx(encodedTx)
				if err != nil {
					fmt.Printf("SendRawTx error: %v\n", err)
					return
				}
				utxoManager.CacheUTXOsByTxID(privateKey, txID, []UTXO{utxo})
				fmt.Printf("TxID: %+v\n", txID)
			}()
		}
		wg.Wait()

		cntLoop++
		if cntLoop >= MaxLoopTime {
			break
		}

		time.Sleep(15 * time.Second)
	}
	return nil
}

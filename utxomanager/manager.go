package utxomanager

import (
	"math/big"
	"sync"

	"github.com/incognitochain/go-incognito-sdk-v2/coin"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
)

type UTXO struct {
	Coin  coin.PlainCoin
	Index *big.Int
}

type UTXOManager struct {
	Unspent   map[string][]UTXO            // public key: UTXO
	Caches    map[string]map[string][]UTXO // public key: txID: UTXO
	mux       sync.Mutex
	TmpIdx    int
	IncClient *incclient.IncClient
}

func NewUTXOManager(incClient *incclient.IncClient) *UTXOManager {
	return &UTXOManager{
		Unspent:   map[string][]UTXO{},
		Caches:    map[string]map[string][]UTXO{},
		TmpIdx:    0,
		IncClient: incClient,
	}
}

package utxomanager

import (
	"fmt"

	"github.com/incognitochain/go-incognito-sdk-v2/common"
	"github.com/incognitochain/go-incognito-sdk-v2/common/base58"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
)

const (
	TmpPrefix = "Tmp"
)

func (c *UTXOManager) getCachedUTXOByPublicKey(publicKey string) map[string][]UTXO {
	_, isExisted := c.Caches[publicKey]
	if isExisted {
		return c.Caches[publicKey]
	}
	return map[string][]UTXO{}
}

func (c *UTXOManager) getListUTXOKeyImagesWithoutCached(utxos []UTXO, publicKey string) []UTXO {
	cachedUTXOs := c.getCachedUTXOByPublicKey(publicKey)

	cached := map[string]bool{} // key image: interface
	for _, utxos := range cachedUTXOs {
		for _, utxo := range utxos {
			keyImage := base58.Base58Check{}.Encode(utxo.Coin.GetKeyImage().ToBytesS(), common.ZeroByte)
			cached[keyImage] = true
		}
	}

	result := utxos
	idx := 0

	for idx < len(result) {
		keyImage := base58.Base58Check{}.Encode(result[idx].Coin.GetKeyImage().ToBytesS(), common.ZeroByte)
		_, isExisted := cached[keyImage]
		if isExisted {
			result = append(result[:idx], result[idx+1:]...)
		} else {
			idx++
		}
	}

	return result
}

func (c *UTXOManager) cacheUTXOsByTmpTxID(publicKey string, txID string, utxos []UTXO) {
	cachedUTXOs := c.getCachedUTXOByPublicKey(publicKey)
	cachedUTXOs[txID] = utxos
	c.Caches[publicKey] = cachedUTXOs
}

func (c *UTXOManager) uncachedUTXOsByCheckingTxID(publicKey string, incClient *incclient.IncClient) {
	cachedUTXOs := c.getCachedUTXOByPublicKey(publicKey)
	for txID := range cachedUTXOs {
		if len(txID) > len(TmpPrefix) && txID[:3] == TmpPrefix {
			continue
		}

		txDetail, err := incClient.GetTxDetail(txID)
		// tx was rejected or tx was confirmed
		if (txDetail == nil && err != nil) || (txDetail.IsInBlock) {
			delete(cachedUTXOs, txID)
		}
	}
	c.Caches[publicKey] = cachedUTXOs
}

func (c *UTXOManager) UpdateTxID(privateKey string, tmpTxID string, txID string) {
	c.mux.Lock()
	defer c.mux.Unlock()

	publicKey, err := getPublicKeyStr(privateKey)
	if err != nil {
		return
	}

	cachedUTXOs := c.getCachedUTXOByPublicKey(publicKey)
	val, isExisted := cachedUTXOs[tmpTxID]
	if isExisted {
		delete(cachedUTXOs, tmpTxID)
		cachedUTXOs[txID] = val
	}
	c.Caches[publicKey] = cachedUTXOs
}

func (c *UTXOManager) UncachedUTXOByTmpTxID(privateKey string, tmpTxID string) {
	c.mux.Lock()
	defer c.mux.Unlock()

	publicKey, err := getPublicKeyStr(privateKey)
	if err != nil {
		return
	}

	cachedUTXOs := c.getCachedUTXOByPublicKey(publicKey)
	_, isExisted := cachedUTXOs[tmpTxID]
	if isExisted {
		delete(cachedUTXOs, tmpTxID)
	}
	c.Caches[publicKey] = cachedUTXOs
}

func (c *UTXOManager) GetUTXOsByAmount(privateKey string, amount uint64) ([]UTXO, string, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	publicKey, err := getPublicKeyStr(privateKey)
	if err != nil {
		return []UTXO{}, "", err
	}

	rescan := true
	_, isExisted := c.Unspent[publicKey]
	if isExisted {
		sum := uint64(0)
		for _, utxo := range c.Unspent[publicKey] {
			sum += utxo.Coin.GetValue()
		}
		if sum >= amount {
			rescan = false
		}
	}

	if rescan {
		c.uncachedUTXOsByCheckingTxID(publicKey, c.IncClient)
		utxos, err := getListUTXOs(c.IncClient, privateKey)
		if err != nil {
			return []UTXO{}, "", err
		}
		c.Unspent[publicKey] = c.getListUTXOKeyImagesWithoutCached(utxos, publicKey)
	}

	sum := uint64(0)
	for idx, utxo := range c.Unspent[publicKey] {
		sum += utxo.Coin.GetValue()
		if sum >= amount {
			utxos := c.Unspent[publicKey][:idx+1]
			c.TmpIdx = (c.TmpIdx + 1) % 10000
			tmpTxID := fmt.Sprintf("%v%v", TmpPrefix, c.TmpIdx)
			c.cacheUTXOsByTmpTxID(publicKey, tmpTxID, utxos)
			c.Unspent[publicKey] = c.Unspent[publicKey][idx+1:]
			return utxos, tmpTxID, nil
		}
	}

	return []UTXO{}, "", fmt.Errorf("Not enough UTXO")
}

func (c *UTXOManager) CacheUTXOsByTxID(privateKey string, txID string, utxos []UTXO) {
	c.mux.Lock()
	defer c.mux.Unlock()

	publicKey, err := getPublicKeyStr(privateKey)
	if err != nil {
		return
	}

	cachedUTXOs := c.getCachedUTXOByPublicKey(publicKey)
	cachedUTXOs[txID] = utxos
	c.Caches[publicKey] = cachedUTXOs
}

func (c *UTXOManager) GetListUnspentUTXO(privateKey string) ([]UTXO, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	publicKey, err := getPublicKeyStr(privateKey)
	if err != nil {
		return []UTXO{}, err
	}

	c.uncachedUTXOsByCheckingTxID(publicKey, c.IncClient)
	utxos, err := getListUTXOs(c.IncClient, privateKey)
	if err != nil {
		return []UTXO{}, err
	}

	bigUTXOs := []UTXO{}
	for _, u := range utxos {
		if u.Coin.GetValue() >= MinUTXOAmount*MaxReceiver {
			bigUTXOs = append(bigUTXOs, u)
		}
	}
	c.Unspent[publicKey] = c.getListUTXOKeyImagesWithoutCached(bigUTXOs, publicKey)

	return c.Unspent[publicKey], nil
}

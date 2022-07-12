package submitproof

import (
	"bytes"
	"encoding/base64"
	"math/big"
	"strconv"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	"github.com/pkg/errors"
)

type Receipt struct {
	Result *types.Receipt `json:"result"`
}

type NormalResult struct {
	Result interface{} `json:"result"`
}

func encodeForDerive(list types.DerivableList, i int, buf *bytes.Buffer) []byte {
	buf.Reset()
	list.EncodeIndex(i, buf)
	// It's really unfortunate that we need to do perform this copy.
	// StackTrie holds onto the values until Hash is called, so the values
	// written to it must not alias.
	return common.CopyBytes(buf.Bytes())
}

// deriveBufferPool holds temporary encoder buffers for DeriveSha and TX encoding.
var encodeBufferPool = sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

const ADDRESS_0 = "0x0000000000000000000000000000000000000000"

func getETHDepositProof(
	client *incclient.IncClient,
	evmNetworkID int,
	txHash string,
) (*big.Int, string, uint, []string, error) {
	// Get tx content
	txContent, err := client.GetEVMTxByHash(txHash, evmNetworkID)
	if err != nil {
		return nil, "", 0, nil, err
	}
	blockHash, success := txContent["blockHash"].(string)
	if !success {
		return nil, "", 0, nil, err
	}
	txIndexStr, success := txContent["transactionIndex"].(string)
	if !success {
		return nil, "", 0, nil, errors.New("Cannot find transactionIndex field")
	}
	txIndex, err := strconv.ParseUint(txIndexStr[2:], 16, 64)
	if err != nil {
		return nil, "", 0, nil, err
	}

	// Get tx's block for constructing receipt trie
	blockNumString, success := txContent["blockNumber"].(string)
	if !success {
		return nil, "", 0, nil, errors.New("Cannot find blockNumber field")
	}
	blockNumber := new(big.Int)
	_, success = blockNumber.SetString(blockNumString[2:], 16)
	if !success {
		return nil, "", 0, nil, errors.New("Cannot convert blockNumber into integer")
	}

	blockHeader, err := client.GetEVMBlockByHash(blockHash, evmNetworkID)
	if err != nil {
		return nil, "", 0, nil, err
	}

	// Get all sibling Txs
	siblingTxs, success := blockHeader["transactions"].([]interface{})
	if !success {
		return nil, "", 0, nil, errors.New("Cannot find transactions field")
	}
	// Constructing the receipt trie (source: go-ethereum/core/types/derive_sha.go)
	keybuf := new(bytes.Buffer)
	receiptTrie := new(trie.Trie)
	receipts := make([]*types.Receipt, 0)
	for i, tx := range siblingTxs {
		siblingReceipt, err := client.GetEVMTxReceipt(tx.(string), evmNetworkID)
		if err != nil {
			return nil, "", 0, nil, err
		}
		if i == len(siblingTxs)-1 {
			txOut, err := client.GetEVMTxByHash(tx.(string), evmNetworkID)
			if err != nil {
				return nil, "", 0, nil, err
			}
			if txOut["to"] == ADDRESS_0 && txOut["from"] == ADDRESS_0 {
				break
			}
		}
		receipts = append(receipts, siblingReceipt)
	}

	receiptList := types.Receipts(receipts)
	receiptTrie.Reset()

	valueBuf := encodeBufferPool.Get().(*bytes.Buffer)
	defer encodeBufferPool.Put(valueBuf)

	// StackTrie requires values to be inserted in increasing hash order, which is not the
	// order that `list` provides hashes in. This insertion sequence ensures that the
	// order is correct.
	var indexBuf []byte
	for i := 1; i < receiptList.Len() && i <= 0x7f; i++ {
		indexBuf = rlp.AppendUint64(indexBuf[:0], uint64(i))
		value := encodeForDerive(receiptList, i, valueBuf)
		receiptTrie.Update(indexBuf, value)
	}
	if receiptList.Len() > 0 {
		indexBuf = rlp.AppendUint64(indexBuf[:0], 0)
		value := encodeForDerive(receiptList, 0, valueBuf)
		receiptTrie.Update(indexBuf, value)
	}
	for i := 0x80; i < receiptList.Len(); i++ {
		indexBuf = rlp.AppendUint64(indexBuf[:0], uint64(i))
		value := encodeForDerive(receiptList, i, valueBuf)
		receiptTrie.Update(indexBuf, value)
	}

	// Constructing the proof for the current receipt (source: go-ethereum/trie/proof.go)
	proof := light.NewNodeSet()
	keybuf.Reset()
	rlp.Encode(keybuf, uint(txIndex))
	err = receiptTrie.Prove(keybuf.Bytes(), 0, proof)
	if err != nil {
		return nil, "", 0, nil, err
	}
	nodeList := proof.NodeList()
	encNodeList := make([]string, 0)
	for _, node := range nodeList {
		str := base64.StdEncoding.EncodeToString(node)
		encNodeList = append(encNodeList, str)
	}
	return blockNumber, blockHash, uint(txIndex), encNodeList, nil
}

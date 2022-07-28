package submitproof

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"github.com/adjust/rmq/v4"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
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
) (*big.Int, string, uint, []string, string, error) {
	var contractID string
	// Get tx content
	txContent, err := client.GetEVMTxByHash(txHash, evmNetworkID)
	if err != nil {
		return nil, "", 0, nil, "", err
	}
	blockHash, success := txContent["blockHash"].(string)
	if !success {
		return nil, "", 0, nil, "", err
	}
	txIndexStr, success := txContent["transactionIndex"].(string)
	if !success {
		return nil, "", 0, nil, "", errors.New("Cannot find transactionIndex field")
	}
	txIndex, err := strconv.ParseUint(txIndexStr[2:], 16, 64)
	if err != nil {
		return nil, "", 0, nil, "", err
	}

	// Get tx's block for constructing receipt trie
	blockNumString, success := txContent["blockNumber"].(string)
	if !success {
		return nil, "", 0, nil, "", errors.New("Cannot find blockNumber field")
	}
	blockNumber := new(big.Int)
	_, success = blockNumber.SetString(blockNumString[2:], 16)
	if !success {
		return nil, "", 0, nil, "", errors.New("Cannot convert blockNumber into integer")
	}

	blockHeader, err := client.GetEVMBlockByHash(blockHash, evmNetworkID)
	if err != nil {
		return nil, "", 0, nil, "", err
	}

	// Get all sibling Txs
	siblingTxs, success := blockHeader["transactions"].([]interface{})
	if !success {
		return nil, "", 0, nil, "", errors.New("Cannot find transactions field")
	}
	// Constructing the receipt trie (source: go-ethereum/core/types/derive_sha.go)
	keybuf := new(bytes.Buffer)
	receiptTrie := new(trie.Trie)
	receipts := make([]*types.Receipt, 0)
	for i, tx := range siblingTxs {
		siblingReceipt, err := client.GetEVMTxReceipt(tx.(string), evmNetworkID)
		if err != nil {
			return nil, "", 0, nil, "", err
		}
		if i == len(siblingTxs)-1 {
			txOut, err := client.GetEVMTxByHash(tx.(string), evmNetworkID)
			if err != nil {
				return nil, "", 0, nil, "", err
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

	vaultABI, err := abi.JSON(strings.NewReader(VaultABI))
	if err != nil {
		fmt.Println("abi.JSON", err.Error())
		return nil, "", 0, nil, "", err
	}
	// erc20ABI, err := abi.JSON(strings.NewReader(IERC20ABI))
	// if err != nil {
	// 	fmt.Println("erc20ABI", err.Error())
	// 	return nil, "", 0, nil, "", err
	// }
	// erc20ABINoIndex, err := abi.JSON(strings.NewReader(Erc20ABINoIndex))
	// if err != nil {
	// 	fmt.Println("erc20ABINoIndex", err.Error())
	// 	return nil, "", 0, nil, "", err
	// }

	for _, v := range receiptList {
		for _, d := range v.Logs {
			switch len(d.Data) {
			// case 32:
			// 	unpackResult, err := erc20ABI.Unpack("Transfer", d.Data)
			// 	if err != nil {
			// 		fmt.Println("Unpack", err)
			// 		continue
			// 	}
			// 	if len(unpackResult) < 1 || len(d.Topics) < 3 {
			// 		err = errors.New(fmt.Sprintf("Unpack event error match data needed %v\n", unpackResult))
			// 		// b.notifyShieldDecentalized(queryAtHeight.Uint64(), err.Error(), conf)
			// 		fmt.Println("len(unpackResult)", err)
			// 		continue
			// 	}
			// 	fmt.Println("32", d.Address.String())
			// case 96:
			// 	unpackResult, err := erc20ABINoIndex.Unpack("Transfer", d.Data)
			// 	if err != nil {
			// 		fmt.Println("Unpack2", err)
			// 		continue
			// 	}
			// 	if len(unpackResult) < 3 {
			// 		err = errors.New(fmt.Sprintf("Unpack event not match data needed %v\n", unpackResult))
			// 		fmt.Println("len(unpackResult)2", err)
			// 		continue
			// 	}
			// 	fmt.Println("96", d.Address.String(), d.Address.Hex())
			// event indexed both from and to
			case 256, 288:
				unpackResult, err := vaultABI.Unpack("Deposit", d.Data)
				if err != nil {
					log.Println("unpackResult err", err)
					continue
				}
				if len(unpackResult) < 3 {
					err = errors.New(fmt.Sprintf("Unpack event not match data needed %v\n", unpackResult))
					log.Println("len(unpackResult) err", err)
					continue
				}
				fmt.Println("unpackResult", unpackResult)
				contractID = unpackResult[0].(common.Address).String()
			default:
			}
		}

	}
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
		return nil, "", 0, nil, "", err
	}
	nodeList := proof.NodeList()
	encNodeList := make([]string, 0)
	for _, node := range nodeList {
		str := base64.StdEncoding.EncodeToString(node)
		encNodeList = append(encNodeList, str)
	}
	return blockNumber, blockHash, uint(txIndex), encNodeList, contractID, nil
}

func findTokenByContractID(contractID string, networkID int) (string, string, error) {
	var pUtokenID string
	var linkedTokenID string
	tokenList, err := getTokenList()
	if err != nil {
		return "", "", err
	}
	contractID = strings.ToLower(contractID)
	if contractID == EthAddrStr {
		for _, token := range tokenList {
			if token.IsBridge && token.Verified && token.NetworkID == networkID {
				linkedTokenID = token.TokenID
				if token.MovedUnifiedToken {
					for _, pUtokenInfo := range tokenList {
						if pUtokenInfo.CurrencyType == 25 {
							for _, v := range pUtokenInfo.ListUnifiedToken {
								if v.TokenID == linkedTokenID {
									pUtokenID = pUtokenInfo.TokenID
									break
								}
							}
						}
					}
				} else {
					pUtokenID = token.TokenID
				}
				break
			}
		}
	} else {
		for _, token := range tokenList {
			if token.IsBridge && token.Verified {
				if token.ContractID == contractID && token.NetworkID == networkID && !token.MovedUnifiedToken { //non-punified
					pUtokenID = token.TokenID
					linkedTokenID = token.TokenID
					break
				}
				for _, childToken := range token.ListUnifiedToken { //punified
					if childToken.ContractID == contractID && childToken.NetworkID == networkID {
						pUtokenID = token.TokenID
						linkedTokenID = childToken.TokenID
						return pUtokenID, linkedTokenID, nil
					}
				}
			}
		}
	}
	return pUtokenID, linkedTokenID, nil
}

func getTokenInfo(pUTokenID string) (*wcommon.TokenInfo, error) {

	type APIRespond struct {
		Result []wcommon.TokenInfo
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

func getTokenList() ([]wcommon.TokenInfo, error) {
	result, err := retrieveTokenList()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func retrieveTokenList() ([]wcommon.TokenInfo, error) {
	type APIRespond struct {
		Result []wcommon.TokenInfo
		Error  *string
	}

	var responseBodyData APIRespond
	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).
		Get(config.CoinserviceURL + "/coins/tokenlist")
	if err != nil {
		return nil, err
	}
	return responseBodyData.Result, nil
}

// func getNativeToken(networkID string) (string, error) {

// }

func initIncClient(network string) error {
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
	}
	if err != nil {
		return err
	}
	return nil
}

func rejectDelivery(delivery rmq.Delivery, payload string) {
	if err := delivery.Reject(); err != nil {
		log.Printf("failed to reject %s: %s", payload, err)
		return
	} else {
		log.Printf("rejected %s", payload)
		return
	}
}

func ackDelivery(delivery rmq.Delivery, payload string) {
	if err := delivery.Ack(); err != nil {
		log.Printf("failed to ack %s: %s", payload, err)
		return
	} else {
		log.Printf("acked %s", payload)
		return
	}
}

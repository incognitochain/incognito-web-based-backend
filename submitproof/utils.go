package submitproof

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"

	"github.com/adjust/rmq/v4"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/light"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/incognitochain/bridge-eth/bridge/vault"
	"github.com/incognitochain/go-incognito-sdk-v2/coin"
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
	evmClient *ethclient.Client,
	txHashStr string,
) (*big.Int, string, uint, []string, string, string, bool, string, uint64, string, bool, error) {
	var contractID string
	var paymentaddress string
	var otaStr string
	var shieldAmount uint64
	var isRedeposit bool
	var logResult string
	var isTxPass bool

	txHash := common.Hash{}
	err := txHash.UnmarshalText([]byte(txHashStr))
	if err != nil {
		return nil, "", 0, nil, "", "", false, "", 0, "", isTxPass, err
	}
	txReceipt, err := evmClient.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		return nil, "", 0, nil, "", "", false, "", 0, "", isTxPass, err
	}

	txIndex := txReceipt.TransactionIndex
	blockHash := txReceipt.BlockHash.String()
	blockNumber := txReceipt.BlockNumber

	blk, err := evmClient.BlockByHash(context.Background(), txReceipt.BlockHash)
	if err != nil {
		return nil, "", 0, nil, "", "", false, "", 0, "", isTxPass, err
	}
	log.Println("txReceipt.Status", txReceipt.Status, txReceipt)

	if txReceipt.Status == 1 {
		isTxPass = true
	}
	siblingTxs := blk.Body().Transactions
	keybuf := new(bytes.Buffer)
	receiptTrie := new(trie.Trie)
	receipts := make([]*types.Receipt, 0)

	for i, siblingTx := range siblingTxs {
		siblingReceipt, err := evmClient.TransactionReceipt(context.Background(), siblingTx.Hash())
		if err != nil {
			return nil, "", 0, nil, "", "", false, "", 0, "", isTxPass, err
		}
		if i == len(siblingTxs)-1 {
			txData, _, err := evmClient.TransactionByHash(context.Background(), siblingTx.Hash())
			if err != nil {
				return nil, "", 0, nil, "", "", false, "", 0, "", isTxPass, err
			}
			from, err := evmClient.TransactionSender(context.Background(), txData, txReceipt.BlockHash, uint(i))
			if err != nil {
				return nil, "", 0, nil, "", "", false, "", 0, "", isTxPass, err
			}
			if txData.To().String() == ADDRESS_0 && from.String() == ADDRESS_0 {
				break
			}
		}
		receipts = append(receipts, siblingReceipt)
	}

	receiptList := types.Receipts(receipts)
	receiptTrie.Reset()

	valueBuf := encodeBufferPool.Get().(*bytes.Buffer)
	defer encodeBufferPool.Put(valueBuf)

	vaultABI, err := abi.JSON(strings.NewReader(vault.VaultABI))
	if err != nil {
		fmt.Println("abi.JSON", err.Error())
		return nil, "", 0, nil, "", "", false, "", 0, "", isTxPass, err
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
				if contractID == "" {
					unpackResult, err := vaultABI.Unpack("Redeposit", d.Data)
					if err != nil {
						unpackResult, err = vaultABI.Unpack("Deposit", d.Data)
						if err != nil {
							log.Println("unpackResult3 err", err)
							continue
						}
					}
					if len(unpackResult) < 3 {
						err = errors.New(fmt.Sprintf("Unpack event not match data needed %v\n", unpackResult))
						log.Println("len(unpackResult) err", err)
						return nil, "", 0, nil, "", "", false, "", 0, "", isTxPass, err
					}
					contractID = unpackResult[0].(common.Address).String()
					amount := unpackResult[2].(*big.Int)
					shieldAmount = amount.Uint64()

					var ok bool
					paymentaddress, ok = unpackResult[1].(string)
					if !ok {
						OTAReceiver := unpackResult[1].([]byte)
						newOTA := coin.OTAReceiver{}
						err = newOTA.SetBytes(OTAReceiver)
						if err != nil {
							log.Println("unpackResult4 err", err)
							continue
						}
						isRedeposit = true
						otaStr, err = newOTA.String()
						if err != nil {
							log.Println("unpackResult5err", err)
							return nil, "", 0, nil, "", "", false, "", 0, "", isTxPass, err
						}
					}
				}
				unpackResult, err := vaultABI.Unpack("ExecuteFnLog", d.Data)
				if err != nil {
					log.Println("unpackResult2 err", err)
					continue
				} else {
					logResult = fmt.Sprintf("%s", unpackResult)
					log.Println("logResult", logResult)
				}
			default:
				unpackResult, err := vaultABI.Unpack("ExecuteFnLog", d.Data)
				if err != nil {
					log.Println("unpackResult2 err", err)
					continue
				} else {
					logResult = fmt.Sprintf("%s", unpackResult)
					log.Println("logResult", logResult)
				}
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
	err = rlp.Encode(keybuf, uint(txIndex))
	if err != nil {
		return nil, "", 0, nil, "", "", false, "", 0, "", isTxPass, err
	}
	err = receiptTrie.Prove(keybuf.Bytes(), 0, proof)
	if err != nil {
		return nil, "", 0, nil, "", "", false, "", 0, "", isTxPass, err
	}
	nodeList := proof.NodeList()
	encNodeList := make([]string, 0)
	for _, node := range nodeList {
		str := base64.StdEncoding.EncodeToString(node)
		encNodeList = append(encNodeList, str)
	}
	return blockNumber, blockHash, uint(txIndex), encNodeList, contractID, paymentaddress, isRedeposit, otaStr, shieldAmount, logResult, isTxPass, nil
}

func findTokenByContractID(contractID string, networkID int) (string, string, error) {
	var pUtokenID string
	var linkedTokenID string
	tokenList, err := getTokenList()
	if err != nil {
		return "", "", err
	}
	contractID = strings.ToLower(contractID)
	if contractID == EthNativeAddrStr {
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
				tkContractID := strings.ToLower(token.ContractID)
				tkNetworkID, err := wcommon.GetNetworkIDFromCurrencyType(token.CurrencyType)
				if err != nil {
					continue
				}
				if tkContractID == contractID && tkNetworkID == networkID && !token.MovedUnifiedToken { //non-punified
					pUtokenID = token.TokenID
					linkedTokenID = token.TokenID
					break
				}
				for _, childToken := range token.ListUnifiedToken { //punified
					ctkContractID := strings.ToLower(childToken.ContractID)

					ctkNetworkID, err := wcommon.GetNetworkIDFromCurrencyType(childToken.CurrencyType)
					if err != nil {
						continue
					}
					if ctkContractID == contractID && ctkNetworkID == networkID {
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
	default:
		incClient, err = incclient.NewIncClient(config.FullnodeURL, "", 2, network)
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

func faucetPRV(paymentaddress string) {
	if config.FaucetService != "" && paymentaddress != "" {
		req := struct {
			PaymentAddress string `json:"paymentaddress"`
		}{PaymentAddress: paymentaddress}
		_, err := restyClient.R().
			EnableTrace().
			SetHeader("Content-Type", "application/json").SetBody(req).
			Post(config.FaucetService)
		if err != nil {
			log.Println("faucetPRV err:", err)
			return
		}
	}
}

func getEVMBlockHeight(endpoints []string) (uint64, error) {
	for _, endpoint := range endpoints {
		evmClient, err := ethclient.Dial(endpoint)
		if err != nil {
			return 0, err
		}
		result, err := evmClient.BlockNumber(context.Background())
		if err != nil {
			log.Println(err)
			continue
		}
		return result, nil
	}
	return 0, errors.New("failed to get EVM block height")
}

func getNonceByPrivateKey(c *ethclient.Client, senderPrivKey string) (uint64, error) {
	privateKey, err := crypto.HexToECDSA(senderPrivKey)
	if err != nil {
		return 0, errors.Wrap(err, "crypto.HexToECDSA")
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := c.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return 0, errors.Wrap(err, "s.ethClient.PendingNonceAt")
	}

	return nonce, nil
}

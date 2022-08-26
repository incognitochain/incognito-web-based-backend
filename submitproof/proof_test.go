package submitproof

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/incognitochain/bridge-eth/bridge/vault"
	"github.com/incognitochain/incognito-web-based-backend/evmproof"
)

// type HelpTestSuite struct {
// 	suite.Suite

// 	MasterKey string
// 	Token     map[string]string

// 	FN string
// }

// func (t *HelpTestSuite) SetupTest() {
// 	t.FN = "http://51.161.117.193:11334"
// }

// func TestHelpTestSuite(t *testing.T) {
// 	suite.Run(t, new(HelpTestSuite))
// }

func TestSubmitProof(t *testing.T) {
	incTxHash := "a2e2c585a02f25860c1a13a8b0401a2e9d6f60a26af254a48b45cdb5c892a98e"
	fullnode := "http://172.105.114.134:8334"
	endpoint := "https://matic-mumbai.chainstacklabs.com"
	privKeyStr := "cf83433a251a6e6c5a7fea5eb6448bb7e7366d8f65d1fbf61bac517412ccc4bd"
	privKey, _ := crypto.HexToECDSA(privKeyStr)

	t.Log("GetProof abc")

	// proof, err := bridgeeth.GetAndDecodeBurnProofV2(fullnode, incTxHash, "getburnplgprooffordeposittosc")

	proof, err := evmproof.GetAndDecodeBurnProofUnifiedToken(fullnode, incTxHash, 0, uint(3))
	t.Log(err)

	t.Log(proof)

	evmClient, err := ethclient.Dial(endpoint)
	if err != nil {
		t.Fatal(err)
	}

	c, err := vault.NewVault(common.HexToAddress("0x76318093c374e39B260120EBFCe6aBF7f75c8D28"), evmClient)
	if err != nil {
		t.Fatal(err)
	}
	maxGasTip, err := evmClient.SuggestGasTipCap(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	gasPrice, err := evmClient.SuggestGasPrice(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// nonce, err := getNonceByPrivateKey(evmClient, config.EVMKey)
	// if err != nil {
	// 	log.Println(err)
	// 	continue
	// }

	chainID := new(big.Int).SetInt64(80001)

	auth, err := bind.NewKeyedTransactorWithChainID(privKey, chainID)
	if err != nil {
		t.Fatal(err)
	}
	auth.GasPrice = gasPrice

	t.Logf("\n maxGasTip: %v \n", maxGasTip)
	t.Logf("\n GasPrice: %v \n", auth.GasPrice)
	tx, err := evmproof.ExecuteWithBurnProof(c, auth, proof)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(tx.Hash().String())
}

func TestRedpositEvent(t *testing.T) {

	endpoint := "https://matic-mumbai.chainstacklabs.com"
	txStr := "0x2a378a3ee8d1d346ef698643336e4b29c26e1cf8c03aa7b85c20013aeef1af0f"
	evmClient, _ := ethclient.Dial(endpoint)

	blockNumber, blockHash, txIdx, proof, contractID, paymentAddr, isRedeposit, otaStr, shieldAmount, logResult, err := getETHDepositProof(evmClient, txStr)

	t.Log(err)

	t.Logf("\n blockNumber: %v, blockHash: %v, txIdx: %v, contractID: %v, isRedeposit: %v, paymentAddr: %v, otaStr: %v, shieldAmount: %v\n", blockNumber, blockHash, txIdx, contractID, isRedeposit, paymentAddr, otaStr, shieldAmount)

	t.Logf("\n logResult: %v \n", logResult)
	_ = proof
	// t.Log("proof", proof)

	// txHash := common.Hash{}
	// err := txHash.UnmarshalText([]byte("0x30eb9dca48fd5ccb4533e92f4f1926765ec0eade6478b7726ae204a45d1b14bf"))
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// txReceipt, err := evmClient.TransactionReceipt(context.Background(), txHash)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// blk, err := evmClient.BlockByHash(context.Background(), txReceipt.BlockHash)

	// for _, v := range blk.Body().Transactions {
	// 	v.Hash()
	// }

	// // if currentEVMHeight >= txReceipt.BlockNumber.Uint64()+finalizeRange {

	// // check sc re-deposit event
	// valueBuf := encodeBufferPool.Get().(*bytes.Buffer)
	// defer encodeBufferPool.Put(valueBuf)

	// vaultABI, err := abi.JSON(strings.NewReader(vault.VaultABI))
	// if err != nil {
	// 	t.Log("abi.JSON", err.Error())
	// 	return
	// }

	// // erc20ABI, err := abi.JSON(strings.NewReader(IERC20ABI))
	// // if err != nil {
	// // 	fmt.Println("erc20ABI", err.Error())
	// // 	return nil, "", 0, nil, "", err
	// // }
	// // erc20ABINoIndex, err := abi.JSON(strings.NewReader(Erc20ABINoIndex))
	// // if err != nil {
	// // 	fmt.Println("erc20ABINoIndex", err.Error())
	// // 	return nil, "", 0, nil, "", err
	// // }

	// t.Logf("txReceipt.Logs %v \n", len(txReceipt.Logs))

	// for _, d := range txReceipt.Logs {
	// 	t.Logf("d.Data %v \n", len(d.Data))

	// 	switch len(d.Data) {
	// 	// case 32:
	// 	// 	unpackResult, err := erc20ABI.Unpack("Transfer", d.Data)
	// 	// 	if err != nil {
	// 	// 		fmt.Println("Unpack", err)
	// 	// 		continue
	// 	// 	}
	// 	// 	if len(unpackResult) < 1 || len(d.Topics) < 3 {
	// 	// 		err = errors.New(fmt.Sprintf("Unpack event error match data needed %v\n", unpackResult))
	// 	// 		// b.notifyShieldDecentalized(queryAtHeight.Uint64(), err.Error(), conf)
	// 	// 		fmt.Println("len(unpackResult)", err)
	// 	// 		continue
	// 	// 	}
	// 	// 	fmt.Println("32", d.Address.String())
	// 	// case 96:
	// 	// 	unpackResult, err := erc20ABINoIndex.Unpack("Transfer", d.Data)
	// 	// 	if err != nil {
	// 	// 		fmt.Println("Unpack2", err)
	// 	// 		continue
	// 	// 	}
	// 	// 	if len(unpackResult) < 3 {
	// 	// 		err = errors.New(fmt.Sprintf("Unpack event not match data needed %v\n", unpackResult))
	// 	// 		fmt.Println("len(unpackResult)2", err)
	// 	// 		continue
	// 	// 	}
	// 	// 	fmt.Println("96", d.Address.String(), d.Address.Hex())
	// 	// event indexed both from and to
	// 	case 256, 288:
	// 		unpackResult, err := vaultABI.Unpack("Redeposit", d.Data)
	// 		if err != nil {
	// 			t.Log("unpackResult err", err)
	// 			continue
	// 		}
	// 		if len(unpackResult) < 3 {
	// 			err = errors.New(fmt.Sprintf("Unpack event not match data needed %v\n", unpackResult))
	// 			t.Log("len(unpackResult) err", err)
	// 			continue
	// 		}
	// 		t.Log("unpackResult", unpackResult)
	// 		contractID := unpackResult[0].(common.Address).String()
	// 		OTAReceiver := unpackResult[1].([]byte)

	// 		newOTA := coin.OTAReceiver{}

	// 		err = newOTA.SetBytes(OTAReceiver)
	// 		t.Log("len(OTAReceiver)", len(OTAReceiver))
	// 		t.Log("err", err)

	// 		amount := unpackResult[2].(*big.Int)

	// 		newOTA.String()

	// 		t.Logf("\n contractID: %v, paymentaddress: %v, amount: %v\n", contractID, base58.FastBase58Encoding(OTAReceiver), amount.Uint64())
	// 	default:
	// 		// log.Println("invalid event index")
	// 	}
	// 	// }
	// 	// txReceipt.CumulativeGasUsed
	// 	// txReceipt.Logs
	// 	// vault.VaultABI
	// }
}

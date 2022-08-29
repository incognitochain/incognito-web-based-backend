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
	txStr := "0xd831f7af21c79ea7b69a0a003c4b71536c016c1a9e3a0b13625d54f66afadf63"
	evmClient, _ := ethclient.Dial(endpoint)

	blockNumber, blockHash, txIdx, proof, contractID, paymentAddr, isRedeposit, otaStr, shieldAmount, logResult, err := getETHDepositProof(evmClient, txStr)

	t.Log(err)

	t.Logf("\n blockNumber: %v, blockHash: %v, txIdx: %v, contractID: %v, isRedeposit: %v, paymentAddr: %v, otaStr: %v, shieldAmount: %v\n", blockNumber, blockHash, txIdx, contractID, isRedeposit, paymentAddr, otaStr, shieldAmount)

	t.Logf("\n logResult: %v \n", logResult)
	_ = proof
}

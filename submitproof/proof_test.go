package submitproof

import (
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/incognitochain/go-incognito-sdk-v2/coin"
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
	incTxHash := "21cfd0d3ed2f922f03b9907783e26d76fcc6aaa55df645c29afb6e8cc67ff481"
	fullnode := "https://testnet.incognito.org/fullnode"
	// endpoint := "https://matic-mumbai.chainstacklabs.com"
	// privKeyStr := "cf83433a251a6e6c5a7fea5eb6448bb7e7366d8f65d1fbf61bac517412ccc4bd"
	// privKey, _ := crypto.HexToECDSA(privKeyStr)

	t.Log("GetProof abc")

	// proof, err := bridgeeth.GetAndDecodeBurnProofV2(fullnode, incTxHash, "getburnplgprooffordeposittosc")

	proof, err := evmproof.GetAndDecodeBurnProofUnifiedToken(fullnode, incTxHash, 0)
	t.Log(err)

	t.Log(proof)

	// evmClient, err := ethclient.Dial(endpoint)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// c, err := vault.NewVault(common.HexToAddress("0x76318093c374e39B260120EBFCe6aBF7f75c8D28"), evmClient)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// maxGasTip, err := evmClient.SuggestGasTipCap(context.Background())
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// gasPrice, err := evmClient.SuggestGasPrice(context.Background())
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// // nonce, err := getNonceByPrivateKey(evmClient, config.EVMKey)
	// // if err != nil {
	// // 	log.Println(err)
	// // 	continue
	// // }

	// chainID := new(big.Int).SetInt64(80001)

	// auth, err := bind.NewKeyedTransactorWithChainID(privKey, chainID)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// auth.GasPrice = gasPrice

	// t.Logf("\n maxGasTip: %v \n", maxGasTip)
	// t.Logf("\n GasPrice: %v \n", auth.GasPrice)
	// tx, err := evmproof.ExecuteWithBurnProof(c, auth, proof)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// t.Log(tx.Hash().String())
}

func TestRedpositEvent(t *testing.T) {

	endpoint := "https://matic-mumbai.chainstacklabs.com"
	// endpoint := "https://data-seed-prebsc-1-s1.binance.org:8545"
	txStr := "0x85ca18aaa57f2c553e5ec4fa0abf9c5744976d5eaf17f2ae93d48b2fb278dbe4"
	evmClient, _ := ethclient.Dial(endpoint)

	blockNumber, blockHash, txIdx, proof, contractID, paymentAddr, isRedeposit, otaStr, shieldAmount, logResult, status, err := getETHDepositProof(evmClient, txStr)
	_ = status
	t.Log(err)

	t.Logf("\n blockNumber: %v, blockHash: %v, txIdx: %v, contractID: %v, isRedeposit: %v, paymentAddr: %v, otaStr: %v, shieldAmount: %v\n", blockNumber, blockHash, txIdx, contractID, isRedeposit, paymentAddr, otaStr, shieldAmount)

	t.Logf("\n logResult: %v \n", logResult)
	_ = proof
}

func TestOTADecode(t *testing.T) {
	ota := "16SWHaPJFNNx9GBsDW658eFji8AkVQUaWEhkzUnCYWc1DTeMyrVHGnebj9LjtjZFeMtvpYaEmSQA3zK1gBrZbbTu8LYm5NM48VPVLg21CsySHUxCtNjfGB5dACYzmsYWAhcbDqdBefosMsuL"
	otacoin := new(coin.OTAReceiver)
	err := otacoin.FromString(ota)
	t.Log(err)

	t.Logf("\n PublicKey: %v \n", otacoin.PublicKey.String())
	t.Logf("\n TxRandom: %v \n", string(otacoin.TxRandom.Bytes()))
	otacoin.String()
}

// func TestGenKey(t *testing.T) {
// 	var err error
// 	incClient, err = incclient.NewIncClient("https://testnet.incognito.org/fullnode", "", 2, "testnet")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	err = genShardsAccount("112t8rnXUbFHzsnX7zdQouzxXEWArruE4rYzeswrEtvL3iBkcgXAXsQk4kQk23XfLNU6wMknyKk8UAu8fLBfkcUVMgxTNsfrYZURAnPqhffA")
// 	t.Log(err)
// }

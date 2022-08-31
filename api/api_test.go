package api

import (
	"testing"

	"github.com/incognitochain/go-incognito-sdk-v2/coin"
	"github.com/incognitochain/go-incognito-sdk-v2/common"
	"github.com/incognitochain/go-incognito-sdk-v2/common/base58"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	"github.com/incognitochain/go-incognito-sdk-v2/metadata/bridge"
	"github.com/incognitochain/go-incognito-sdk-v2/wallet"
	"github.com/stretchr/testify/suite"
)

type HelpTestSuite struct {
	suite.Suite

	incClient *incclient.IncClient
	MasterKey string
	Token     map[string]string

	Service string
	FN      string
	Network string
}

func (t *HelpTestSuite) SetupTest() {
	t.Service = "http://51.161.117.193:9898"
	t.FN = "http://172.105.114.134:8334"
	t.Network = "testnet"
	t.MasterKey = "112t8roafGgHL1rhAP9632Yef3sx5k8xgp8cwK4MCJsCL1UWcxXvpzg97N4dwvcD735iKf31Q2ZgrAvKfVjeSUEvnzKJyyJD3GqqSZdxN4or"
	incClient, err := incclient.NewIncClient(t.FN, "", 2, t.Network)
	if err != nil {
		t.FailNow(err.Error())
	}
	t.incClient = incClient
}

func TestHelpTestSuite(t *testing.T) {
	suite.Run(t, new(HelpTestSuite))
}

func (t *HelpTestSuite) TestSwapFlowUniSwapSuccess() {

}

func (t *HelpTestSuite) TestSwapFlowUniSwapRedeposit() {

}

func (t *HelpTestSuite) TestSwapFlowUniSwapRedepositWrongFee() {

	// 112t8rnX6JdMsZJngeixtnKLziMYm6tyH9w2Z2qd7zdHyd2nSxybzqMDFAwjTqZrk4UHsybDtodLKbWoj8tbP2rdtQXNyjM1s7K2AK3DyCuJ
	tokenID := common.Hash{}
	_, _ = tokenID.NewHashFromStr("f5d88e2e3c8f02d6dc1e01b54c90f673d730bef7d941aeec81ad1e1db690961f")

	receiveTokenID := common.Hash{}
	_, _ = receiveTokenID.NewHashFromStr("dae027b21d8d57114da11209dce8eeb587d01adf59d4fc356a8be5eedc146859")

	keyWallet, _ := wallet.Base58CheckDeserialize(t.MasterKey)

	var recv coin.OTAReceiver
	err := recv.FromAddress(keyWallet.KeySet.PaymentAddress)
	if err != nil {
		t.T().Fatal(err)
	}

	burnData := bridge.BurnForCallRequestData{
		BurningAmount:       2000000,
		ExternalNetworkID:   3,
		ExternalCallAddress: "68b3465833fb72a70ecdf485e0e4c7bd8665fc45",
		IncTokenID:          receiveTokenID,
		ReceiveToken:        "a6fa4fb5f76172d178d61b04b0ecd319c5d1c0aa",
		WithdrawAddress:     "0000000000000000000000000000000000000000",
		RedepositReceiver:   recv,
		ExternalCalldata:    "c8dc75e600000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000008000000000000000000000000076318093c374e39b260120ebfce6abf7f75c8d2800000000000000000000000000000000000000000000000000470de4df8200000000000000000000000000000000000000000000000000000000cb5c9ef7a1a2000000000000000000000000000000000000000000000000000000000000002b9c3c9283d3e44854697cd22d3faa240cfb03288904c6a8a6fa4fb5f76172d178d61b04b0ecd319c5d1c0aa000000000000000000000000000000000000000000",
	}

	md := bridge.BurnForCallRequest{
		BurnTokenID: tokenID,
		Data:        []bridge.BurnForCallRequestData{burnData},
	}

	// receiverList := []string{"12seyCLbpyNuz3mjtiWKegnE3dGY1nDtvrhgYwxfoHvyj6pDA3Bw1rkSE9HUwCnGeJn5ai4mLmbhB4CgNi8KRCbaR49BvbiuAxfLM6sjhJqVfkGWvrEBbAMsuEMNvZymnGmLZdnmvt7Q9Grc8qBY"}
	receiverList := []string{}
	amount := []uint64{}

	receiverListToken := []string{common.BurningAddress2}
	amountToken := []uint64{2000000}
	tokenParam := incclient.NewTxTokenParam("f5d88e2e3c8f02d6dc1e01b54c90f673d730bef7d941aeec81ad1e1db690961f", 1, receiverListToken, amountToken, false, 0, nil)

	txParam := incclient.NewTxParam(t.MasterKey, receiverList, amount, 100, tokenParam, &md, nil)

	txRaw, txHash, err := t.incClient.CreateRawTokenTransactionVer2(txParam)
	if err != nil {
		t.T().Fatal(err)
	}

	t.T().Log(txHash, base58.Base58Check{}.Encode(txRaw, 0x00))
}

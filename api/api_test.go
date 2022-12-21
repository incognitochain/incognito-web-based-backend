package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"strings"
	"testing"

	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
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

func (t *HelpTestSuite) TestSwapFlowUniSwapRedeposit() {
	o := "43000"
	amountOutBig, _ := new(big.Int).SetString(o, 10)
	t.T().Log("amountOutBig", amountOutBig.Uint64())
	amountOutBig = amountOutBig.Div(amountOutBig, big.NewInt(100))
	amountOutBig = amountOutBig.Div(amountOutBig, big.NewInt(10))
	t.T().Log("amountOutBig", amountOutBig.String())
}

// func (t *HelpTestSuite) TestSwapFlowUniSwapRedepositWrongFee() {

// 	// 112t8rnX6JdMsZJngeixtnKLziMYm6tyH9w2Z2qd7zdHyd2nSxybzqMDFAwjTqZrk4UHsybDtodLKbWoj8tbP2rdtQXNyjM1s7K2AK3DyCuJ
// 	tokenID, err := common.Hash{}.NewHashFromStr("f5d88e2e3c8f02d6dc1e01b54c90f673d730bef7d941aeec81ad1e1db690961f")
// 	if err != nil {
// 		t.T().Fatal(err)
// 	}

// 	receiveTokenID, err := common.Hash{}.NewHashFromStr("dae027b21d8d57114da11209dce8eeb587d01adf59d4fc356a8be5eedc146859")
// 	if err != nil {
// 		t.T().Fatal(err)
// 	}
// 	keyWallet, _ := wallet.Base58CheckDeserialize(t.MasterKey)

// 	var recv coin.OTAReceiver
// 	err = recv.FromAddress(keyWallet.KeySet.PaymentAddress)
// 	if err != nil {
// 		t.T().Fatal(err)
// 	}

// 	burnData := bridge.BurnForCallRequestData{
// 		BurningAmount:       2000000,
// 		ExternalNetworkID:   3,
// 		ExternalCallAddress: "68b3465833fb72a70ecdf485e0e4c7bd8665fc45",
// 		IncTokenID:          *receiveTokenID,
// 		ReceiveToken:        "a6fa4fb5f76172d178d61b04b0ecd319c5d1c0aa",
// 		WithdrawAddress:     "0000000000000000000000000000000000000000",
// 		RedepositReceiver:   recv,
// 		ExternalCalldata:    "c8dc75e600000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000008000000000000000000000000076318093c374e39b260120ebfce6abf7f75c8d2800000000000000000000000000000000000000000000000000470de4df8200000000000000000000000000000000000000000000000000000000cb5c9ef7a1a2000000000000000000000000000000000000000000000000000000000000002b9c3c9283d3e44854697cd22d3faa240cfb03288904c6a8a6fa4fb5f76172d178d61b04b0ecd319c5d1c0aa000000000000000000000000000000000000000000",
// 	}

// 	md := bridge.BurnForCallRequest{
// 		BurnTokenID:  *tokenID,
// 		Data:         []bridge.BurnForCallRequestData{burnData},
// 		MetadataBase: *metadata.NewMetadataBase(metadata.BurnForCallRequestMeta),
// 	}

// 	// receiverList := []string{"12seyCLbpyNuz3mjtiWKegnE3dGY1nDtvrhgYwxfoHvyj6pDA3Bw1rkSE9HUwCnGeJn5ai4mLmbhB4CgNi8KRCbaR49BvbiuAxfLM6sjhJqVfkGWvrEBbAMsuEMNvZymnGmLZdnmvt7Q9Grc8qBY"}
// 	receiverList := []string{}
// 	amount := []uint64{}

// 	receiverListToken := []string{common.BurningAddress2}
// 	amountToken := []uint64{2000000}
// 	tokenParam := incclient.NewTxTokenParam("f5d88e2e3c8f02d6dc1e01b54c90f673d730bef7d941aeec81ad1e1db690961f", 1, receiverListToken, amountToken, false, 0, nil)

// 	txParam := incclient.NewTxParam(t.MasterKey, receiverList, amount, 100, tokenParam, &md, nil)

// 	txRaw, txHash, err := t.incClient.CreateRawTokenTransactionVer2(txParam)
// 	if err != nil {
// 		t.T().Fatal(err)
// 	}

// 	// err = t.incClient.SendRawTokenTx(txRaw)
// 	// if err != nil {
// 	// 	t.T().Fatal(err)
// 	// }

// 	rawTxBytes, _, err := base58.Base58Check{}.Decode(string(txRaw))
// 	if err != nil {
// 		t.T().Fatal(err)
// 	}

// 	// t.T().Log(txHash, base58.Base58Check{}.Encode(txRaw, 0x00))
// 	t.T().Logf("\n")

// 	mdRaw, isPRVTx, outCoins, txHash, err := extractDataFromRawTx(rawTxBytes)
// 	t.T().Logf("mdRaw %v, isPRVTx %v, outCoins %v, txHash %v\n", mdRaw, isPRVTx, len(outCoins), txHash)
// 	if err != nil {
// 		t.T().Fatal(err)
// 	}

// 	t.T().Logf("\n")
// 	md2, ok := mdRaw.(*bridge.BurnForCallRequest)
// 	if !ok {
// 		t.T().Fatal(errors.New("invalid tx metadata type"))
// 	}

// 	t.T().Logf("\n")

// 	wl, err := wallet.Base58CheckDeserialize("112t8rnX6JdMsZJngeixtnKLziMYm6tyH9w2Z2qd7zdHyd2nSxybzqMDFAwjTqZrk4UHsybDtodLKbWoj8tbP2rdtQXNyjM1s7K2AK3DyCuJ")
// 	if err != nil {
// 		t.T().Fatal(err)
// 	}
// 	if wl.KeySet.OTAKey.GetOTASecretKey() == nil {
// 		t.T().Fatal(err)
// 	}
// 	incFeeKeySet = wl

// 	valid, networkList, feeToken, feeAmount, feeDiff, receiveToken, receiveAmount, err := checkValidTxSwap(md2, outCoins)
// 	if err != nil {
// 		t.T().Fatal(err)
// 	}

// 	t.T().Logf("valid %v, networkList %v, feeToken %v, feeAmount %v, receiveToken %v, receiveAmount %v, feeDiff %v", valid, networkList, feeToken, feeAmount, receiveToken, receiveAmount, feeDiff)
// 	// tx, _ := transaction.DeserializeTransactionJSON(txRaw)
// 	// t.T().Log(tx.TokenVersion2)
// }

// func estimateFee(tokenSource string, amount int64, tokenDest string, network string) (string, int64, string, error) {
// 	var err error
// 	var feeToken string
// 	var feeAddress string
// }

func TestSwapPDex(t *testing.T) {
	config.CoinserviceURL = "http://51.161.117.193:8096"
	tkFromInfo, _ := getTokenInfo("0000000000000000000000000000000000000000000000000000000000000004")
	result := estimateSwapFeeWithPdex("0000000000000000000000000000000000000000000000000000000000000004", "00000000000000000000000000000000000000000000000000000000000115dc", "0.0001", nil, tkFromInfo)
	t.Log("result", result)
}

func TestCompareTokenList(t *testing.T) {

	type TokenStruct struct {
		ID string `json:"contractid"`
	}

	type TokenDBStruct struct {
		TokenID    string `json:"tokenid"`
		ContractID string `json:"contractid"`
		Verified   bool   `json:"verify"`
	}

	contractList := []TokenStruct{}

	data, err := ioutil.ReadFile("./tokens.json")
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			log.Fatalln(err)
		} else {
			log.Fatalln(err)
		}
	}
	err = json.Unmarshal(data, &contractList)
	if err != nil {
		log.Fatalln(err)
	}
	tokenMap := make(map[string]struct{})
	for _, v := range contractList {
		tokenMap[strings.ToLower(v.ID)] = struct{}{}
	}

	err = parseDefaultToken()
	if err != nil {
		t.Fatal(err)
	}
	notExistList := []TokenDBStruct{}
	for tk, _ := range whiteListTokenContract {
		if _, exist := tokenMap[strings.ToLower(tk)]; !exist {
			a := TokenDBStruct{
				TokenID:    "",
				ContractID: tk,
				Verified:   true,
			}
			notExistList = append(notExistList, a)
		}
	}
	fmt.Println("len(notExistList)", len(notExistList))
	fmt.Println(notExistList)

}

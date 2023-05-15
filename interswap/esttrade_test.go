package interswap

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/incognitochain/incognito-web-based-backend/common"
)

func TestEstimateSwap(t *testing.T) {
	// InitSupportedMidTokens("mainnet")

	// XMR := "c01e7dc1d1aba995c19b257412340b057f8ad1482ccb6a9bb0adce61afbf05d4"
	// USDC := "545ef6e26d4d428b16117523935b6be85ec0a63e8c2afeb0162315eb0ce3d151"
	// from := XMR
	// to := USDC

	// params := &EstimateSwapParam{
	// 	Network:   "bsc",
	// 	Amount:    "1",
	// 	FromToken: from,
	// 	ToToken:   to,
	// 	Slippage:  "0.5",
	// }
	// res, err := EstimateSwap(params)
	// fmt.Printf("Res: %+v\n", res)
	// fmt.Printf("err: %+v\n", err)
}

func TestMashalListKeys(t *testing.T) {
	keys := map[string]string{}
	keyBytes, err := json.Marshal(keys)
	if err != nil {
		fmt.Printf("Err: $%v\n", err)
	}

	fmt.Printf(string(keyBytes))

}

func TestGetTxsByCoinPubKey(t *testing.T) {
	cfg := common.Config{
		FullnodeURL: "https://beta-fullnode.incognito.org/fullnode",
	}
	InitIncClient(MainnetStr, cfg)
	fmt.Printf("incClient: %+v\n", incClient)
	findResponseUTXOs("", "", "", 0, config)

	// txMap, err := incClient.GetTxs([]string{"3d54bc8e85d318d383e263312ff623885ac4c7456cfab03747dd63d2f2dbc836"},
	// 	true)
	// if err != nil {
	// 	fmt.Printf("Err: %v\n", err)
	// }
	// fmt.Printf("txMap: %+v\n", txMap)
}

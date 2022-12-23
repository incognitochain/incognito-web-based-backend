package interswap

import (
	"fmt"
	"testing"
)

func TestEstimateSwap(t *testing.T) {
	APIEndpoint = "http://51.161.117.193:8898"
	InitSupportedMidTokens("mainnet")

	XMR := "c01e7dc1d1aba995c19b257412340b057f8ad1482ccb6a9bb0adce61afbf05d4"
	USDC := "545ef6e26d4d428b16117523935b6be85ec0a63e8c2afeb0162315eb0ce3d151"
	from := XMR
	to := USDC

	params := &EstimateSwapParam{
		Network:   "bsc",
		Amount:    "1",
		FromToken: from,
		ToToken:   to,
		Slippage:  "0.5",
	}
	res, err := EstimateSwap(params)
	fmt.Printf("Res: %+v\n", res)
	fmt.Printf("err: %+v\n", err)
}

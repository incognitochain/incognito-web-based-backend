package interswap

import (
	"fmt"
	"math"
	"testing"
)

func TestCallEstimateSwap(t *testing.T) {
	// from := "3ee31eba6376fc16cadb52c8765f20b6ebff92c0b1c5ab5fc78c8c25703bb19e"
	// to := "545ef6e26d4d428b16117523935b6be85ec0a63e8c2afeb0162315eb0ce3d151"

	// params := &EstimateSwapParam{
	// 	Network:   "inc",
	// 	Amount:    "1",
	// 	FromToken: from,
	// 	ToToken:   to,
	// 	Slippage:  "0.5",
	// }
	// res, err := CallEstimateSwap(params)
	// fmt.Printf("Res: %+v\n", res)
	// fmt.Printf("err: %+v\n", err)

	num, err := strToFloat64("0.01123")
	fmt.Printf("num: %v\n", num)
	fmt.Printf("err: %v\n", err)

	tmp := uint64(float64(num) * float64(math.Pow(10, float64(9))))
	fmt.Printf("tmp: %v\n", tmp)

}

// func SendSlackAlert

package papps

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type UniswapDecodeData struct {
	TokenIn           common.Address `json:"tokenIn"`
	TokenOut          common.Address `json:"tokenOut"`
	Fee               *big.Int       `json:"fee"`
	Recipient         common.Address `json:"recipient"`
	AmountIn          *big.Int       `json:"amountIn"`
	AmountOutMinimum  *big.Int       `json:"amountOutMinimum"`
	SqrtPriceLimitX96 *big.Int       `json:"sqrtPriceLimitX96"`
	Path              []byte         `json:"path"`
}

type PancakeDecodeData struct {
	AmountOutMin *big.Int         `json:"amountOutMin"`
	Deadline     *big.Int         `json:"deadline"`
	Path         []common.Address `json:"path"`
	SrcQty       *big.Int         `json:"srcQty"`
}

type CurveDecodeData struct {
	Amount    *big.Int       `json:"amount"`
	MinAmount *big.Int       `json:"minAmount"`
	I         *big.Int       `json:"i"`
	J         *big.Int       `json:"j"`
	CurvePool common.Address `json:"curvePool"`
}

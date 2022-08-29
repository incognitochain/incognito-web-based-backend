package papps

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestCreateCalldata(t *testing.T) {

	token1 := common.Address{}
	token2 := common.Address{}
	err := token1.UnmarshalText([]byte("0x9c3C9283D3e44854697Cd22D3Faa240Cfb032889"))
	require.Equal(t, nil, err)
	err = token2.UnmarshalText([]byte("0xA6FA4fB5f76172d178d61B04b0ecd319C5d1C0aa"))
	require.Equal(t, nil, err)

	paths := []common.Address{}
	paths = append(paths, token1)
	paths = append(paths, token2)
	recipient := common.Address{}
	err = recipient.UnmarshalText([]byte("0x76318093c374e39B260120EBFCe6aBF7f75c8D28"))
	require.Equal(t, nil, err)
	srcQty := new(big.Int).SetInt64(1000000000000)
	dstQty := new(big.Int).SetInt64(2000000)

	result, err := BuildCallDataUniswap(paths, recipient, []int64{1000000000}, srcQty, dstQty, true)
	require.Equal(t, nil, err)

	t.Logf("result: %s\n", result)
}

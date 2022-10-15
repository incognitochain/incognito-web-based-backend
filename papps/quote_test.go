package papps

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestCreateCalldataUniswap(t *testing.T) {

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

	decode, err := DecodeUniswapCalldata(result)
	require.Equal(t, nil, err)

	t.Logf("decode: %s\n", decode)
}

func TestCreateCalldataPancake(t *testing.T) {

	token1 := common.Address{}
	token2 := common.Address{}
	err := token1.UnmarshalText([]byte("0x78867bbeef44f2326bf8ddd1941a4439382ef2a7"))
	require.Equal(t, nil, err)
	err = token2.UnmarshalText([]byte("0x84b9B910527Ad5C03A9Ca831909E21e236EA7b06"))
	require.Equal(t, nil, err)

	paths := []common.Address{}
	paths = append(paths, token1)
	paths = append(paths, token2)
	srcQty := new(big.Int).SetInt64(10000)
	dstQty := new(big.Int).SetInt64(13)

	result, err := BuildCallDataPancake(paths, srcQty, dstQty, true)
	require.Equal(t, nil, err)

	t.Logf("result: %s\n", result)

	decode, err := DecodePancakeCalldata(result)
	require.Equal(t, nil, err)

	t.Logf("decode: %s\n", decode)
}

func TestDecodeCurve(t *testing.T) {

	decode, err := DecodeCurveCalldata("8c0b65930000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000f40b000000000000000000000000000000000000000000000000000000000000f3f830000000000000000000000001d8b86e3d88cdb2d34688e87e72f388cb541b7c8")
	require.Equal(t, nil, err)

	t.Logf("decode: %s\n", decode)

}

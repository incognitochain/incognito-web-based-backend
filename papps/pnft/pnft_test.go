package pnft

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/incognitochain/bridge-eth/bridge/pnft"
)

func TestCheckNFTOwnerQuicknode(t *testing.T) {
	type args struct {
		OSEndpoint string
		address    string
		assets     map[string][]string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string][]string
		wantErr bool
	}{
		{name: "test", args: args{
			OSEndpoint: "https://quaint-multi-card.discover.quiknode.pro/f7bcb8645ed1039ee4a8be74a9eb27a97d9bece3",
			address:    "0x3FC4053980c04Ea4c517D82AfBBb1ceDBBbaa15b",
			assets:     map[string][]string{"0x65c5493e6d4d7bf2da414571eb87ed547eb0abed": []string{"3693"}},
		}, wantErr: false, want: make(map[string][]string)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckNFTOwnerQuicknode(tt.args.OSEndpoint, tt.args.address, tt.args.assets)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckNFTOwnerQuicknode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CheckNFTOwnerQuicknode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerifyOrderSignature(t *testing.T) {
	trader := new(common.Address)
	trader.UnmarshalText([]byte("3420349203489"))
	ETHHost := "https://eth-goerli.g.alchemy.com/v2/uaJphkFTwcgwaLUWLB8fEen0FqoXVj1N"
	client, _ := ethclient.Dial(ETHHost)

	privKey, _ := crypto.HexToECDSA("15681448451d0a925d17408e6c3f33a3e1b5a60b89ab5096c2381a37ab58e234")
	type args struct {
		order           *pnft.Input
		orderHash       string
		ethClient       *ethclient.Client
		exchangeAddress string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "test1",
			wantErr: false,
			args: args{
				order: &pnft.Input{
					Order: pnft.Order{
						Trader: *trader,
					},
				},
				exchangeAddress: "0x87E5Ffa37503487691c75359401080B1e2FBdE5E",
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, orderHash, err := SignSingle(&tt.args.order.Order, privKey, client)
			if err != nil {
				t.Errorf("SignSingle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := VerifyOrderSignature(input, orderHash, client, tt.args.exchangeAddress); (err != nil) != tt.wantErr {
				t.Errorf("VerifyOrderSignature() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func SignSingle(order *pnft.Order, privKey *ecdsa.PrivateKey, client *ethclient.Client) (*pnft.Input, string, error) {
	orderBytes, _ := json.Marshal(order)
	pnftInst, err := pnft.NewBlurExchange(common.HexToAddress("0x87E5Ffa37503487691c75359401080B1e2FBdE5E"), client)
	if err != nil {
		return nil, "", err
	}
	orderHash := crypto.Keccak256(orderBytes)
	domainSeparator, _ := pnftInst.DOMAINSEPARATOR(nil)
	hashToSign := crypto.Keccak256Hash([]byte("\x19\x01"), domainSeparator[:], orderHash[:])
	signBytes, err := crypto.Sign(hashToSign[:], privKey)
	if err != nil {
		return nil, "", err
	}

	return &pnft.Input{
		Order:       *order,
		V:           signBytes[64] + 27,
		R:           toByte32(signBytes[:32]),
		S:           toByte32(signBytes[32:64]),
		BlockNumber: big.NewInt(0),
	}, hex.EncodeToString(orderHash), nil
}

func toByte32(s []byte) [32]byte {
	a := [32]byte{}
	copy(a[:], s)
	return a
}

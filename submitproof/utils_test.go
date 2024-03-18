package submitproof

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/ethclient"
)

func Test_getETHDepositProof(t *testing.T) {

	client, err := ethclient.Dial("https://eth.llamarpc.com")
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	defer client.Close()

	type args struct {
		evmClient *ethclient.Client
		txHashStr string
	}
	tests := []struct {
		name    string
		args    args
		want    *big.Int
		want1   string
		want2   uint
		want3   []string
		want4   string
		want5   string
		want6   bool
		want7   string
		want8   uint64
		want9   string
		want10  bool
		wantErr bool
	}{
		{
			name: "Test case 1",
			args: args{
				evmClient: client,
				txHashStr: "0x2ec1884f68f0787c3132ba06b7008805966bfadd0e117de191c1731f328faafb",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, got3, got4, got5, got6, got7, got8, got9, got10, err := getETHDepositProof(tt.args.evmClient, tt.args.txHashStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("getETHDepositProof() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getETHDepositProof() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getETHDepositProof() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("getETHDepositProof() got2 = %v, want %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(got3, tt.want3) {
				t.Errorf("getETHDepositProof() got3 = %v, want %v", got3, tt.want3)
			}
			if got4 != tt.want4 {
				t.Errorf("getETHDepositProof() got4 = %v, want %v", got4, tt.want4)
			}
			if got5 != tt.want5 {
				t.Errorf("getETHDepositProof() got5 = %v, want %v", got5, tt.want5)
			}
			if got6 != tt.want6 {
				t.Errorf("getETHDepositProof() got6 = %v, want %v", got6, tt.want6)
			}
			if got7 != tt.want7 {
				t.Errorf("getETHDepositProof() got7 = %v, want %v", got7, tt.want7)
			}
			if got8 != tt.want8 {
				t.Errorf("getETHDepositProof() got8 = %v, want %v", got8, tt.want8)
			}
			if got9 != tt.want9 {
				t.Errorf("getETHDepositProof() got9 = %v, want %v", got9, tt.want9)
			}
			if got10 != tt.want10 {
				t.Errorf("getETHDepositProof() got10 = %v, want %v", got10, tt.want10)
			}
		})
	}
}

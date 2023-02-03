package pnft

import (
	"reflect"
	"testing"
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

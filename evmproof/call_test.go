package evmproof

import (
	"encoding/json"
	"testing"
)

func TestRedpositEvent(t *testing.T) {
	u := "https://testnet.incognito.org/fullnode"
	tx := "317214f2d884e926f65d42b19a4f358994fa8f0fb532c0401e41353da7b8d8fa"
	p, e := GetAndDecodeBurnProofUnifiedToken(u, tx, 0)
	if e != nil {
		t.Fatal(e)
	}
	agrBytes, _ := json.MarshalIndent(p, "", " ")

	t.Logf("Checking proof %v", string(agrBytes))
}

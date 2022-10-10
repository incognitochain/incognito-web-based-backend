package evmproof

import (
	"encoding/json"
	"testing"
)

func TestRedpositEvent(t *testing.T) {
	u := "https://lb-fullnode.incognito.org/fullnode"
	tx := "cbf034678320566417c6fd8b683373ea5a29244082f512d217925c47f60a076e"
	p, e := GetAndDecodeBurnProofUnifiedToken(u, tx, 0)
	if e != nil {
		t.Fatal(e)
	}
	agrBytes, _ := json.MarshalIndent(p, "", " ")

	t.Logf("Checking proof %v", string(agrBytes))
}

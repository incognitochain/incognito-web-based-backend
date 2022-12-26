package evmproof

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/incognitochain/bridge-eth/bridge/incognito_proxy"
	"github.com/incognitochain/bridge-eth/bridge/prveth"
	"github.com/incognitochain/bridge-eth/bridge/vault"
	"github.com/pkg/errors"
)

func Withdraw(v *vault.Vault, auth *bind.TransactOpts, proof *DecodedProof) (*types.Transaction, error) {
	// auth.GasPrice = big.NewInt(20000000000)
	tx, err := v.Withdraw(
		auth,
		proof.Instruction,
		proof.Heights[0],

		proof.InstPaths[0],
		proof.InstPathIsLefts[0],
		proof.InstRoots[0],
		proof.BlkData[0],
		proof.SigIdxs[0],
		proof.SigVs[0],
		proof.SigRs[0],
		proof.SigSs[0],
	)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func ExecuteWithBurnProof(v *vault.Vault, auth *bind.TransactOpts, proof *DecodedProof) (*types.Transaction, error) {
	// auth.GasPrice = big.NewInt(20000000000)
	tx, err := v.ExecuteWithBurnProof(
		auth,
		proof.Instruction,
		proof.Heights[0],

		proof.InstPaths[0],
		proof.InstPathIsLefts[0],
		proof.InstRoots[0],
		proof.BlkData[0],
		proof.SigIdxs[0],
		proof.SigVs[0],
		proof.SigRs[0],
		proof.SigSs[0],
	)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func SubmitBurnProof(v *vault.Vault, auth *bind.TransactOpts, proof *DecodedProof) (*types.Transaction, error) {
	// auth.GasPrice = big.NewInt(20000000000)
	tx, err := v.SubmitBurnProof(
		auth,
		proof.Instruction,
		proof.Heights[0],

		proof.InstPaths[0],
		proof.InstPathIsLefts[0],
		proof.InstRoots[0],
		proof.BlkData[0],
		proof.SigIdxs[0],
		proof.SigVs[0],
		proof.SigRs[0],
		proof.SigSs[0],
	)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func SwapBeacon(inc *incognito_proxy.IncognitoProxy, auth *bind.TransactOpts, proof *DecodedProof) (*types.Transaction, error) {
	// auth.GasPrice = big.NewInt(20000000000)
	tx, err := inc.SwapBeaconCommittee(
		auth,
		proof.Instruction,

		proof.InstPaths[0],
		proof.InstPathIsLefts[0],
		proof.InstRoots[0],
		proof.BlkData[0],
		proof.SigIdxs[0],
		proof.SigVs[0],
		proof.SigRs[0],
		proof.SigSs[0],
	)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func SwapBridge(inc *incognito_proxy.IncognitoProxy, auth *bind.TransactOpts, proof *DecodedProof) (*types.Transaction, error) {
	// auth.GasPrice = big.NewInt(20000000000)
	tx, err := inc.SwapBridgeCommittee(
		auth,
		proof.Instruction,

		proof.InstPaths,
		proof.InstPathIsLefts,
		proof.InstRoots,
		proof.BlkData,
		proof.SigIdxs,
		proof.SigVs,
		proof.SigRs,
		proof.SigSs,
	)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func SubmitMintPRVProof(v *prveth.Prveth, auth *bind.TransactOpts, proof *DecodedProof) (*types.Transaction, error) {
	// auth.GasPrice = big.NewInt(20000000000)
	tx, err := v.Mint(
		auth,
		proof.Instruction,
		proof.Heights[0],

		proof.InstPaths[0],
		proof.InstPathIsLefts[0],
		proof.InstRoots[0],
		proof.BlkData[0],
		proof.SigIdxs[0],
		proof.SigVs[0],
		proof.SigRs[0],
		proof.SigSs[0],
	)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func GetAndDecodeBurnProofUnifiedToken(
	incBridgeHost string,
	txID string,
	dataIndex int,
) (*DecodedProof, error) {
	body, err := getBurnProofUnifiedToken(incBridgeHost, txID, dataIndex)
	if err != nil {
		return nil, err
	}
	if len(body) < 1 {
		return nil, fmt.Errorf("burn proof for deposit to SC not found")
	}

	r := getProofResult{}
	err = json.Unmarshal([]byte(body), &r)
	if err != nil {
		return nil, err
	}
	return decodeProof(&r)
}

func getBurnProofUnifiedToken(
	incBridgeHost string,
	txID string,
	dataIndex int,
) (string, error) {
	if len(txID) == 0 {
		return "", errors.New("the tx invalid!")
	}
	payload := strings.NewReader(fmt.Sprintf(`
	{
		"id": 1,
		"jsonrpc": "1.0",
		"method": "bridgeaggGetBurnProof",
		"params": [
			{
				"TxReqID": "%v",
				"DataIndex": %v
			}
		]
	}
	`, txID, dataIndex))

	req, _ := http.NewRequest("POST", incBridgeHost, payload)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func GetAndDecodeBurnProofV2(
	incBridgeHost string,
	txID string,
	rpcMethod string,
) (*DecodedProof, error) {
	body, err := getBurnProofV2(incBridgeHost, txID, rpcMethod)
	if err != nil {
		return nil, err
	}
	if len(body) < 1 {
		return nil, fmt.Errorf("burn proof for deposit to SC not found")
	}

	r := getProofResult{}
	err = json.Unmarshal([]byte(body), &r)
	if err != nil {
		return nil, err
	}
	return decodeProof(&r)
}

package pdao

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/pdao/governance"
)

const (
	CREATE_PROP  = 1
	VOTE_PROP    = 2
	CANCEL_PROP  = 3
	EXECUTE_PROP = 4
)

func CreateGovernanceOutChainTx(network string, incTxHash string, payload []byte, requestType uint8, config wcommon.Config, pappType int) (*wcommon.ExternalTxStatus, error) {
	var result wcommon.ExternalTxStatus

	// networkID := wcommon.GetNetworkID(network)
	networkInfo, err := database.DBGetBridgeNetworkInfo(network)
	if err != nil {
		return nil, err
	}

	papps, err := database.DBRetrievePAppsByNetwork(network)
	if err != nil {
		return nil, err
	}
	contract := papps.AppContracts["pdao"]
	networkChainId := networkInfo.ChainID

	networkChainIdInt := new(big.Int)
	networkChainIdInt.SetString(networkChainId, 10)

	privKey, _ := crypto.HexToECDSA(config.EVMKey)
	i := 0
retry:
	if i == 10 {
		return nil, errors.New("submit tx outchain failed")
	}
	for _, endpoint := range networkInfo.Endpoints {
		evmClient, err := ethclient.Dial(endpoint)
		if err != nil {
			log.Println(err)
			continue
		}

		gv, err := governance.NewGovernance(common.HexToAddress(contract), evmClient)
		if err != nil {
			log.Println(err)
			continue
		}

		gasPrice, err := evmClient.SuggestGasPrice(context.Background())
		if err != nil {
			log.Println(err)
			continue
		}

		auth, err := bind.NewKeyedTransactorWithChainID(privKey, networkChainIdInt)
		if err != nil {
			log.Println(err)
			continue
		}

		gasPrice = gasPrice.Mul(gasPrice, big.NewInt(11))
		gasPrice = gasPrice.Div(gasPrice, big.NewInt(10))

		auth.GasPrice = gasPrice
		auth.GasLimit = wcommon.EVMGasLimitETH
		result.Type = pappType
		result.Network = network
		result.IncRequestTx = incTxHash

		address, err := wcommon.GetEVMAddress(config.EVMKey)
		if err != nil {
			log.Println(err)
			continue
		}
		account := common.HexToAddress(address)
		pendingNonce, _ := evmClient.PendingNonceAt(context.Background(), account)
		auth.Nonce = new(big.Int).SetUint64(pendingNonce + 1)

		tx, err := submitTxOutChain(auth, requestType, payload, gv)
		if err != nil {
			log.Println(err)
			if strings.Contains(err.Error(), "insufficient funds") {
				return nil, errors.New("submit tx outchain failed err insufficient funds")
			}
			continue
		}
		result.Txhash = tx.Hash().String()
		result.Status = wcommon.StatusPending
		result.Nonce = tx.Nonce()
		break
	}

	if result.Txhash == "" {
		i++
		time.Sleep(2 * time.Second)
		goto retry
	}

	return &result, nil
}

func submitTxOutChain(executor *bind.TransactOpts, submitType uint8, payload []byte, gov *governance.Governance) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	switch submitType {
	case CREATE_PROP:
		var prop wcommon.Proposal
		err = json.Unmarshal(payload, &prop)
		if err != nil {
			return nil, err
		}
		var targets []common.Address
		var values []*big.Int
		var calldatas [][]byte
		for i, _ := range strings.Split(prop.Targets, ",") {
			targets = append(targets, common.HexToAddress(strings.Split(prop.Targets, ",")[i]))
			value, _ := new(big.Int).SetString(strings.Split(prop.Values, ",")[i], 10)
			values = append(values, value)
			calldatas = append(calldatas, common.Hex2Bytes(strings.Split(prop.Calldatas, ",")[i]))
		}
		signature := common.Hex2Bytes(prop.CreatePropSignature)
		if len(signature) != 65 {
			return nil, errors.New("Governance: invalid signature length")
		}
		tx, err = gov.ProposeBySig(
			executor,
			targets, values, calldatas,
			prop.Title,
			signature[64]+27,
			toByte32(signature[:32]),
			toByte32(signature[32:64]),
		)
	case VOTE_PROP:
		var vote wcommon.Vote
		err = json.Unmarshal(payload, &vote)
		if err != nil {
			return nil, err
		}
		propID, _ := new(big.Int).SetString(vote.ProposalID, 10)
		signature := common.Hex2Bytes(vote.PropVoteSignature)
		if len(signature) != 65 {
			return nil, errors.New("Governance: invalid signature length")
		}
		tx, err = gov.CastVoteBySig(
			executor,
			propID,
			vote.Vote,
			signature[64]+27,
			toByte32(signature[:32]),
			toByte32(signature[32:64]),
		)
	case CANCEL_PROP:
		var cancel wcommon.Cancel
		err = json.Unmarshal(payload, &cancel)
		if err != nil {
			return nil, err
		}
		propID, _ := new(big.Int).SetString(cancel.ProposalID, 10)
		signature := common.Hex2Bytes(cancel.CancelSignature)
		if len(signature) != 65 {
			return nil, errors.New("Governance: invalid signature length")
		}
		tx, err = gov.CancelBySig(
			executor,
			propID,
			signature[64]+27,
			toByte32(signature[:32]),
			toByte32(signature[32:64]),
		)
	default:
		return nil, errors.New("invalid submit type")
	}
	return tx, err
}

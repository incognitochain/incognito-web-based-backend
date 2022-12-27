package pdao

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/evmproof"
	"github.com/incognitochain/incognito-web-based-backend/pdao/prvvote"
)

const (
	UNSHIELD_PRV     = 1
	SHIELD_BY_SIGN   = 2
	RESHIELD_BY_SIGN = 3
)

func CreatePRVOutChainTx(network string, incTxHash string, payload []byte, requestType uint8, config wcommon.Config, pappType int) (*wcommon.ExternalTxStatus, error) {
	var result wcommon.ExternalTxStatus

	// networkID := wcommon.GetNetworkID(network)
	networkInfo, err := database.DBGetBridgeNetworkInfo(network)
	if err != nil {
		return nil, err
	}

	// todo thachtb: update query prv contract address
	pappAddress, err := database.DBGetPappVaultData(network, wcommon.ExternalTxTypeSwap)
	if err != nil {
		return nil, err
	}

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

		prv, err := prvvote.NewPrvvote(common.HexToAddress(pappAddress.ContractAddress), evmClient)
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

		tx, err := submitTxPRVVoteOutChain(auth, requestType, payload, prv, config)
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

func submitTxPRVVoteOutChain(executor *bind.TransactOpts, submitType uint8, payload []byte, prv *prvvote.Prvvote, config wcommon.Config) (*types.Transaction, error) {
	var tx *types.Transaction
	var err error
	switch submitType {
	case UNSHIELD_PRV:
		proof, err := evmproof.GetAndDecodeBurnProofV2(config.FullnodeURL, string(payload), "getprverc20burnproof")
		if err != nil {
			return nil, err
		}
		tx, err = prv.Mint(
			executor,
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
	case SHIELD_BY_SIGN:
		var rShield Reshield
		err = json.Unmarshal(payload, &rShield)
		if err != nil {
			return nil, err
		}
		signature := common.Hex2Bytes(rShield.Signature)
		amount, _ := new(big.Int).SetString(rShield.Amount, 10)
		tx, err = prv.BurnBySign(
			executor,
			rShield.IncognitoAddress,
			amount,
			[]byte(strconv.FormatInt(rShield.Timestamp, 10)),
			signature[64]+27,
			toByte32(signature[:32]),
			toByte32(signature[32:64]),
		)
	case RESHIELD_BY_SIGN:
		var rShield Reshield
		err = json.Unmarshal(payload, &rShield)
		if err != nil {
			return nil, err
		}
		signature := common.Hex2Bytes(rShield.Signature)
		tx, err = prv.BurnBySignUnShieldTx(
			executor,
			common.HexToHash(rShield.BurnTxId),
			signature[64]+27,
			toByte32(signature[:32]),
			toByte32(signature[32:64]),
		)
	default:
		return nil, errors.New("invalid submit type")
	}
	return tx, err
}

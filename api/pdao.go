package api

import (
	"errors"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/incognitochain/bridge-eth/common/base58"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/pdao/governance"
	"github.com/incognitochain/incognito-web-based-backend/pdao/prvvote"
	"log"
	"math/big"
	"net/http"
	"strings"
)

const GOVERNANCE_CONTRACT_ADDRESS = "0x74E9a67bf51eaa27999d8D699d3Ae4bAdc8c2Af4"
const PRV_VOTE = "0x89b147db2f49c3bc03b3e737453457bEecb3D572"
const PRV_THRESHOLD = "10000000000"

func CreateNewProposal(c *gin.Context) {
	var req CreatProposal
	userAgent := c.Request.UserAgent()
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	// verify request: request no empty, burn tx must be valid
	if len(req.Calldatas) != len(req.Targets) || len(req.Targets) != len(req.Values) || len(req.Targets) == 0 ||
		len(common.Hex2Bytes(req.Signature)) != 65 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "Invalid proposal data"})
		return
	}

	rawTxBytes, _, err := base58.Base58Check{}.Decode(req.TxRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid txhash")})
		return
	}

	// extract metadata from tx
	// two outputs: 1 burn prv (optional) - 1 pay fee
	mdRaw, isPRVTx, outCoins, txHash, err := extractDataFromRawTx(rawTxBytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	// query eth network info
	networkInfo, err := database.DBGetBridgeNetworkInfo("eth")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	var gv *governance.Governance
	var pv *prvvote.Prvvote
	for _, endpoint := range networkInfo.Endpoints {
		evmClient, err := ethclient.Dial(endpoint)
		if err != nil {
			log.Println(err)
			continue
		}

		gv, err = governance.NewGovernance(common.HexToAddress(GOVERNANCE_CONTRACT_ADDRESS), evmClient)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		pv, err = prvvote.NewPrvvote(common.HexToAddress(PRV_VOTE), evmClient)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
			return
		}

		break
	}

	if gv == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "no endpoint available"})
		return
	}

	// recover address from user's signature
	gvAbi, _ := abi.JSON(strings.NewReader(governance.GovernanceMetaData.ABI))
	propEncode, _ := gvAbi.Pack("BuildSignProposalEncodeAbi", keccak256([]byte("proposal")), req.Targets, req.Values, req.Calldatas, req.Description)
	signData, _ := gv.GetDataSign(nil, keccak256(propEncode[4:]))
	rcAddr, err := crypto.Ecrecover(signData[:], common.Hex2Bytes(req.Signature))
	// todo: compare address recover and address from burning metadata if has
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": "invalid signature"})
		return
	}

	// if total burn prv + current prv balance of recover address from signature must pass threshold
	bal, _ := pv.BalanceOf(nil, common.HexToAddress(hexutil.Encode(rcAddr[12:])))
	var threshold *big.Int
	threshold.SetString(PRV_THRESHOLD, 10)
	if bal.Cmp(threshold) < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("insufficient balance to create prop")})
		return
	}

	// check proposal existed
	propId, _ := gv.HashProposal(nil, req.Targets, req.Values, req.Calldatas, keccak256([]byte(req.Description)))
	prop, _ := gv.Proposals(nil, propId)
	if prop.StartBlock.Cmp(big.NewInt(0)) != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("prop id has created")})
		return
	}

	// check fee

	// update request status

	// store request to DB

	//burntAmount, _ := md.TotalBurningAmount()
	//if valid {
	//	status, err := submitproof.SubmitPappTx(txHash, []byte(req.TxRaw), isPRVTx, feeToken, feeAmount, pfeeAmount, md.BurnTokenID.String(), burntAmount, swapInfo, isUnifiedToken, networkList, req.FeeRefundOTA, req.FeeRefundAddress, userAgent)
	//	if err != nil {
	//		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
	//		return
	//	}
	//	c.JSON(200, gin.H{"Result": map[string]interface{}{"inc_request_tx_status": status}, "feeDiff": feeDiff})
	//	return
	//}
}

func GetPdaoStatus(c *gin.Context) {
	var responseBodyData APIRespond
	_, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").
		SetResult(&responseBodyData).
		Get(config.CoinserviceURL + "/bridge/aggregatestate")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(200, responseBodyData)
}

func keccak256(b ...[]byte) [32]byte {
	h := crypto.Keccak256(b...)
	r := [32]byte{}
	copy(r[:], h)
	return r
}
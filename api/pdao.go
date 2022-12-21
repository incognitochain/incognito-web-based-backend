package api

import (
	"errors"
	"log"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/incognitochain/bridge-eth/common/base58"
	"github.com/incognitochain/go-incognito-sdk-v2/metadata/bridge"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/pdao/governance"
	"github.com/incognitochain/incognito-web-based-backend/pdao/prvvote"
	"github.com/incognitochain/incognito-web-based-backend/submitproof"
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
	propEncode, _ := gvAbi.Pack("BuildSignProposalEncodeAbi", keccak256([]byte("proposal")), req.Targets, req.Values, req.Calldatas, req.DescriptionLink)
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

	// convert Targets address to hex address:
	var targetsArr []common.Address
	for _, address := range req.Targets {
		//convert
		targetsArr = append(targetsArr, common.HexToAddress(address))
	}

	var valuesArr []*big.Int
	for _, value := range req.Values {
		//convert
		valueBigInt := big.NewInt(0)
		valueBigInt, ok := valueBigInt.SetString(value, 10)

		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("can not convert values to bigInt")})
			return
		}
		valuesArr = append(valuesArr, valueBigInt)
	}

	var calldataArr [][]byte
	for _, calldata := range req.Calldatas {
		calldataArr = append(calldataArr, []byte(calldata))
	}

	// check proposal existed
	propId, _ := gv.HashProposal(nil, targetsArr, valuesArr, calldataArr, keccak256([]byte(req.DescriptionLink)))
	prop, _ := gv.Proposals(nil, propId)
	if prop.StartBlock.Uint64() != 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("prop id has created")})
		return
	}

	// update request status

	// store request to DB

	// todo: submit transaction to CS and update to DB to get next step
	//burntAmount, _ := md.TotalBurningAmount()
	//if valid {

	rawTxBytes, _, err := base58.Base58Check{}.Decode(req.TxRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("invalid txhash")})
		return
	}

	mdRaw, isPRVTx, _, txHash, err := extractDataFromRawTx(rawTxBytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	var md *bridge.BurningRequest
	md, ok := mdRaw.(*bridge.BurningRequest)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid metadata type"})
		return
	}
	var burnTokenInfo *wcommon.TokenInfo
	var burntAmount uint64
	isUnifiedToken := false
	networkList := []string{}
	tokenID := ""
	uTokenID := ""
	externalAddr := ""
	returnOTA := ""
	if md != nil {
		burnTokenInfo, err = getTokenInfo(md.TokenID.String())
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": errors.New("not supported token")})
			return
		}
		tokenID = burnTokenInfo.TokenID
		uTokenID = burnTokenInfo.TokenID
		burntAmount = md.BurningAmount
		externalAddr = md.RemoteAddress
		// todo: update sdk to get returnOTA
	}

	feeToken := "0000000000000000000000000000000000000000000000000000000000000004"
	feeAmount := 0
	pfeeAmount := 0

	status, err := submitproof.SubmitUnshieldTx(txHash, []byte(req.TxRaw), isPRVTx, feeToken, uint64(feeAmount), uint64(pfeeAmount), tokenID, uTokenID, burntAmount, isUnifiedToken, externalAddr, networkList, returnOTA, "", userAgent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"Result": map[string]interface{}{"inc_request_tx_status": status}})

	return
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

package api

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/incognitochain/bridge-eth/common/base58"
	"github.com/incognitochain/go-incognito-sdk-v2/coin"
	scommon "github.com/incognitochain/go-incognito-sdk-v2/common"
	wcrypto "github.com/incognitochain/go-incognito-sdk-v2/crypto"
	"github.com/incognitochain/go-incognito-sdk-v2/metadata"
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

func APIPDaoFeeEstimate(c *gin.Context) {

	feeAmount, err := estimatePDaoFee()
	if err != nil {
		fmt.Println("estimatePDaoFee", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"Result": feeAmount})
}

func APIPDaoCreateNewProposal(c *gin.Context) {
	var req CreatProposalReq
	userAgent := c.Request.UserAgent()
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	// query eth network info
	networkInfo, err := database.DBGetBridgeNetworkInfo(wcommon.NETWORK_ETH)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"Error": err.Error()})
		return
	}

	log.Println("check network ok!")

	if len(req.Calldatas) > 0 && len(req.Values) > 0 && len(req.Signatures) > 0 {

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

		rcAddr, err := crypto.Ecrecover(signData[:], common.Hex2Bytes(req.CreatePropSignature))
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
			c.JSON(http.StatusBadRequest, gin.H{"Error": "insufficient balance to create prop"})
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
		propId, _ := gv.HashProposal(nil, targetsArr, valuesArr, calldataArr, keccak256([]byte(req.Description)))
		prop, _ := gv.Proposals(nil, propId)
		if prop.StartBlock.Uint64() != 0 {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "prop id has created"})
			return
		}
	}

	// check valid info:
	var feeAmount uint64
	var pfeeAmount uint64 // 0.3% no care
	var feeToken string

	var requireFee uint64
	var requireFeeToken string
	var externalAddr string

	var burntAmount uint64

	feeDiff := int64(-1)

	rawTxBytes, _, err := base58.Base58Check{}.Decode(req.TxRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid txhash"})
		return
	}

	mdRaw, isPRVTx, outCoins, txHash, err := extractDataFromRawTx(rawTxBytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	var md *metadata.BurningPRVRequest
	md, ok := mdRaw.(*metadata.BurningPRVRequest)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid metadata type"})
		return
	}
	var burnTokenInfo *wcommon.TokenInfo

	isUnifiedToken := false
	networkList := []string{}

	tokenID := ""
	uTokenID := ""
	returnOTA := ""

	if md != nil {
		burnTokenInfo, err = getTokenInfo(md.TokenID.String())
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "not supported token"})
			return
		}
		tokenID = burnTokenInfo.TokenID
		uTokenID = burnTokenInfo.TokenID

		burntAmount = md.BurningAmount

		externalAddr = md.RemoteAddress
		// todo: update sdk to get returnOTA
	}

	// verify fee eth (UT):
	for _, cn := range outCoins {
		feeCoin, rK := cn.DoesCoinBelongToKeySet(&incFeeKeySet.KeySet)
		if feeCoin {
			if cn.GetAssetTag() == nil {
				feeToken = scommon.PRVCoinID.String()
			} else {
				assetTag := cn.GetAssetTag()
				blinder, err := coin.ComputeAssetTagBlinder(rK)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid tx err:" + err.Error()})
					return
				}
				rawAssetTag := new(wcrypto.Point).Sub(
					assetTag,
					new(wcrypto.Point).ScalarMult(wcrypto.PedCom.G[coin.PedersenRandomnessIndex], blinder),
				)
				_ = rawAssetTag
				feeToken = burnTokenInfo.TokenID
			}

			coin, err := cn.Decrypt(&incFeeKeySet.KeySet)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid tx err:" + err.Error()})
			}
			feeAmount = coin.GetValue()
		}
	}

	// get fee info from estFee function for checking:
	feeDao, err := estimatePDaoFee()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid tx err:" + err.Error()})
		return
	}

	requireFee = feeDao.FeeAmount
	requireFeeToken = feeDao.TokenID

	// check token fee:
	if feeToken != requireFeeToken {
		c.JSON(http.StatusBadRequest, gin.H{"Error": fmt.Sprintf("invalid fee token, fee token can't be %v must be %v ", feeToken, requireFeeToken)})
		return
	}

	// feeDiff >= 5%
	feeDiff = int64(feeAmount) - int64(feeDao.FeeAmount)
	if feeDiff < 0 {
		feeDiffFloat := math.Abs(float64(feeDiff))
		diffPercent := feeDiffFloat / float64(feeDao.FeeAmount) * 100
		if diffPercent > wcommon.PercentFeeDiff {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid fee amount, fee amount must be at least: " + fmt.Sprintf("%v", requireFee)})
			return
		}
	}

	status, err := submitproof.SubmitUnshieldTx(txHash, []byte(req.TxRaw), isPRVTx, feeToken, uint64(feeAmount), pfeeAmount, tokenID, uTokenID, burntAmount, isUnifiedToken, externalAddr, networkList, returnOTA, "", userAgent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	log.Println("Begin store db!")
	// store request to DB
	proposal := &wcommon.Proposal{
		SubmitBurnTx:        req.Txhash,
		SubmitProposalTx:    "",
		Status:              wcommon.StatusSubmitting,
		ProposalID:          "",
		Proposer:            "",
		Targets:             strings.Join(req.Targets, ","),
		Values:              strings.Join(req.Values, ","),
		Signatures:          strings.Join(req.Signatures, ","),
		Calldatas:           strings.Join(req.Calldatas, ","),
		CreatePropSignature: req.CreatePropSignature,
		PropVoteSignature:   req.PropVoteSignature,
		ReShieldSignature:   req.ReShieldSignature,
		Description:         req.Description,
		Title:               req.Title,
	}
	// insert db
	if err = database.DBInsertProposalTable(proposal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	log.Println("Insert db success!")

	c.JSON(200, gin.H{"Result": map[string]interface{}{"inc_request_tx_status": status}, "feeDiff": feeDiff})
	return

}

func APIPDaoListProposal(c *gin.Context) {
	c.JSON(200, gin.H{"Result": database.DBListProposalTable()})
}
func APIPDaoDetailProposal(c *gin.Context) {
	p, err := database.DBgetProposalTable(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"Result": p})
}

func APIPDaoVoting(c *gin.Context) {
	var req SubmitVoteReq
	userAgent := c.Request.UserAgent()
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	log.Println("Begin store db!")
	// store request to DB
	vote := &wcommon.Vote{
		SubmitBurnTx: req.Txhash,
		Status:       wcommon.StatusSubmitting,
		ProposalID:   req.ProposalID,

		PropVoteSignature: req.PropVoteSignature,
		ReShieldSignature: req.ReShieldSignature,

		Vote: req.Vote,
	}
	// insert db
	if err = database.DBInsertVoteTable(vote); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	log.Println("Insert the voting to db successful!")

	rawTxBytes, _, err := base58.Base58Check{}.Decode(req.TxRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "invalid txhash"})
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

	feeToken := wcommon.ETH_UT_TOKEN_MAINNET
	if config.NetworkID == "testnet" {
		feeToken = wcommon.ETH_UT_TOKEN_TESTNET
	}

	feeAmount := 0
	pfeeAmount := 0

	status, err := submitproof.SubmitUnshieldTx(txHash, []byte(req.TxRaw), isPRVTx, feeToken, uint64(feeAmount), uint64(pfeeAmount), tokenID, uTokenID, burntAmount, isUnifiedToken, externalAddr, networkList, returnOTA, "", userAgent)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"Result": map[string]interface{}{"inc_request_tx_status": status}})

	return

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

// todo move this to lib folder ...
func keccak256(b ...[]byte) [32]byte {
	h := crypto.Keccak256(b...)
	r := [32]byte{}
	copy(r[:], h)
	return r
}

// util function
func estimatePDaoFee() (*PDaoNetworkFeeResp, error) {

	networkFees, err := database.DBRetrieveFeeTable()
	if err != nil {
		fmt.Println("DBRetrieveFeeTable", err.Error())
		return nil, err
	}

	gasPrice := networkFees.GasPrice[wcommon.NETWORK_BSC]

	gasFee := (PDAO_CREATE_PROPOSAL_GAS_LIMIT * gasPrice)

	feeAddress := ""

	feeAddressShardID := byte(0)
	if incFeeKeySet != nil {
		feeAddress, err = incFeeKeySet.GetPaymentAddress()
		if err != nil {
			return nil, err
		}
		_, feeAddressShardID = scommon.GetShardIDsFromPublicKey(incFeeKeySet.KeySet.PaymentAddress.Pk)
	}

	ethToken := wcommon.ETH_UT_TOKEN_MAINNET
	if config.NetworkID == "testnet" {
		ethToken = wcommon.ETH_UT_TOKEN_TESTNET
	}

	feeAmountEth := ConvertNanoAmountOutChainToIncognitoNanoTokenAmountString(fmt.Sprintf("%v", gasFee), 18, 9)

	ethTokenInfo, err := getTokenInfo(ethToken)
	if err != nil {
		fmt.Println("getTokenInfo", err.Error())
		return nil, err
	}

	// for now, get PRV fee, will be remove when we have new update....
	privacyFee := uint64(float64(feeAmountEth) * ethTokenInfo.PricePrv)
	fmt.Println("PRV Fee =================> ", privacyFee)

	feeToken := wcommon.PRV_TOKEN
	feeAmount := privacyFee

	return &PDaoNetworkFeeResp{
		FeeAddress:        feeAddress,
		FeeAddressShardID: int(feeAddressShardID),
		TokenID:           feeToken,
		FeeAmount:         feeAmount,
	}, nil

}

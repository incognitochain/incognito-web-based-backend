package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/incognitochain/incognito-web-based-backend/common"
	wcommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"go.mongodb.org/mongo-driver/mongo"
)

func APIGetStatusByShieldService(c *gin.Context) {
	pyd := c.Query("paymentaddress")
	shieldType := c.Query("type")

	var responseBodyData struct {
		Result []HistoryAddressResp `json:"Result"`
		Error  *struct {
			Code    int
			Message string
		} `json:"Error"`
	}

	var requestBody struct {
		WalletAddress       string
		PrivacyTokenAddress string
	}
	requestBody.WalletAddress = pyd
retry:
	re, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/json").SetHeader("Authorization", "Bearer "+usa.token).SetBody(requestBody).
		Post(config.ShieldService + "/eta/history")
	if err != nil {
		c.JSON(400, gin.H{"Error": err.Error()})
		return
	}

	err = json.Unmarshal(re.Body(), &responseBodyData)
	if err != nil {
		c.JSON(400, gin.H{"Error": err.Error()})
		return
	}

	if responseBodyData.Error != nil {
		if responseBodyData.Error.Code != 401 {
			c.JSON(400, gin.H{"Error": responseBodyData.Error})
			return
		} else {
			err = requestUSAToken(config.ShieldService)
			if err != nil {
				c.JSON(400, gin.H{"Error": err.Error()})
				return
			}
			goto retry
		}
	}

	filteredHistory := []HistoryAddressResp{}
	if shieldType == "unshield" {
		for _, v := range responseBodyData.Result {
			// 2 == unshield
			if v.AddressType == 2 {
				filteredHistory = append(filteredHistory, v)
			}
		}
	} else {
		for _, v := range responseBodyData.Result {
			// 1 == shield
			if v.AddressType == 1 {
				filteredHistory = append(filteredHistory, v)
			}
		}
	}

	resp := struct {
		Result []HistoryAddressResp
		Error  interface{}
	}{filteredHistory, nil}

	c.JSON(200, resp)
}

func APIGetFailedShieldTx(c *gin.Context) {

}

func APIGetPendingShieldTx(c *gin.Context) {

}

func APIGetUnshieldStatus(c *gin.Context) {

}

func APIGetShieldStatus(c *gin.Context) {
	var req SubmitTxListRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
}

func APIGetSwapTxStatus(c *gin.Context) {
	var req SubmitTxListRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}
	result := make(map[string]interface{})
	for _, txHash := range req.TxList {
		statusResult := checkPappTxSwapStatus(txHash)
		if len(statusResult) == 0 {
			statusResult["error"] = "tx not found"
			result[txHash] = statusResult
		} else {
			result[txHash] = statusResult
		}
	}
	c.JSON(200, gin.H{"Result": result})
}

func checkPappTxSwapStatus(txhash string) map[string]interface{} {
	result := make(map[string]interface{})
	data, err := database.DBGetPappTxData(txhash)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			result["error"] = err.Error()
		}
		return result
	}

	result["inc_request_tx_status"] = data.Status
	if data.Status != common.StatusAccepted {
		if data.Error != "" {
			result["error"] = data.Error
		}
	} else {
		networkList := []interface{}{}
		for _, network := range data.Networks {
			networkResult := make(map[string]interface{})
			networkResult["network"] = network
			outchainTx, err := database.DBGetExternalTxByIncTx(txhash, network)
			if err != nil {
				if err != mongo.ErrNoDocuments {
					networkResult["error"] = err.Error()
				} else {
					networkResult["swap_tx_status"] = common.StatusSubmitting
				}
				networkList = append(networkList, networkResult)
				continue
			}
			networkResult["swap_tx_status"] = outchainTx.Status
			networkResult["swap_tx"] = outchainTx.Txhash
			if outchainTx.Error != "" {
				networkResult["swap_err"] = outchainTx.Error
			}
			if outchainTx.Status == common.StatusAccepted && outchainTx.OtherInfo != "" {
				var outchainTxResult wcommon.ExternalTxSwapResult
				err = json.Unmarshal([]byte(outchainTx.OtherInfo), &outchainTxResult)
				if err != nil {
					networkResult["error"] = err.Error()
					networkList = append(networkList, networkResult)
					continue
				}
				if outchainTxResult.IsReverted {
					networkResult["swap_outcome"] = "reverted"
				} else {
					networkResult["swap_outcome"] = "success"
				}
				networkResult["is_redeposit"] = outchainTxResult.IsRedeposit
				if outchainTxResult.IsFailed {
					networkResult["swap_outcome"] = "failed"
				}
				if outchainTxResult.IsRedeposit {
					networkID := wcommon.GetNetworkID(network)
					redepositTx, err := database.DBGetShieldTxByExternalTx(outchainTx.Txhash, networkID)
					if err != nil {
						if err != mongo.ErrNoDocuments {
							networkResult["error"] = err.Error()
						} else {
							networkResult["redeposit_status"] = common.StatusSubmitting
						}
						networkList = append(networkList, networkResult)
						continue
					}
					networkResult["redeposit_status"] = redepositTx.Status
					networkResult["redeposit_inctx"] = redepositTx.IncTx
					if data.BurntToken == "" {
						networkResult["swap_outcome"] = "unvailable"
					} else {
						if redepositTx.UTokenID == data.BurntToken {
							networkResult["swap_outcome"] = "reverted"
						} else {
							if redepositTx.UTokenID == "" {
								networkResult["swap_outcome"] = "pending"
							} else {
								networkResult["swap_outcome"] = "success"
							}
						}
					}
				}
			}
			networkList = append(networkList, networkResult)
		}
		result["network_result"] = networkList
	}
	return result
}

// 0xa807b5bb000000000000000000000000a6fa4fb5f76172d178d61b04b0ecd319c5d1c0aa00000000000000000000000000000000000000000000000000005af30e8bdd8000000000000000000000000000000000000000000000000000000000000000a0307834386430343936306639323634353639653261626331323534356361353100000000000000000000000000000000000000000000000000000000000001600000000000000000000000000000000000000000000000000000000000000094313273665637566f3237527a3361543463326b79695470767a6977586a7669514d4d72703567734666757041766f4476654868514c756e4157767154616f343644534559706e624d7047597875633461394b475537427070504d39755a7466564371504151313857745045696a734c6d5978564c314d5757446767445a4866526d6874786d56696a6164436a58797237694339580000000000000000000000000000000000000000000000000000000000000000000000000000000000000041f2e35d271efe09879a8128ea18a83acddac22ce69c76e6e6604be49346a2c27824b6a2d20cb5e8e55b2eca729a6c4e06b815970df9cbd319b39121ed6d35f34a0000000000000000000000000000000000000000000000000000000000000000

// 0x3ed1b37600000000000000000000000000000000000000000000000000000000000001400000000000000000000000000000000000000000000000000000000000519aa500000000000000000000000000000000000000000000000000000000000003c00000000000000000000000000000000000000000000000000000000000000460543a436f135cec22ead354cfbaa15336006a3ecb5d3376a56d0b7c3e8e984aba68cae1c062302fe28526bdddc455822e55b72f255b999b71e1c681604997d39f000000000000000000000000000000000000000000000000000000000000050000000000000000000000000000000000000000000000000000000000000005a0000000000000000000000000000000000000000000000000000000000000064000000000000000000000000000000000000000000000000000000000000006e0000000000000000000000000000000000000000000000000000000000000024c9e01030000000000000000000000000000000000000000000000000000000000000000000000000000000000000000b806dc43e5494845795ca75ba49406cd0ffea2e000000000000000000000000000000000000000000000000000071afd498d000034337c84ff64682a4cbb6586e31a0513052630cbd5a0b9d9a0c10f6a598405ff0000000000000000000000003813e82e6f7098b9583fc0f33a962d02018b680300000000000000000000000000000000000000000000000000000000000000000463d20e1a6733f569c11235c0d13452785ab51d1aa464cf48fb87f651adf344c0ce6cbcc994b04bc514d42f92b45f29fc31c2560b410b2ad321755f722c073f490000000a348c306f823f032d9451a24708bde7ca95d69c216fb374e1490224e11be8049ec8dc75e600000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000008000000000000000000000000076318093c374e39b260120ebfce6abf7f75c8d2800000000000000000000000000000000000000000000000000071afd498cffff000000000000000000000000000000000000000000000000000000000000000b000000000000000000000000000000000000000000000000000000000000002b9c3c9283d3e44854697cd22d3faa240cfb03288904c6a83813e82e6f7098b9583fc0f33a962d02018b680300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004071d728babb1280fe594ec623df44c480e6243d1af72da71b17f36830b0efd0b39e42507733dd2473e9a33724d101847a31339cfd21f47f90dd4861731de2c005fcd89ec568c55fe46a7bc09034c7ffffbee2bad345ecf90e821797e60cdd5b4400c0d130ed8448f08b24f50659bc68ebbd84674fe07f687b0a22d955cefbea500000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000030000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000001b000000000000000000000000000000000000000000000000000000000000001c000000000000000000000000000000000000000000000000000000000000001b000000000000000000000000000000000000000000000000000000000000001c0000000000000000000000000000000000000000000000000000000000000004976abad0e209fe5a74ac2722ef683717cc26d1e46970fa9ee39272fdb8b5305f1150b3382a2f6598a839f47d3341f21a8bae9d36137bff194b078ae1ca306d79050ec96360e31086e46a191160732c134c88102f73d934619485d9a9b463077510fc5455668265c36fb1170975186d3daaad6700347142dc4ecde112403d12ab000000000000000000000000000000000000000000000000000000000000000417bf1d0b9d5e8273c9294317f81c1c220078bdd5b551c6a4f310979799ec72397d75767faea1781487ce447cfdb70ca00c6c37a33363fb8ffa20f5b6cd4c86036871c1b38db44b68806c070713deba58d1a7b09a3897c01f4e0efe6e279aef2132072fddc250054793b76c6938355f7660962b35d1af915ec573ad5937d4c812

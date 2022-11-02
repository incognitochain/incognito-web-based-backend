package submitproof

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
	"time"

	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/slacknoti"
)

func watchVaultState() {
	for {

		var responseBodyData struct {
			ID     int `json:"Id"`
			Result *struct {
				UnifiedTokenVaults map[string]map[string]struct {
					Amount    uint64 `json:"Amount"`
					NetworkID int    `json:"NetworkID"`
				}
			} `json:"Result"`
			Error *struct {
				Code       int    `json:"Code"`
				Message    string `json:"Message"`
				StackTrace string `json:"StackTrace"`
			} `json:"Error"`
		}
		_, err := restyClient.R().
			EnableTrace().
			SetHeader("Content-Type", "application/json").
			SetResult(&responseBodyData).
			Get(config.CoinserviceURL + "/bridge/aggregatestate")
		if err != nil {
			log.Println(err)
			continue
		}

		if responseBodyData.Error != nil {
			log.Println(responseBodyData.Error)
			continue
		}

		result := make(map[string]map[string]string)

		for uTokenID, vault := range responseBodyData.Result.UnifiedTokenVaults {
			uTkInfo, _ := getTokenInfo(uTokenID)
			result[uTkInfo.Name] = make(map[string]string)
			for cTkID, v := range vault {
				cTkInfo, _ := getTokenInfo(cTkID)
				amount := new(big.Float).SetUint64(v.Amount)
				decimal := new(big.Float).SetFloat64(math.Pow10(-9))
				amountFloat := new(big.Float).Mul(amount, decimal)
				afl64, _ := amountFloat.Float64()
				network := common.GetNetworkName(v.NetworkID)
				result[uTkInfo.Name][cTkInfo.Name+fmt.Sprintf(" (%v)", network)] = fmt.Sprintf("%.2f", afl64)
			}
		}

		resultJson, err := json.MarshalIndent(result, "", "\t")
		if err != nil {
			log.Println(responseBodyData.Error)
			continue
		}
		slackep := os.Getenv("SLACK_VAULT_ALERT")
		go slacknoti.SendWithCustomChannel(string(resultJson), slackep)

		time.Sleep(10 * time.Minute)
	}
}

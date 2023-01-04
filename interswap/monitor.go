package interswap

import (
	"fmt"
	"time"

	beCommon "github.com/incognitochain/incognito-web-based-backend/common"
)

const MinBalance = 400000000

func monitorBalanceISIncKeys(config beCommon.Config) {
	privKeys := config.ISIncPrivKeys

	for {
		for shardID, privateKey := range privKeys {
			balance, err := incClient.GetBalance(privateKey, beCommon.PRV_TOKENID)
			if err != nil {
				continue
			}

			if balance < MinBalance {
				SendSlackAlert(fmt.Sprintf("Interswap Account shardID %v low balance %v. Please top up now!", balance, shardID))
			}
		}

		time.Sleep(10 * time.Minute)
	}
}

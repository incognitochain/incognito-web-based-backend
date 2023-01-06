package interswap

import (
	"fmt"
	"os"

	beCommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/slacknoti"
)

var SlackInfo = os.Getenv("INTERSWAP_SLACK_INFO")
var SlackAlert = os.Getenv("INTERSWAP_SLACK_ALERT")

func SendSlackInfo(msg string) {
	slacknoti.SendWithCustomChannel(msg, SlackInfo)
}

func SendSlackAlert(msg string) {
	slacknoti.SendWithCustomChannel(msg, SlackAlert)
}

func SendSlackSwapInfo(
	interswapID string, userAgent string, status string,
	fromAmt uint64, fromToken string, toAmt uint64, toToken string,
	receiveAmt uint64, receiveToken string,
	config beCommon.Config,
) error {
	uaStr := ParseUserAgent(userAgent)
	tokenIDs := []string{fromToken, toToken}
	if receiveToken != "" {
		tokenIDs = append(tokenIDs, receiveToken)
	}
	tokenInfos, err := getTokensInfo(tokenIDs, config)
	if err != nil || len(tokenInfos) < 2 {
		return err
	}

	fromAmtWithoutDec, err := ConvertUint64ToWithoutDecStr2(fromAmt, uint64(tokenInfos[0].PDecimals))
	if err != nil {
		return err
	}
	toAmtWithoutDec, err := ConvertUint64ToWithoutDecStr2(toAmt, uint64(tokenInfos[1].PDecimals))
	if err != nil {
		return err
	}

	receiveAmtStr := ""
	receiveTokenSymbol := ""
	if receiveAmt > 0 && len(tokenInfos) >= 3 {
		receiveAmtWithoutDec, err := ConvertUint64ToWithoutDecStr2(receiveAmt, uint64(tokenInfos[2].PDecimals))
		if err != nil {
			return err
		}
		receiveAmtStr = receiveAmtWithoutDec
		receiveTokenSymbol = tokenInfos[2].Symbol
	}

	swapAlert := fmt.Sprintf("`[%v | %v]` swap %v 🛰\n SwapID: `%v`\n Requested: `%v %v` to `%v %v` | Received: `%v %v`\n--------------------------------------------------------",
		"Interswap", uaStr, status, interswapID, fromAmtWithoutDec, tokenInfos[0].Symbol, toAmtWithoutDec, tokenInfos[1].Symbol, receiveAmtStr, receiveTokenSymbol)
	SendSlackInfo(swapAlert)
	return nil
}

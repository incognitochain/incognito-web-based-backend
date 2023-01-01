package interswap

import (
	"os"

	"github.com/incognitochain/incognito-web-based-backend/slacknoti"
)

const InterswapSlackChannel = "INTERSWAP_SLACK_ALERT"

var SlackInfo = os.Getenv("INTERSWAP_SLACK_INFO")
var SlackAlert = os.Getenv("INTERSWAP_SLACK_ALERT")

func SendSlackInfo(msg string) {
	slacknoti.SendWithCustomChannel(msg, SlackInfo)
}

func SendSlackAlert(msg string) {
	slacknoti.SendWithCustomChannel(msg, SlackAlert)
}

package interswap

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
	"github.com/incognitochain/incognito-web-based-backend/common"
	beCommon "github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/mileusna/useragent"
)

var BEEndpoint = os.Getenv("BE_ENDPOINT")

var incClient *incclient.IncClient

func InitIncClient(network string, config common.Config) error {
	var err error
	switch network {
	case "mainnet":
		incClient, err = incclient.NewIncClient(config.FullnodeURL, incclient.MainNetETHHost, 2, network)
	default:
		incClient, err = incclient.NewIncClient(config.FullnodeURL, "", 2, network)
	}
	if err != nil {
		return err
	}
	return nil
}

var restyClient = resty.New()

var SupportedMidTokens = []string{}

func InitSupportedMidTokens(network string) error {
	switch network {
	case "mainnet":
		SupportedMidTokens = []string{
			"076a4423fa20922526bd50b0d7b0dc1c593ce16e15ba141ede5fb5a28aa3f229", // USDT
			// "3ee31eba6376fc16cadb52c8765f20b6ebff92c0b1c5ab5fc78c8c25703bb19e", // ETH
		}
	case "testnet":
		SupportedMidTokens = []string{
			"b35756452dc1fa1260513fa121c20c2b516a8645f8d496fa4235274dac0b1b52", // LINK (unified)
			// "3a526c0fa9abfc3e3e37becc52c5c10abbb7897b0534ad17018e766fc6133590", // USDT
			// "b366fa400c36e6bbcf24ac3e99c90406ddc64346ab0b7ba21e159b83d938812d", // ETH
		}
	default:
		return errors.New("Invalid supported network")
	}
	return nil
}

func IsMidTokens(tokenID string) bool {
	for _, item := range SupportedMidTokens {
		if item == tokenID {
			return true
		}
	}
	return false
}

func IsExist(strArr []string, str string) bool {
	for _, item := range strArr {
		if item == str {
			return true
		}
	}
	return false
}

func strToFloat64(str string) (float64, error) {
	if str == "" {
		return 0, nil
	}
	return strconv.ParseFloat(str, 64)
}

func float64ToStr(f float64) string {
	return fmt.Sprintf("%.9f", f)
}

func addStrs(str1, str2 string) (string, error) {
	f1, err := strToFloat64(str1)
	if err != nil {
		return "", err
	}

	f2, err := strToFloat64(str2)
	if err != nil {
		return "", err
	}
	return float64ToStr(f1 + f2), nil
}

func subStrs(str1, str2 string) (string, error) {
	f1, err := strToFloat64(str1)
	if err != nil {
		return "", err
	}

	f2, err := strToFloat64(str2)
	if err != nil {
		return "", err
	}
	if f1 < f2 {
		return "", fmt.Errorf("%v is less than %v", str1, str2)
	}
	return float64ToStr(f1 - f2), nil
}

func divStrs(str1, str2 string) (string, error) {
	f1, err := strToFloat64(str1)
	if err != nil {
		return "", err
	}

	f2, err := strToFloat64(str2)
	if err != nil {
		return "", err
	}

	return float64ToStr(f1 / f2), nil
}

func convertAmountDec(amt uint64, fromToken, toToken string, config common.Config) (uint64, error) {
	tokenInfos, err := getTokensInfo([]string{fromToken, toToken}, config)
	if err != nil {
		return 0, err
	}

	fromTokenInfo := common.TokenInfo{}
	toTokenInfo := common.TokenInfo{}
	for _, info := range tokenInfos {
		if info.TokenID == fromToken {
			fromTokenInfo = info
		} else if info.TokenID == toToken {
			toTokenInfo = info
		}
	}

	if fromTokenInfo.TokenID == "" || toTokenInfo.TokenID == "" {
		return 0, errors.New("Can not get token info")
	}

	tmp := float64(amt) * float64(math.Pow(10, float64(toTokenInfo.PDecimals)))
	tmp = float64(tmp) / float64(math.Pow(10, float64(fromTokenInfo.PDecimals)))
	return uint64(tmp), nil
}

func convertAmountUint64(amt uint64, fromToken, toToken string, config common.Config) (uint64, error) {
	tokenInfos, err := getTokensInfo([]string{fromToken, toToken}, config)
	if err != nil {
		return 0, err
	}

	fromTokenInfo := common.TokenInfo{}
	toTokenInfo := common.TokenInfo{}
	for _, info := range tokenInfos {
		if info.TokenID == fromToken {
			fromTokenInfo = info
		} else if info.TokenID == toToken {
			toTokenInfo = info
		}
	}

	if fromTokenInfo.TokenID == "" || toTokenInfo.TokenID == "" {
		return 0, errors.New("Can not get token info")
	}

	amtInUSD := float64(amt) * fromTokenInfo.ExternalPriceUSD
	amtTo := amtInUSD / toTokenInfo.ExternalPriceUSD

	return uint64(amtTo), nil
}

func convertAmountFromToTokenInfo(amt uint64, fromTokenInfo, toTokenInfo beCommon.TokenInfo) (uint64, error) {
	// tokenInfos, err := getTokensInfo([]string{fromToken, toToken}, config)
	// if err != nil {
	// 	return 0, err
	// }

	// fromTokenInfo := common.TokenInfo{}
	// toTokenInfo := common.TokenInfo{}
	// for _, info := range tokenInfos {
	// 	if info.TokenID == fromToken {
	// 		fromTokenInfo = info
	// 	} else if info.TokenID == toToken {
	// 		toTokenInfo = info
	// 	}
	// }

	if fromTokenInfo.TokenID == "" || toTokenInfo.TokenID == "" {
		return 0, errors.New("Invalid token info")
	}

	amtFloat64, err := ConvertAmountToWithoutDec(amt, uint64(fromTokenInfo.PDecimals))
	if err != nil {
		return 0, err
	}

	amtInUSD := amtFloat64 * fromTokenInfo.ExternalPriceUSD
	amtToFloat64 := amtInUSD / toTokenInfo.ExternalPriceUSD

	return convertToDecAmtWithTokenInfo(amtToFloat64, toTokenInfo)
}

func convertToWithoutDecStr(amt uint64, tokenID string, config common.Config) (string, error) {
	tokenInfo, err := getTokenInfo(tokenID, config)
	if err != nil {
		return "", nil
	}
	tmp := float64(amt) / float64(math.Pow(10, float64(tokenInfo.PDecimals)))
	return float64ToStr(tmp), nil
}

func convertToWithoutDecStrWithTokenInfo(amt uint64, tokenInfo *beCommon.TokenInfo) (string, error) {
	tmp := float64(amt) / float64(math.Pow(10, float64(tokenInfo.PDecimals)))
	return float64ToStr(tmp), nil
}

func convertToDecAmtStr(amt string, tokenID string, config common.Config) (string, error) {
	tokenInfo, err := getTokenInfo(tokenID, config)
	if err != nil {
		return "", nil
	}
	amtTmp, err := strToFloat64(amt)
	if err != nil {
		return "", nil
	}
	tmp := uint64(float64(amtTmp) * float64(math.Pow(10, float64(tokenInfo.PDecimals))))
	return fmt.Sprint(tmp), nil
}

func convertToDecAmtUint64(amt string, tokenID string, config common.Config) (uint64, error) {
	tokenInfo, err := getTokenInfo(tokenID, config)
	if err != nil {
		return 0, nil
	}
	amtFloat64, err := strToFloat64(amt)
	if err != nil {
		return 0, nil
	}

	return convertToDecAmtWithTokenInfo(amtFloat64, *tokenInfo)
}

func convertToDecAmtWithTokenInfo(amt float64, tokenInfo beCommon.TokenInfo) (uint64, error) {
	tmp := uint64(amt * float64(math.Pow(10, float64(tokenInfo.PDecimals))))
	return tmp, nil
}

func ConvertUint64ToWithoutDecStr(amt uint64, tokenID string, config common.Config) (string, error) {
	tokenInfo, err := getTokenInfo(tokenID, config)
	if err != nil {
		return "", nil
	}
	tmp := float64(amt) / float64(math.Pow(10, float64(tokenInfo.PDecimals)))
	return float64ToStr(tmp), nil
}

func ConvertUint64ToWithoutDecStr2(amt uint64, pDecimal uint64) (string, error) {
	tmp := float64(amt) / float64(math.Pow(10, float64(pDecimal)))
	return float64ToStr(tmp), nil
}

func ConvertAmountToWithoutDec(amt uint64, pDecimal uint64) (float64, error) {
	tmp := float64(amt) / float64(math.Pow(10, float64(pDecimal)))
	return tmp, nil
}

func convertAmtExtDecToAmtPDec(amt *big.Int, tokenID string, config common.Config) (uint64, error) {
	tokenInfo, err := getTokenInfo(tokenID, config)
	if err != nil {
		return 0, nil
	}
	tmp := new(big.Int).Mul(amt, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(tokenInfo.PDecimals)), nil))
	tmp = new(big.Int).Div(tmp, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(tokenInfo.Decimals)), nil))

	return tmp.Uint64(), nil
}

func IsUniqueSlices(s []string) bool {
	m := map[string]bool{}
	for _, i := range s {
		if m[i] {
			return false
		}
		m[i] = true
	}
	return true
}

func Has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

// Remove0xPrefix removes 0x prefix (if there) from string
func Remove0xPrefix(str string) string {
	if Has0xPrefix(str) {
		return str[2:]
	}
	return str
}

func ParseUserAgent(userAgent string) string {
	ua := useragent.Parse(userAgent)
	uaStr := ""

	if ua.IsAndroid() {
		uaStr = "mobile-android"
	}
	if ua.IsIOS() {
		uaStr = "mobile-ios"
	}
	if ua.IsEdge() {
		uaStr = "web-edge"
	}
	if ua.IsFirefox() {
		uaStr = "web-firefox"
	}
	if ua.IsChrome() {
		uaStr = "web-chrome"
	}
	if ua.IsOpera() {
		uaStr = "web-opera"
	}

	if uaStr == "" {
		if ua.Mobile || ua.Tablet {
			uaStr = "mobile"
		}
		if ua.Bot {
			uaStr = "bot"
		}
		if ua.Desktop {
			uaStr = "web"
		}
	}
	if uaStr == "" {
		if strings.Contains(userAgent, "okhttp") {
			uaStr = "mobile-android-ish"
		}
		if strings.Contains(userAgent, "CFNetwork") {
			uaStr = "mobile-ios-ish"
		}
	}
	if uaStr == "" {
		uaStr = userAgent
	}
	return uaStr
}

func getTokenInfoWithCache(tokenID string, tokenInfoCaches map[string]*beCommon.TokenInfo, config beCommon.Config) (*beCommon.TokenInfo, map[string]*beCommon.TokenInfo, error) {
	tmp := tokenInfoCaches[tokenID]
	if tmp.TokenID == "" || tmp.TokenID != tokenID {
		tokenInfo, err := getTokenInfo(tokenID, config)
		if err != nil {
			return nil, tokenInfoCaches, err
		}
		tokenInfoCaches[tokenID] = tokenInfo
		return tokenInfo, tokenInfoCaches, nil
	}
	return tmp, tokenInfoCaches, nil
}

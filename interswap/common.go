package interswap

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/incognitochain/go-incognito-sdk-v2/incclient"
)

var incClient *incclient.IncClient

func InitIncClient(network string) error {
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
			"3ee31eba6376fc16cadb52c8765f20b6ebff92c0b1c5ab5fc78c8c25703bb19e", // ETH
		}
	case "testnet":
		SupportedMidTokens = []string{
			"3a526c0fa9abfc3e3e37becc52c5c10abbb7897b0534ad17018e766fc6133590", // USDT
			"b366fa400c36e6bbcf24ac3e99c90406ddc64346ab0b7ba21e159b83d938812d", // ETH
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

func strToFloat64(str string) (float64, error) {
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

// TODO:
func convertAmountUint64(amt uint64, fromToken, toToken string) uint64 {
	return amt
}


// TODO:
func convertAmountStr(amt string, fromToken, toToken string) string {
	return amt
}

func convertToWithoutDecStr(amt uint64, tokenID string) string {
	tmp := float64(amt) / float64(math.Pow(10, DefaultDecimal))
	return float64ToStr(tmp)
}

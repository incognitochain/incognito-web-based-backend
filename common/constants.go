package common

var DefaultConfig = Config{
	Port:           9898,
	CoinserviceURL: "http://51.161.117.193:8096",
	NetworkID:      "testnet-2",
	FullnodeURL:    "https://testnet.incognito.org/fullnode",
	ShieldService:  "https://staging-api-service.incognito.org",
}

const (
	MODE_TXSUBMITWATCHER = "submitwatcher"
	MODE_TXSUBMITWORKER  = "submitworker"
	MODE_API             = "api"
	MODE_FEEESTIMATOR    = "feeestimator"
	MODE_UNSHIELDWATCHER = "unshieldwatcher"
)

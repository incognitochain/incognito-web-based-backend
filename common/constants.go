package common

var DefaultConfig = Config{
	Port:           9000,
	DatabaseURLs:   []string{"127.0.0.1:6379"},
	CoinserviceURL: "http://51.161.117.193:8096",
	FullnodeURL:    "https://testnet.incognito.org/fullnode",
	ShieldService:  "https://staging-api-service.incognito.org",
}

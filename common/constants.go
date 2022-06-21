package common

var DefaultConfig = Config{
	Port:           9000,
	DatabaseURL:    "mongodb://localhost:27017",
	CoinserviceURL: "http://localhost:9000",
	FullnodeURL:    "http://localhost:9000",
	CrossChainURL:  "http://localhost:9000",
}

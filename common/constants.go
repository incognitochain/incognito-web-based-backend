package common

var DefaultConfig = Config{
	Port:           9000,
	DatabaseURL:    "mongodb://localhost:27017",
	CoinserviceURL: "http://51.161.117.193:6002",
	FullnodeURL:    "http://51.161.117.193:11334",
	ShieldService:  "https://dev-api-service.incognito.org",
}

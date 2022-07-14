package common

type Config struct {
	Port           int
	DatabaseURLs   []string
	CoinserviceURL string
	FullnodeURL    string
	ShieldService  string
}

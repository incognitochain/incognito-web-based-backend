package api

import (
	"encoding/json"
	"sync"
	"time"
)

type UnshieldServiceAuth struct {
	token    string
	expireAt time.Time
	lock     sync.Mutex
}

var usa UnshieldServiceAuth

func requestUSAToken(endpoint string) error {
	usa.lock.Lock()
	defer usa.lock.Unlock()

	if usa.token == "" || time.Now().Unix() > usa.expireAt.Unix() {
		type Respond struct {
			Result struct {
				Token   string
				Expired string
			}
			Error interface{}
		}

		var responseBodyData Respond
		res, err := restyClient.R().
			EnableTrace().
			SetHeader("Content-Type", "application/json").
			Get(config.ShieldService + "/auth/new-token")
		if err != nil {
			return err
		}
		err = json.Unmarshal(res.Body(), &responseBodyData)
		if err != nil {
			return err
		}
	}
	return nil
}

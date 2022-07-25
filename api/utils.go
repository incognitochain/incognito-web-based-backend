package api

import (
	"github.com/mongodb/mongo-tools/common/json"
)

func genRPCBody(method string, params []interface{}) interface{} {
	type RPC struct {
		ID      int           `json:"id"`
		JsonRPC string        `json:"jsonrpc"`
		Method  string        `json:"method"`
		Params  []interface{} `json:"params"`
	}

	req := RPC{
		ID:      1,
		JsonRPC: "1.0",
		Method:  method,
		Params:  params,
	}
	return req
}

func VerifyCaptcha(clientCaptcha string, secret string) (bool, error) {

	data := make(map[string]string)
	data["response"] = clientCaptcha
	data["secret"] = secret

	re, err := restyClient.R().
		EnableTrace().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").SetHeader("Authorization", "Bearer "+usa.token).SetFormData(data).
		Post("https://hcaptcha.com/siteverify")
	if err != nil {
		return false, err
	}

	var responseBodyData struct {
		Success bool `json:"success"`
	}

	err = json.Unmarshal(re.Body(), &responseBodyData)
	if err != nil {
		return false, err
	}

	return responseBodyData.Success, nil
}

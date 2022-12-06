package api

// type UnshieldServiceAuth struct {
// 	token    string
// 	expireAt time.Time
// 	lock     sync.Mutex
// }

// var usa UnshieldServiceAuth

// func requestUSAToken(endpoint string) error {
// 	usa.lock.Lock()
// 	defer usa.lock.Unlock()

// 	if usa.token == "" || time.Now().Unix() > usa.expireAt.Unix() {
// 		type Respond struct {
// 			Result struct {
// 				Token   string
// 				Expired string
// 			}
// 			Error interface{}
// 		}
// 		var reqBody = struct {
// 			DeviceID    string
// 			DeviceToken string
// 		}{
// 			"be123", "be123",
// 		}
// 		var responseBodyData Respond
// 		res, err := restyClient.R().
// 			EnableTrace().
// 			SetHeader("Content-Type", "application/json").SetBody(reqBody).
// 			Post(config.ShieldService + "/auth/new-token")
// 		if err != nil {
// 			return err
// 		}
// 		err = json.Unmarshal(res.Body(), &responseBodyData)
// 		if err != nil {
// 			return err
// 		}
// 		if responseBodyData.Error != nil {
// 			return errors.New("can't gen auth token")
// 		}
// 		usa.expireAt, err = time.Parse(time.RFC3339, responseBodyData.Result.Expired)
// 		if err != nil {
// 			return err
// 		}
// 		usa.token = responseBodyData.Result.Token
// 		log.Println("usa.expireAt", usa.expireAt, responseBodyData.Result.Expired)
// 	}
// 	return nil
// }

package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"testing"

	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

type PappSupportedTokenData struct {
	TokenID    string `json:"tokenid" bson:"tokenid"`
	ContractID string `json:"contractid" bson:"contractid"`
	Verify     bool   `json:"verify" bson:"verify"`
}

func TestBuildUniswapTokenList(t *testing.T) {
	type TokenStruct struct {
		Address string `json:"address"`
		ChainID int    `json:"chainId"`
	}

	contractList := []TokenStruct{}

	data, err := ioutil.ReadFile("./uniswap_eth.json")
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			log.Fatalln(err)
		} else {
			log.Fatalln(err)
		}
	}

	err = json.Unmarshal(data, &contractList)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("len(contractList)", len(contractList))
	service_endpoint := "https://api-coinservice.incognito.org"
	dbname := "data"
	dbendpoint := ""

	_ = dbname
	_ = dbendpoint

	chainTokenList, err := retrieveTokenList(service_endpoint)
	if err != nil {
		log.Fatalln(err)
	}
	chainTokenMap := make(map[string]common.TokenInfo)

	for _, v := range chainTokenList {
		if v.CurrencyType == common.ERC20 || v.CurrencyType == common.ETH && v.Verified {
			chainTokenMap[strings.ToLower(v.ContractID)] = v
		}
	}

	result := []PappSupportedTokenData{}

	for _, v := range contractList {
		if v.ChainID != 1 {
			continue
		}
		if tk, ok := chainTokenMap[strings.ToLower(v.Address)]; ok {
			token := PappSupportedTokenData{
				Verify:     true,
				ContractID: v.Address,
				TokenID:    tk.TokenID,
			}
			result = append(result, token)
		}
	}

	log.Printf("found %v tokens\n", len(result))

	savePappSupportedToken(result)

	// err = addTokenToDB(result, dbendpoint, dbname)
	// if err != nil {
	// 	log.Fatalln(err)
	// }
}

func savePappSupportedToken(token []PappSupportedTokenData) error {
	agrBytes, _ := json.MarshalIndent(token, "", "\t")
	writeToFile(fmt.Sprintln(string(agrBytes)), "tokens.json")
	return nil
}

func addTokenToDB(list []common.PappSupportedTokenData, dbendpoint string, dbname string) error {
	wc := writeconcern.New(writeconcern.W(1), writeconcern.J(true))
	err := mgm.SetDefaultConfig(nil, dbname, options.Client().ApplyURI(dbendpoint).SetWriteConcern(wc))
	if err != nil {
		return err
	}

	docs := []interface{}{}
	for _, data := range list {
		docs = append(docs, data)
	}

	_, err = mgm.Coll(&common.PappSupportedTokenData{}).InsertMany(context.Background(), docs)
	if err != nil {
		return err
	}
	return nil
}

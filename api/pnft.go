package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
	"github.com/incognitochain/incognito-web-based-backend/papps/pnft"
)

func APIPnftGetNftsFromAddress(c *gin.Context) {

	address, _ := c.GetQuery("address")

	log.Println("address: ", address)

	if len(address) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "address is empty"})
		return
	}

	address = strings.ToLower(address)

	// get them from db:
	result, _ := database.DBPNftGetListNftCacheTableByAddress(address)
	response := ""
	var err error

	if result != nil {
		response = result.Data
	}

	if len(response) == 0 {
		// response, err = pnft.RetrieveGetNftListDeBank(config.DebankAPI, config.DebankToken, address)
		// if err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})

		// }
		response, err = pnft.RetrieveGetNftListQuickNode(config.QuickNodeAPI, address)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})

		}
		// save db:
		database.DBPNftInsertListNftCacheTable(&common.ListNftCache{
			Address: address,
			Data:    response,
		})
	}

	// var returnData interface{}
	// err = json.Unmarshal([]byte(response), &returnData)
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
	// 	return
	// }
	var jsonMap []map[string]interface{}
	json.Unmarshal([]byte(response), &jsonMap)

	c.JSON(200, gin.H{"Result": jsonMap})
}

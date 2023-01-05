package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/incognitochain/incognito-web-based-backend/database"
)

func APIGetNetworksFee(c *gin.Context) {
	data, err := database.DBRetrieveFeesTable(5)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err})
		return
	}
	c.JSON(200, gin.H{"Result": data})
}

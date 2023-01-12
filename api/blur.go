package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/incognitochain/incognito-web-based-backend/common"
	"github.com/incognitochain/incognito-web-based-backend/database"
)

func APIPBlurGetCollections(c *gin.Context) {

	// page := 1
	// var err error
	// if len(c.Query("page")) > 0 {
	// 	page, err = strconv.Atoi(c.Query("page"))
	// 	if err != nil {
	// 		c.JSON(http.StatusBadRequest, gin.H{"Error": "page invalid"})
	// 		return
	// 	}
	// }

	filter := c.Query("filters")

	fmt.Println("filter param: ", filter)

	var filterObj common.Filter
	if len(filter) > 0 {
		filter, err := url.QueryUnescape(filter)
		fmt.Println("filter: ", filter)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "filters invalid"})
			return
		}
		err = json.Unmarshal([]byte(filter), &filterObj)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "filters invalid: can not parse filter object"})
			return
		}
	}

	list, err := database.DBBlurGetCollectionList(&filterObj)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Limit": filterObj.Limit, "Page": filterObj.Page, "Offset": filterObj.Offset, "Query": filterObj.Query, "Total": len(list), "Result": list})
}

func APIPBlurGetCollectionDetail(c *gin.Context) {

	filter := c.Query("filters")

	fmt.Println("filter param: ", filter)

	var filterObj common.Filter
	if len(filter) > 0 {
		filter, err := url.QueryUnescape(filter)
		fmt.Println("filter: ", filter)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "filters invalid"})
			return
		}
		err = json.Unmarshal([]byte(filter), &filterObj)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "filters invalid: can not parse filter object"})
			return
		}
	}

	slug := c.Param("slug")

	fmt.Println("slug: ", slug)

	collection, _ := database.DBBlurGetCollectionDetail(slug)

	if collection == nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": "collection invalid"})
		return
	}

	list, err := database.DBBlurGetCollectionNFTs(collection.ContractAddress, &filterObj)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Result": list})
}

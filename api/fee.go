package api

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func APIEstimateSwap(c *gin.Context) {
	var req EstimateSwapRequest
	err := c.MustBindWith(&req, binding.JSON)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
}

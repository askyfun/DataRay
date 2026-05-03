package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// getPaginationParams extracts pagination parameters from query string
func getPaginationParams(c *gin.Context) (limit, offset int) {
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ = strconv.Atoi(limitStr)
	offset, _ = strconv.Atoi(offsetStr)

	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return
}

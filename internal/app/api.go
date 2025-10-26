package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /
func (a App) apiGetRootHandler(c *gin.Context) {
	requests, err := a.Store.LoadDatabase(c.Request.Context())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.IndentedJSON(http.StatusOK, requests)
}

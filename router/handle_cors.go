package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func handleCors(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("Access-Control-Max-Age", "86400")
	c.Status(http.StatusNoContent)
}

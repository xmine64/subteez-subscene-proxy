package router

import (
	"net/http"
	"subteez/subteez"

	"github.com/gin-gonic/gin"
)

type bannerRequest struct {
	ID string `uri:"id" binding:"required"`
}

func handleBanner(c *gin.Context) {
	var request bannerRequest
	if c.ShouldBindUri(&request) != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"status": "bad request",
			},
		)
		return
	}

	subteez.ProxyFile(`https://i.jeded.com/i/`+request.ID, c)
}

package router

import (
	"net/http"
	"subteez/subteez"

	"github.com/gin-gonic/gin"
)

func handleSearch(c *gin.Context) {
	var request subteez.SearchRequest
	err := c.ShouldBind(&request)
	if err != nil {
		c.Error(err)
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"status": "bad request",
			},
		)
		return
	}

	result, err := subteezApi.Search(request)

	if err != nil {
		c.Error(err)
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status": "server error",
			},
		)
		return
	}

	c.Header("Access-Control-Allow-Origin", "*")

	c.JSON(
		http.StatusOK,
		gin.H{
			"status": "ok",
			"result": result,
		},
	)
}

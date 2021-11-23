package router

import (
	"net/http"
	"subteez/subteez"

	"github.com/gin-gonic/gin"
)

var subteezApi subteez.SubteezApi

func InitializeAndRun(api subteez.SubteezApi, port string) {
	subteezApi = api

	router := gin.New()
	router.Use(gin.Logger())

	router.GET("/", func(c *gin.Context) {
		c.String(
			http.StatusOK,
			"Subteez server is running.",
		)
	})

	// Subteez API
	router.POST("/api/search", handleSearch)
	router.POST("/api/details", handleDetails)
	router.POST("/api/download", handleDownload)

	// direct download link
	router.GET("/subtitles/:movieName/:language/:file", handleDirectDownload)

	// banner download link
	router.GET("/i/:id", handleBanner)

	// cors support
	router.OPTIONS("/api/search", handleCors)
	router.OPTIONS("/api/details", handleCors)
	router.OPTIONS("/api/download", handleCors)

	// run router
	router.Run(":" + port)
}

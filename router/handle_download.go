package router

import (
	"net/http"
	"subteez/subteez"

	"github.com/gin-gonic/gin"
)

func handleDownload(c *gin.Context) {
	var request subteez.SubtitleDownloadRequest
	if c.ShouldBind(&request) != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"status": subteez.StatusBadRequest,
			},
		)
		return
	}

	url, err := subteezApi.GetDownloadLink(request)
	if err != nil {
		c.Error(err)

		if _, ok := err.(*subteez.NotFoundError); ok {
			c.JSON(
				http.StatusNotFound,
				gin.H{
					"status": subteez.StatusNotFound,
				},
			)
		} else {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"status": subteez.StatusServerError,
				},
			)
		}

		return
	}

	subteez.ProxyFile(url, c)
}

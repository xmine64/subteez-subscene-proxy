package router

import (
	"fmt"
	"net/http"
	"subteez/subscene"
	"subteez/subteez"

	"github.com/gin-gonic/gin"
)

type directDownloadRequest struct {
	MovieName string `uri:"movieName" binding:"required"`
	Language  string `uri:"language" binding:"required"`
	FileID    int    `uri:"file" binding:"required"`
}

func handleDirectDownload(c *gin.Context) {
	var request directDownloadRequest
	err := c.ShouldBindUri(&request)
	if err != nil {
		c.Error(err)
		c.String(
			http.StatusNotFound,
			http.StatusText(http.StatusNotFound),
		)
		return
	}

	fileUrl := fmt.Sprintf("/subtitles/%s/%s/%d",
		request.MovieName,
		request.Language,
		request.FileID,
	)

	url, err := subscene.GetDownloadLink(fileUrl)
	if err != nil {
		c.Error(err)
		code := http.StatusInternalServerError
		if err, ok := err.(*subscene.ResponseError); ok && err.StatusCode == http.StatusNotFound {
			code = http.StatusNotFound
		}
		c.String(code, http.StatusText(code))
		return
	}

	subteez.ProxyFile(url, c)
}

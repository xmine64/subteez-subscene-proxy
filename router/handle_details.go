package router

import (
	"net/http"
	"net/url"
	"subteez/subteez"

	"github.com/gin-gonic/gin"
)

// get banner proxied url, if there's a banner, or nil
func getBannerOrNil(c *gin.Context, posterUrl string) interface{} {
	if posterUrl == "" {
		return nil
	}
	parsedPosterUrl, err := url.Parse(posterUrl)
	if err != nil {
		return nil
	}
	var result url.URL
	if c.Request.TLS == nil {
		result.Scheme = "http"
	} else {
		result.Scheme = "https"
	}
	result.Host = c.Request.Host
	result.Path = parsedPosterUrl.Path
	return result.String()
}

func handleDetails(c *gin.Context) {
	var request subteez.SubtitleDetailsRequest
	err := c.ShouldBindJSON(&request)
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

	result, err := subteezApi.GetDetails(request)

	if err != nil {
		c.Error(err)

		if _, ok := err.(*subteez.NotFoundError); ok {
			c.JSON(
				http.StatusNotFound,
				gin.H{
					"status": "not found",
				},
			)
		} else {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"status": "server error",
				},
			)
		}
		return
	}

	c.Header("Access-Control-Allow-Origin", "*")

	result.Banner = getBannerOrNil(c, result.Banner.(string))

	c.JSON(
		http.StatusOK,
		result,
	)
}

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
	result := url.URL{
		Scheme: "http",
		Host:   c.Request.Host,
		Path:   parsedPosterUrl.Path,
	}
	forwardedProtocol := c.Request.Header.Get("X-Forwarded-Proto")
	if forwardedProtocol != "" {
		result.Scheme = forwardedProtocol
	} else {
		if c.Request.TLS != nil {
			result.Scheme = "https"
		}
	}
	return result.String()
}

func handleDetails(c *gin.Context) {
	var request subteez.SubtitleDetailsRequest
	err := c.ShouldBind(&request)
	if err != nil {
		c.Error(err)
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"status": subteez.StatusBadRequest,
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

	c.Header("Access-Control-Allow-Origin", "*")

	result.Banner = getBannerOrNil(c, result.Banner.(string))

	c.JSON(
		http.StatusOK,
		result,
	)
}

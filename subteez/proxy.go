package subteez

import (
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"

func ProxyFile(url string, c *gin.Context) {
	// create http client
	client := http.Client{
		Timeout: time.Second * 10,
	}
	// create http request
	httpRequest, err := http.NewRequest(
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		c.Error(err)
		c.String(
			http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError),
		)
		return
	}
	httpRequest.Header.Add("User-Agent", UserAgent)

	// send request
	response, err := client.Do(httpRequest)
	if response != nil {
		defer response.Body.Close()
	}
	if err != nil {
		c.Error(err)
		c.String(
			http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError),
		)
		return
	}

	// show error if response was not ok
	if response.StatusCode != http.StatusOK {
		c.String(
			response.StatusCode,
			http.StatusText(response.StatusCode),
		)
		return
	}

	// download file
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		c.String(
			http.StatusInternalServerError,
			"Error while downloading "+url,
		)
		return
	}

	// get content-type and content-disposition from response and forward them
	contentType := response.Header.Get("Content-Type")
	contentDisposition := response.Header.Get("Content-Disposition")
	if contentDisposition != "" {
		c.Header("Content-Disposition", contentDisposition)
	}

	c.Header("Access-Control-Allow-Origin", "*")

	c.Data(http.StatusOK, contentType, bytes)
}

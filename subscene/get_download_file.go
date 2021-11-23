package subscene

import (
	"errors"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func GetDownloadLink(href string) (string, error) {
	responseBody, err := sendRequest(http.MethodGet, href, nil, nil)
	if responseBody != nil {
		defer responseBody.Close()
	}
	if err != nil {
		return "", err
	}

	responseHtml, err := goquery.NewDocumentFromReader(responseBody)
	if err != nil {
		return "", err
	}

	href, exists := responseHtml.Find(`a[id="downloadButton"]`).Attr("href")
	if !exists {
		return "", errors.New(`"a[id="downloadButton] can not be found in the given address`)
	}

	result := baseURL + href

	return result, nil
}

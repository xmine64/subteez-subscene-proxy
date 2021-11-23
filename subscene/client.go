package subscene

import (
	"io"
	"net/http"
	"subteez/subteez"
	"time"
)

const baseURL = "https://subscene.com"

const urlencodedMimeType = "application/x-www-form-urlencoded"

type httpRequestBody struct {
	ContentType string
	ContentBody io.Reader
}

func sendRequest(
	method string,
	address string,
	body *httpRequestBody,
	languageFilter []string,
) (io.ReadCloser, error) {
	// initialize http client
	client := http.Client{
		Timeout: time.Second * 10,
	}

	// create http request
	var contentBody io.Reader = nil
	if body != nil {
		contentBody = body.ContentBody
	}
	httpRequest, err := http.NewRequest(method, baseURL+address, contentBody)
	if err != nil {
		return nil, err
	}

	// set request headers
	httpRequest.Header.Add("User-Agent", subteez.UserAgent)
	if body != nil {
		httpRequest.Header.Add("Content-Type", body.ContentType)
	}
	if languageFilter != nil {
		httpRequest.Header.Add("Cookie", "LanguageFilter="+getLanguagesFilterString(languageFilter))
	}

	// send request
	response, err := client.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return response.Body, &ResponseError{response.StatusCode}
	}

	return response.Body, nil
}

package subscene

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"subteez/subteez"

	"github.com/PuerkitoBio/goquery"
)

const searchPath = "/subtitles/searchbytitle"

func containsID(results []subteez.SearchResultItem, target string) bool {
	for _, result := range results {
		if result.ID == target {
			return true
		}
	}
	return false
}

func search(query string, languageFilter []string) ([]subteez.SearchResultItem, error) {
	// create url-encoded data to send query
	searchRequestBody := httpRequestBody{
		urlencodedMimeType,
		strings.NewReader(url.Values{"query": {query}}.Encode()),
	}

	// send request
	responseBody, err := sendRequest(
		http.MethodPost,
		searchPath,
		&searchRequestBody,
		languageFilter,
	)

	if responseBody != nil {
		defer responseBody.Close()
	}

	if err != nil {
		return nil, err
	}

	// parse response body
	responseHtml, err := goquery.NewDocumentFromReader(responseBody)
	if err != nil {
		return nil, err
	}
	titleNodes := responseHtml.Find(".title")
	countNodes := responseHtml.Find(".subtle.count")

	// get title of each search result
	titles := titleNodes.Map(func(i int, s *goquery.Selection) string {
		return s.Children().First().Text()
	})

	// get destination of each search result
	hrefs := titleNodes.Map(func(i int, s *goquery.Selection) string {
		return s.Children().First().AttrOr("href", "")
	})

	// get subtitles count of each search result
	counts := countNodes.Map(func(i int, s *goquery.Selection) string {
		trimmed := strings.TrimSpace(s.Text())
		countString := strings.SplitAfterN(trimmed, " ", 2)[0]
		return strings.TrimSpace(countString)
	})

	// create output
	results := []subteez.SearchResultItem{}

	for i, href := range hrefs {
		// ignore empty links
		if href == "" {
			continue
		}

		// ignore duplicate links
		if containsID(results, href) {
			continue
		}

		// convert count to integer
		count, err := strconv.Atoi(counts[i])
		if err != nil {
			return nil, err
		}

		// append result to output
		results = append(results, subteez.SearchResultItem{
			ID:    href,
			Name:  titles[i],
			Count: count,
		})
	}

	return results, nil
}

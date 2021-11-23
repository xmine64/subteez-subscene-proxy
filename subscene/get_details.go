package subscene

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"subteez/subteez"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

var languageDomMatch = map[string]string{
	"English":       "en",
	"Farsi/Persian": "fa",
	"Arabic":        "ar",
	"Hindi":         "hi",
	"German":        "de",
	"French":        "fr",
	"Italian":       "it",
	"Polish":        "pl",
	"Russian":       "ru",
	"Spanish":       "es",
	"Turkish":       "tr",
}

func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func contains(source []string, target string) bool {
	targetLower := strings.ToLower(target)
	for _, s := range source {
		sLower := strings.ToLower(s)
		if strings.Contains(targetLower, sLower) {
			return true
		}
	}
	return false
}

func normalizeFileName(fileName string, movieTitle string) string {
	nameParts := []string{}
	for _, part := range strings.FieldsFunc(movieTitle, func(r rune) bool {
		return unicode.IsSpace(r) || r == '-' || r == '_'
	}) {
		if len(part) >= 3 && !strings.EqualFold(part, "Season") {
			nameParts = append(nameParts, part)
		}
	}
	resultParts := []string{}
	for _, part := range strings.FieldsFunc(fileName, func(r rune) bool {
		return unicode.IsSpace(r) || r == '.' || r == '-' || r == '_'
	}) {
		if (len(part) > 2 || isNumeric(part)) && !contains(nameParts, part) {
			resultParts = append(resultParts, part)
		}
	}
	return strings.Join(resultParts, " ")
}

func getDetails(href string, languageFilters []string) (*subteez.SubtitleDetails, error) {
	responseBody, err := sendRequest(
		http.MethodGet,
		href,
		nil,
		languageFilters,
	)
	if responseBody != nil {
		defer responseBody.Close()
	}
	if err != nil {
		return nil, err
	}

	responseHtml, err := goquery.NewDocumentFromReader(responseBody)
	if err != nil {
		return nil, err
	}

	header := responseHtml.Find("div[class=\"header\"]")
	header.Find("h2").Children().Remove()
	title := strings.TrimSpace(header.Find("h2").Text())
	header.Find("li").Children().Remove()
	year := strings.TrimSpace(header.Find("li").Text())
	posterUrl := responseHtml.Find("div[class=\"poster\"]").Parent().AttrOr("href", "")

	ids := make([]string, 0)
	names := make([]string, 0)
	langs := make([]string, 0)

	responseHtml.Find("td[class=\"a1\"]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Find("a").Attr("href")
		if !exists {
			log.Default().Print("td[class=a1] => a: selected but no href, title=" + title)
			return
		}
		lang := strings.TrimSpace(s.Find("a").Children().Eq(0).Text())
		langCode, exists := languageDomMatch[lang]
		if !exists {
			log.Default().Printf("not supported language, title=%s, lang=%s", title, lang)
			return
		}

		name := strings.TrimSpace(s.Find("a").Children().Eq(1).Text())

		ids = append(ids, href)
		names = append(names, name)
		langs = append(langs, langCode)
	})

	authors := make([]string, len(names))

	responseHtml.Find("td[class=\"a5\"]").Each(func(i int, s *goquery.Selection) {
		author := strings.TrimSpace(s.Children().Eq(0).Text())
		authors[i] = author
	})

	comments := make([]string, len(names))

	responseHtml.Find("td[class=\"a6\"]").Each(func(i int, s *goquery.Selection) {
		comment := strings.TrimSpace(s.Children().Eq(0).Text())
		comments[i] = comment
	})

	files := make([]subteez.SubtitleFile, len(names))

	for i, name := range names {
		files[i] = subteez.SubtitleFile{
			ID:       ids[i],
			Language: langs[i],
			Name:     name,
			Author:   authors[i],
			Comment:  comments[i],
			Title:    normalizeFileName(name, title),
		}
	}

	return &subteez.SubtitleDetails{
		Status: subteez.StatusOk,
		Name:   title,
		Year:   year,
		Banner: posterUrl,
		Files:  files,
	}, nil

}

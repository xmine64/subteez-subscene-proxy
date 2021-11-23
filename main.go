package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		//log.Fatal("$PORT must be set")
		port = "5000"
	}

	router := gin.New()
	router.Use(gin.Logger())

	router.Static("/static/", "./static/root/")
	router.Static("/web/", "./static/web/")

	router.StaticFile("/favicon.ico", "./static/resources/favicon.ico")
	//router.GET("/", func(c *gin.Context) {
	//	c.Writer.WriteString("Subteez Server is running. Redirecting to home page...")
	//})

	router.StaticFile("/", "./static/root/index.html")
	router.StaticFile("/app-ads.txt", "./static/root/app-ads.txt")

	router.POST("/api/search", handleSearch)
	router.POST("/api/details", handleDetails)
	router.POST("/api/download", handleDownload)

	router.GET("/subtitles/:movieName/:language/:file", handleDirectDownload)

	router.GET("/i/:id", handleBanner)

	router.OPTIONS("/api/search", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		c.Header("Access-Control-Max-Age", "86400")
		c.Status(http.StatusNoContent)
	})

	router.OPTIONS("/api/details", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Method", "POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		c.Header("Access-Control-Max-Age", "86400")
		c.Status(http.StatusNoContent)
	})

	router.OPTIONS("/api/download", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Method", "POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		c.Header("Access-Control-Max-Age", "86400")
		c.Status(http.StatusNoContent)
	})

	router.OPTIONS("/i/:id", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Method", "GET, OPTIONS")
		c.Header("Access-Control-Max-Age", "86400")
		c.Status(http.StatusNoContent)
	})

	router.OPTIONS("/subtitles/:movieName/:language/:file", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Method", "GET, OPTIONS")
		c.Header("Access-Control-Max-Age", "86400")
		c.Status(http.StatusNoContent)
	})

	router.Run(":" + port)
}

type SearchRequest struct {
	Query    string   `form:"query" json:"query" binding:"required"`
	Language []string `form:"lang" json:"lang" binding:"required"`
}

type DetailsRequest struct {
	Id       string   `form:"id" json:"id" binding:"required"`
	Language []string `form:"lang" json:"lang" binding:"required"`
}

type DownloadRequest struct {
	Id string `form:"id" json:"id" uri:"id" binding:"required"`
}

type BannerRequest struct {
	Id string `uri:"id" binding:"required"`
}

type DirectDownloadRequest struct {
	MovieName string `uri:"movieName" binding:"required"`
	Language  string `uri:"language" binding:"required"`
	FileId    int    `uri:"file" binding:"required"`
}

const BASE_URL = "https://subscene.com"
const USER_AGENT = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
const SEARCH_PATH = "/subtitles/searchbytitle"

var filterMap = map[string]int{
	"en": 13,
	"fa": 46,
	"ar": 2,
	"hi": 51,
	"de": 19,
	"fr": 18,
	"it": 26,
	"pl": 31,
	"ru": 34,
	"es": 38,
	"tr": 41,
}

var domMatch = map[string]string{
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

func getLanguageFilterString(filter []string) string {
	resultArray := make([]string, len(filter))
	for i, v := range filter {
		resultArray[i] = strconv.Itoa(filterMap[v])
	}
	return strings.Join(resultArray, ",")
}

func containsResult(results []map[string]interface{}, targetId string) bool {
	for _, result := range results {
		if result["id"] == targetId {
			return true
		}
	}
	return false
}

func handleSearch(c *gin.Context) {
	var request SearchRequest
	err := c.BindJSON(&request)
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
	client := http.Client{
		Timeout: time.Second * 10,
	}
	httpRequestParams := url.Values{"query": {request.Query}}
	httpRequest, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprint(BASE_URL, SEARCH_PATH),
		strings.NewReader(httpRequestParams.Encode()),
	)
	if err != nil {
		c.Error(err)
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status": "server error",
			},
		)
		return
	}
	httpRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	httpRequest.Header.Add("User-Agent", USER_AGENT)
	httpRequest.Header.Add("Cookie", fmt.Sprint("LanguageFilter=", getLanguageFilterString(request.Language)))

	response, err := client.Do(httpRequest)
	if err != nil {
		c.Error(err)
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status": "server error",
			},
		)
		return
	}
	defer response.Body.Close()
	html, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		c.Error(err)
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status": "server error",
			},
		)
		return
	}

	titles := []string{}
	ids := []string{}

	html.Find("div[class=title]").Each(func(i int, s *goquery.Selection) {
		a := s.Children().First()
		href, exists := a.Attr("href")
		if !exists {
			return
		}
		titles = append(titles, a.Text())
		ids = append(ids, href)
	})

	counts := make([]int, len(titles))

	html.Find("[class=\"subtle count\"]").Each(func(i int, s *goquery.Selection) {
		contentTrimmed := strings.TrimSpace(s.Text())
		countStr := strings.SplitAfter(contentTrimmed, " ")[0]
		countStrTrimmed := strings.TrimSpace(countStr)
		count, err := strconv.Atoi(countStrTrimmed)
		if err != nil {
			c.Error(err)
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"status": "server error",
				},
			)
		}
		counts[i] = count
	})

	result := []map[string]interface{}{}

	for i := 0; i < len(titles); i++ {
		if containsResult(result, ids[i]) {
			continue
		}
		result = append(result, map[string]interface{}{
			"name":  titles[i],
			"count": counts[i],
			"id":    ids[i],
		})
	}

	c.Header("Access-Control-Allow-Origin", "*")

	c.JSON(
		http.StatusOK,
		gin.H{
			"status": "ok",
			"result": result,
		},
	)
}

const EMPTY_POSTER = "No Image"

func posterUrlOrNil(c *gin.Context, posterUrl string) interface{} {
	if posterUrl == EMPTY_POSTER {
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

func handleDetails(c *gin.Context) {
	var request DetailsRequest
	err := c.BindJSON(&request)
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
	client := http.Client{
		Timeout: time.Second * 10,
	}
	httpRequest, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprint(BASE_URL, request.Id),
		nil,
	)
	if err != nil {
		c.Error(err)
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status": "server error",
			},
		)
		return
	}
	httpRequest.Header.Add("User-Agent", USER_AGENT)
	httpRequest.Header.Add("Cookie", fmt.Sprint("LanguageFilter=", getLanguageFilterString(request.Language)))
	response, err := client.Do(httpRequest)
	if err != nil {
		c.Error(err)
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status": "server error",
			},
		)
		return
	}
	defer response.Body.Close()
	html, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		c.Error(err)
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status": "server error",
			},
		)
		return
	}
	header := html.Find("div[class=\"header\"]")
	header.Find("h2").Children().Remove()
	title := strings.TrimSpace(header.Find("h2").Text())
	header.Find("li").Children().Remove()
	year := strings.TrimSpace(header.Find("li").Text())
	posterUrl := html.Find("div[class=\"poster\"]").Parent().AttrOr("href", EMPTY_POSTER)

	ids := make([]string, 0)
	names := make([]string, 0)
	langs := make([]string, 0)

	html.Find("td[class=\"a1\"]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Find("a").Attr("href")
		if !exists {
			c.Error(errors.New(fmt.Sprint("td[class=a1] => a: selected but no href, title=", title)))
			return
		}
		lang := strings.TrimSpace(s.Find("a").Children().Eq(0).Text())
		langCode, exists := domMatch[lang]
		if !exists {
			c.Error(fmt.Errorf("not supported language, title=%s, lang=%s", title, lang))
			return
		}

		name := strings.TrimSpace(s.Find("a").Children().Eq(1).Text())

		ids = append(ids, href)
		names = append(names, name)
		langs = append(langs, langCode)
	})

	authors := make([]string, len(names))

	html.Find("td[class=\"a5\"]").Each(func(i int, s *goquery.Selection) {
		author := strings.TrimSpace(s.Children().Eq(0).Text())
		authors[i] = author
	})

	comments := make([]string, len(names))

	html.Find("td[class=\"a6\"]").Each(func(i int, s *goquery.Selection) {
		comment := strings.TrimSpace(s.Children().Eq(0).Text())
		comments[i] = comment
	})

	files := make([]map[string]interface{}, len(names))

	for i := 0; i < len(names); i++ {
		files[i] = map[string]interface{}{
			"id":      ids[i],
			"lang":    langs[i],
			"name":    names[i],
			"author":  authors[i],
			"comment": comments[i],
			"title":   normalizeFileName(names[i], title),
		}
	}

	c.Header("Access-Control-Allow-Origin", "*")

	c.JSON(
		http.StatusOK,
		gin.H{
			"status":    "ok",
			"name":      title,
			"year":      year,
			"posterUrl": posterUrlOrNil(c, posterUrl),
			"files":     files,
		},
	)
}

type errorPageNotFound struct{}
type errorUnexpectedResponse struct{}

func (e *errorPageNotFound) Error() string {
	return "404 Page not found."
}

func (e *errorUnexpectedResponse) Error() string {
	return "Unexpected response from Subscene."
}

func downloadAndHostFile(address string, c *gin.Context) error {
	client := http.Client{
		Timeout: time.Second * 10,
	}
	httpRequest, err := http.NewRequest(
		http.MethodGet,
		address,
		nil,
	)
	if err != nil {
		return err
	}
	httpRequest.Header.Add("User-Agent", USER_AGENT)
	response, err := client.Do(httpRequest)
	if err != nil {
		return err
	}

	if response.StatusCode == http.StatusNotFound {
		return &errorPageNotFound{}
	}

	if response.StatusCode != http.StatusOK {
		return &errorUnexpectedResponse{}
	}

	contentType := response.Header.Get("Content-Type")

	bytes, err := io.ReadAll(response.Body)

	if err != nil {
		return err
	}

	contentDisposition := response.Header.Get("Content-Disposition")

	if contentDisposition != "" {
		c.Header("Content-Disposition", contentDisposition)
	}

	c.Header("Access-Control-Allow-Origin", "*")

	c.Data(http.StatusOK, contentType, bytes)

	return nil
}

func downloadSubtitle(address string, c *gin.Context) error {
	client := http.Client{
		Timeout: time.Second * 10,
	}
	httpRequest, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprint(BASE_URL, address),
		nil,
	)
	if err != nil {
		return err
	}
	httpRequest.Header.Add("User-Agent", USER_AGENT)
	response, err := client.Do(httpRequest)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	html, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return err
	}
	href, exists := html.Find("a[id=\"downloadButton\"]").Attr("href")
	if !exists {
		return &errorPageNotFound{}
	}

	return downloadAndHostFile(fmt.Sprint(BASE_URL, href), c)
}

func handleDownload(c *gin.Context) {
	var request DownloadRequest
	if c.Bind(&request) != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"status": "bad request",
			},
		)
		return
	}

	err := downloadSubtitle(request.Id, c)

	if err == nil {
		return
	}

	switch err.(type) {
	default:
		c.Error(err)
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status": "server error",
			},
		)
		return
	case *errorPageNotFound:
		c.JSON(
			http.StatusNotFound,
			gin.H{
				"status": "not found",
			},
		)
		return
	}
}

func handleBanner(c *gin.Context) {
	var request BannerRequest
	if c.ShouldBindUri(&request) != nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"status": "bad request",
			},
		)
		return
	}

	fileUrl := fmt.Sprintf("https://i.jeded.com/i/%s", request.Id)

	err := downloadAndHostFile(fileUrl, c)

	if err == nil {
		return
	}

	switch err.(type) {
	default:
		c.String(
			http.StatusInternalServerError,
			"500 Internal Server Error",
			nil,
		)
		return
	case *errorPageNotFound:
		c.String(
			http.StatusNotFound,
			"404 Page not found.",
			nil,
		)
		return
	}
}

func handleDirectDownload(c *gin.Context) {
	var request DirectDownloadRequest
	err := c.ShouldBindUri(&request)
	if err != nil {
		c.Error(err)
		c.String(
			http.StatusNotFound,
			"404 Page not found.",
		)
		return
	}

	fileUrl := fmt.Sprintf("/subtitles/%s/%s/%d",
		request.MovieName,
		request.Language,
		request.FileId,
	)

	err = downloadSubtitle(fileUrl, c)

	if err == nil {
		return
	}

	switch err.(type) {
	default:
		c.Error(err)
		c.String(
			http.StatusInternalServerError,
			"500 Internal Server Error",
			nil,
		)
		return
	case *errorPageNotFound:
		c.String(
			http.StatusNotFound,
			"404 Page not found.",
			nil,
		)
		return
	}
}

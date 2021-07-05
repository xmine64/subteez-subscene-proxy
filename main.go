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

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Subteez Server is running.")
	})

	router.POST("/api/search", handleSearch)
	router.POST("/api/details", handleDetails)
	router.POST("/api/download", handleDownload)

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
	Id string `form:"id" json:"id" binding:"required"`
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

	titles := make([]string, 0)
	ids := make([]string, 0)

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

	result := make([]map[string]interface{}, len(titles))

	for i := 0; i < len(titles); i++ {
		result[i] = map[string]interface{}{
			"name":  titles[i],
			"count": counts[i],
			"id":    ids[i],
		}
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"status": "ok",
			"result": result,
		},
	)
}

const EMPTY_POSTER = "No Image"

func posterUrlOrNil(posterUrl string) interface{} {
	if posterUrl == EMPTY_POSTER {
		return nil
	}
	return posterUrl
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
		}
	}

	c.JSON(
		http.StatusOK,
		gin.H{
			"status":    "ok",
			"name":      title,
			"year":      year,
			"posterUrl": posterUrlOrNil(posterUrl),
			"files":     files,
		},
	)
}

func handleDownload(c *gin.Context) {
	var request DownloadRequest
	if c.BindJSON(&request) != nil {
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
	href, exists := html.Find("a[id=\"downloadButton\"]").Attr("href")
	if !exists {
		c.Error(fmt.Errorf("can not download file: %s", request.Id))
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status": "server error",
			},
		)
	}
	httpRequest, err = http.NewRequest(
		http.MethodGet,
		fmt.Sprint(BASE_URL, href),
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
	response, err = client.Do(httpRequest)
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

	contentType := response.Header.Get("Content-Type")

	if !strings.Contains(contentType, "application/x-zip-compressed") {
		c.Error(fmt.Errorf("unexpected response: %s", contentType))
		c.JSON(
			http.StatusInternalServerError,
			gin.H{
				"status": "server error",
			},
		)
		return
	}

	bytes, err := io.ReadAll(response.Body)

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

	c.Data(http.StatusOK, contentType, bytes)
}

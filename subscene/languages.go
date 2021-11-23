package subscene

import (
	"strconv"
	"strings"
)

var languageCodes = map[string]int{
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

func getLanguagesFilterString(filterMap []string) string {
	resultArray := make([]string, len(filterMap))
	for i, v := range filterMap {
		resultArray[i] = strconv.Itoa(languageCodes[v])
	}
	return strings.Join(resultArray, ",")
}

package subscene

import (
	"net/http"
	"subteez/subteez"
)

type SubsceneApi struct{}

func (SubsceneApi) Search(request subteez.SearchRequest) ([]subteez.SearchResult, error) {
	return search(request.Query, request.Language)
}

func (SubsceneApi) GetDetails(request subteez.SubtitleDetailsRequest) (*subteez.SubtitleDetails, error) {
	result, err := getDetails(request.ID, request.Language)
	if err, ok := err.(*ResponseError); ok && err.StatusCode == http.StatusNotFound {
		return nil, &subteez.NotFoundError{}
	}
	return result, err
}

func (SubsceneApi) GetDownloadLink(request subteez.SubtitleDownloadRequest) (string, error) {
	result, err := GetDownloadLink(request.ID)
	if err, ok := err.(*ResponseError); ok && err.StatusCode == http.StatusNotFound {
		return "", &subteez.NotFoundError{}
	}
	return result, err
}

package subteez

type SubteezApi interface {
	Search(SearchRequest) (*SearchResult, error)
	GetDetails(SubtitleDetailsRequest) (*SubtitleDetails, error)
	GetDownloadLink(SubtitleDownloadRequest) (string, error)
}

package data

import (
	"errors"
	"serve/core"

	"github.com/blevesearch/bleve/v2"
)

type SearchResult struct {
	Url   string
	Score int
}

type SearchPlugin struct {
	index bleve.Index
}

func NewSearchPlugin() *SearchPlugin {
	return &SearchPlugin{}
}

// Implement Plugin interface
func (SearchPlugin) Name() string        { return "builtin/search" }
func (SearchPlugin) Version() string     { return "0.1" }
func (SearchPlugin) Description() string { return "Builtin search" }

func (sp *SearchPlugin) Initialize(params map[string]string) error {
	// Create new index if it doesn't exist
	mapping := bleve.NewIndexMapping()
	var err error
	// TODO create a configuration option for a persistent index
	// and try to open it if it exists (index, err := bleve.Open(indexPath)...)
	index, err := bleve.NewMemOnly(mapping) // use New("index_name") for persistent storage
	if err == nil {
		*sp = SearchPlugin{index: index}
	}
	return err
}

func (sp *SearchPlugin) Shutdown() {
	sp.index.Close()
	*sp = SearchPlugin{index: nil}
}

func (sp *SearchPlugin) AddFile(file *core.File) error {
	if file == nil {
		return errors.New("file cannot be nil")
	}
	if file.CachedContent == nil {
		return errors.New("file content cannot be nil")
	}
	return sp.index.Index(file.Url, file.CachedContent)
}

func (sp *SearchPlugin) Search(query string, limit int) ([]SearchResult, error) {
	searchRequest := bleve.NewSearchRequest(bleve.NewQueryStringQuery(query))
	searchRequest.Size = limit
	searchRequest.Highlight = bleve.NewHighlight()

	searchResults, err := sp.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, len(searchResults.Hits))
	for i, hit := range searchResults.Hits {
		results[i] = SearchResult{
			Url:   hit.ID,
			Score: int(hit.Score * 1000),
		}
	}

	return results, nil

}

func (sp *SearchPlugin) GetData(offset int, length int) ([]string, error) {
	//		if offset < 0 || length < 0 || offset+length > len(sp.data) {
	//			return nil, errors.New("invalid offset or length")
	//	}
	//
	//	return sp.data[offset : offset+length], nil
	return nil, errors.New("GetData not implemented yet")
}

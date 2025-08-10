package plugins

import (
	"path/filepath"
	"serve/core"
	"strings"
	"sync"

	"github.com/blevesearch/bleve/v2"
)

type SearchResult struct {
	Url   string
	Score int
}

type BuiltinSearchPlugin struct {
	index bleve.Index
	mu    sync.RWMutex
}

func NewSearchPlugin(params map[string]string) *BuiltinSearchPlugin {
	mapping := bleve.NewIndexMapping()
	var err error
	// TODO create a configuration option for a persistent index
	// and try to open it if it exists (index, err := bleve.Open(indexPath)...)
	index, err := bleve.NewMemOnly(mapping) // use New("index_name") for persistent storage
	if err == nil {
		return &BuiltinSearchPlugin{index: index}
	}
	return nil
}

func (p *BuiltinSearchPlugin) Name() string {
	return "builtin/search"
}

func (p *BuiltinSearchPlugin) Priority() int {
	return 1000 // Run last
}

// TODO use contenttype from file's metadata
func (p *BuiltinSearchPlugin) CanProcess(file *core.File) bool {
	// Index text-based files
	ext := strings.ToLower(filepath.Ext(file.Name))
	return ext == ".txt" || ext == ".md" || ext == ".markdown" || ext == ".html" || ext == ".htm"
}

func (p *BuiltinSearchPlugin) Process(ctx *core.PluginContext) *core.PluginResult {
	if ctx.File.Content == nil {
		return &core.PluginResult{
			Success: false,
		}
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// TODO remove /content from "Path"
	err := p.index.Index(ctx.File.Path, ctx.File.Content)
	if err != nil {
		return &core.PluginResult{
			Success: false,
		}
	}

	return &core.PluginResult{
		Success: true,
	}
}

// GetSearchResults searches the index for a term
func (p *BuiltinSearchPlugin) GetSearchResults(query string, limit int) ([]SearchResult, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	searchRequest := bleve.NewSearchRequest(bleve.NewQueryStringQuery(query))
	searchRequest.Size = limit
	searchRequest.Highlight = bleve.NewHighlight()

	searchResults, err := p.index.Search(searchRequest)
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

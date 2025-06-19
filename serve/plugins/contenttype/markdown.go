package contenttype

import (
	"bytes"
	"serve/core"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

type ContentTypeMarkdownPlugin struct {
	core.ContentTypePlugin
}

func NewMarkdownPlugin() core.ContentTypePlugin {
	return ContentTypeMarkdownPlugin{}
}

func (ContentTypeMarkdownPlugin) Name() string         { return "builtin/markdown" }
func (ContentTypeMarkdownPlugin) Version() string      { return "0.1" }
func (ContentTypeMarkdownPlugin) Mimetype() string     { return "text/html" }
func (ContentTypeMarkdownPlugin) IgnoreLayout() bool   { return false }
func (ContentTypeMarkdownPlugin) Id() string           { return "2E9C1AB9-2D58-4BB5-989F-8269C6D2007A" }
func (ContentTypeMarkdownPlugin) Description() string  { return "Renders markdown files" }
func (ContentTypeMarkdownPlugin) Extensions() []string { return []string{"md", "markdown"} }

func (ContentTypeMarkdownPlugin) Convert(raw string) (string, error) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"), // or any Chroma style
				highlighting.WithFormatOptions(
					chromahtml.WithLineNumbers(true), // optional line numbers
				),
			),
		),
	)

	var buf bytes.Buffer
	if err := markdown.Convert([]byte(raw), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

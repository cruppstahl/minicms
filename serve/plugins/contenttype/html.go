package contenttype

import (
	"serve/core"
)

type ContentTypeHtmlPlugin struct {
	core.ContentTypePlugin
}

func NewHtmlPlugin() core.ContentTypePlugin {
	return ContentTypeHtmlPlugin{}
}

func (ContentTypeHtmlPlugin) Name() string         { return "builtin/html" }
func (ContentTypeHtmlPlugin) Version() string      { return "0.1" }
func (ContentTypeHtmlPlugin) Mimetype() string     { return "text/html" }
func (ContentTypeHtmlPlugin) IgnoreLayout() bool   { return false }
func (ContentTypeHtmlPlugin) Id() string           { return "8459719D-0E43-42B9-B479-91041E81CDFC" }
func (ContentTypeHtmlPlugin) Description() string  { return "Renders html files" }
func (ContentTypeHtmlPlugin) Extensions() []string { return []string{"html", "htm"} }

func (ContentTypeHtmlPlugin) Convert(context *core.Context, file *core.File) error {
	// no need to do anything to render a html file
	return nil
}

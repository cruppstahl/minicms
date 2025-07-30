package contenttype

import (
	"serve/core"
)

type ContentTypeHtmlPlugin struct {
	core.ContentTypePlugin
}

func NewHtmlPlugin() core.ContentTypePlugin {
	return &ContentTypeHtmlPlugin{}
}

func (*ContentTypeHtmlPlugin) Name() string                              { return "builtin/html" }
func (*ContentTypeHtmlPlugin) Version() string                           { return "0.1" }
func (*ContentTypeHtmlPlugin) Mimetype() string                          { return "text/html" }
func (*ContentTypeHtmlPlugin) IgnoreLayout() bool                        { return false }
func (*ContentTypeHtmlPlugin) Description() string                       { return "Renders html files" }
func (*ContentTypeHtmlPlugin) Extensions() []string                      { return []string{"html", "htm"} }
func (*ContentTypeHtmlPlugin) Initialize(params map[string]string) error { return nil }
func (*ContentTypeHtmlPlugin) Shutdown()                                 {}

func (*ContentTypeHtmlPlugin) Render(raw string) (string, error) {
	// no need to do anything to render a html file
	return raw, nil
}

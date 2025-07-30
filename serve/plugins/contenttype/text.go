package contenttype

import (
	"serve/core"
)

type ContentTypeTextPlugin struct {
	core.ContentTypePlugin
}

func NewTextPlugin() core.ContentTypePlugin {
	return &ContentTypeTextPlugin{}
}

func (*ContentTypeTextPlugin) Name() string                              { return "builtin/text" }
func (*ContentTypeTextPlugin) Version() string                           { return "0.1" }
func (*ContentTypeTextPlugin) Mimetype() string                          { return "text/plain" }
func (*ContentTypeTextPlugin) IgnoreLayout() bool                        { return true }
func (*ContentTypeTextPlugin) Description() string                       { return "Renders text files" }
func (*ContentTypeTextPlugin) Extensions() []string                      { return []string{"text", "txt"} }
func (*ContentTypeTextPlugin) Initialize(params map[string]string) error { return nil }
func (*ContentTypeTextPlugin) Shutdown()                                 {}

func (*ContentTypeTextPlugin) Render(raw string) (string, error) {
	// no need to do anything to render a text file
	return raw, nil
}

package contenttype

import (
	"serve/core"
)

type ContentTypeTextPlugin struct {
	core.ContentTypePlugin
}

func NewTextPlugin() core.ContentTypePlugin {
	return ContentTypeTextPlugin{}
}

func (ContentTypeTextPlugin) Name() string         { return "builtin/text" }
func (ContentTypeTextPlugin) Version() string      { return "0.1" }
func (ContentTypeTextPlugin) Mimetype() string     { return "text/plain" }
func (ContentTypeTextPlugin) IgnoreLayout() bool   { return true }
func (ContentTypeTextPlugin) Id() string           { return "61F80BE4-5805-4C65-8666-62CB972D38EB" }
func (ContentTypeTextPlugin) Description() string  { return "Renders text files" }
func (ContentTypeTextPlugin) Extensions() []string { return []string{"text", "txt"} }

func (ContentTypeTextPlugin) Convert(context *core.Context, file *core.File) error {
	// no need to do anything to render a text file
	return nil
}
